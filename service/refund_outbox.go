package service

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/model"
)

const (
	defaultRefundMaxAttempts = 8
	refundWorkerInterval     = 5 * time.Second
	refundClaimBatch         = 20
)

func refundOutboxEnabled() bool {
	v := strings.TrimSpace(os.Getenv("REFUND_OUTBOX_ENABLED"))
	if v == "" {
		return true
	}
	return v == "true" || v == "1" || strings.EqualFold(v, "yes")
}

func refundMaxAttempts() int {
	n := common.GetEnvOrDefault("REFUND_OUTBOX_MAX_ATTEMPTS", defaultRefundMaxAttempts)
	if n <= 0 {
		return defaultRefundMaxAttempts
	}
	return n
}

// EnqueueRefundIntent persists a refund intent and processes it asynchronously.
func EnqueueRefundIntent(intent *model.RefundIntent) (*model.RefundIntent, bool, error) {
	row, created, err := model.CreateRefundIntentIfAbsent(intent)
	if err != nil {
		return nil, false, err
	}
	RecordRefundIntentMetric(row.Status)
	go processRefundIntent(row.Id)
	return row, created, nil
}

func processRefundIntent(id int) {
	if id <= 0 {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			common.SysLog(fmt.Sprintf("panic process refund intent id=%d: %v", id, r))
		}
	}()

	var intent model.RefundIntent
	if err := model.DB.Where("id = ?", id).First(&intent).Error; err != nil {
		common.SysLog(fmt.Sprintf("load refund intent %d: %v", id, err))
		return
	}
	if intent.Status == model.RefundIntentSucceeded || intent.Status == model.RefundIntentDead {
		return
	}

	// Ensure processing status (worker claim may have set it).
	if intent.Status != model.RefundIntentProcessing {
		_ = model.DB.Model(&model.RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     model.RefundIntentProcessing,
			"updated_at": time.Now().Unix(),
		}).Error
	}

	fail := func(msg string) {
		_ = model.MarkRefundIntentFailed(id, intent.Attempts, refundMaxAttempts(), msg)
		RecordRefundIntentMetric(model.RefundIntentFailed)
	}

	// 1) Wallet (non-idempotent — only once)
	if intent.WalletConsumed > 0 && !intent.WalletDone {
		if err := model.IncreaseUserQuota(intent.UserId, intent.WalletConsumed, false); err != nil {
			fail("wallet: " + err.Error())
			return
		}
		_ = model.DB.Model(&model.RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
			"wallet_done": true,
			"updated_at":  time.Now().Unix(),
		}).Error
		intent.WalletDone = true
	}

	// 2) Subscription pre-consume (idempotent by request id)
	if !intent.SubscriptionDone {
		if intent.FundingSource == BillingSourceSubscription && intent.FundingRequestId != "" {
			if err := model.RefundSubscriptionPreConsume(intent.FundingRequestId); err != nil {
				fail("subscription: " + err.Error())
				return
			}
		}
		if intent.ExtraReserved > 0 && intent.SubscriptionId > 0 {
			if err := model.PostConsumeUserSubscriptionDelta(intent.SubscriptionId, -int64(intent.ExtraReserved)); err != nil {
				fail("sub_extra: " + err.Error())
				return
			}
		}
		_ = model.DB.Model(&model.RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
			"subscription_done": true,
			"updated_at":        time.Now().Unix(),
		}).Error
		intent.SubscriptionDone = true
	}

	// 3) Token quota
	if intent.TokenQuota > 0 && !intent.IsPlayground && !intent.TokenDone {
		tokenKey := intent.TokenKey
		if tokenKey == "" {
			var tok model.Token
			if err := model.DB.Select("id", "key").Where("id = ?", intent.TokenId).First(&tok).Error; err == nil {
				tokenKey = tok.Key
			}
		}
		if err := model.IncreaseTokenQuota(intent.TokenId, tokenKey, intent.TokenQuota); err != nil {
			fail("token: " + err.Error())
			return
		}
		_ = model.DB.Model(&model.RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
			"token_done": true,
			"updated_at": time.Now().Unix(),
		}).Error
	}

	if err := model.MarkRefundIntentSucceeded(id); err != nil {
		common.SysLog("mark refund succeeded error: " + err.Error())
		return
	}
	RecordRefundIntentMetric(model.RefundIntentSucceeded)
}

// RunRefundOutboxOnce claims and processes a batch of refund intents.
func RunRefundOutboxOnce() {
	if !refundOutboxEnabled() {
		return
	}
	claimed, err := model.ClaimRefundIntents(refundClaimBatch)
	if err != nil {
		common.SysLog("claim refund intents: " + err.Error())
		return
	}
	for _, intent := range claimed {
		processRefundIntent(intent.Id)
	}
}

// StartRefundOutboxWorker starts a background loop until stop is closed.
func StartRefundOutboxWorker(stop <-chan struct{}) {
	if !refundOutboxEnabled() {
		common.SysLog("refund outbox worker disabled via REFUND_OUTBOX_ENABLED")
		return
	}
	common.SysLog("refund outbox worker started interval=" + strconv.Itoa(int(refundWorkerInterval.Seconds())) + "s")
	ticker := time.NewTicker(refundWorkerInterval)
	go func() {
		defer ticker.Stop()
		// process any leftovers immediately on boot
		RunRefundOutboxOnce()
		for {
			select {
			case <-stop:
				common.SysLog("refund outbox worker stopped")
				return
			case <-ticker.C:
				RunRefundOutboxOnce()
			}
		}
	}()
}

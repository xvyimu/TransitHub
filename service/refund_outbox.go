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

// refundOutboxEnabled 是否启用持久化退款 outbox。
// 约束：环境变量 REFUND_OUTBOX_ENABLED；空/缺省为 true；false 时 BillingSession 回退 gopool 内联退款。
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

// EnqueueRefundIntent 持久化退款意图并异步处理（至少一次语义的入口）。
//
// 行为：
//   - CreateRefundIntentIfAbsent（按 idempotency_key 去重）
//   - 记录进程内 metric 后 go processRefundIntent
//
// 约束：
//   - 调用方须已构造完整 intent（含幂等键、额度分项）；禁止传 nil
//   - 入队成功不代表已退完；以 status=succeeded / dead 为准
//   - TokenKey 仅存库用于回写额度，列表 API 不得回传
//
// 参见：Billing_session.Refund、model/refund_intent.go、StartRefundOutboxWorker
func EnqueueRefundIntent(intent *model.RefundIntent) (*model.RefundIntent, bool, error) {
	row, created, err := model.CreateRefundIntentIfAbsent(intent)
	if err != nil {
		return nil, false, err
	}
	RecordRefundIntentMetric(row.Status)
	go processRefundIntent(row.Id)
	return row, created, nil
}

// processRefundIntent 执行单条退款意图：钱包 → 订阅 → 令牌，分步打 done 标记。
//
// 约束：
//   - 已 succeeded/dead 直接返回
//   - 钱包/令牌回写非天然幂等：依赖 wallet_done/token_done；进程若在额度增加成功与 done 落库之间崩溃，可能双倍回补（已知窗口，见审查报告）
//   - 订阅预扣按 funding_request_id 走幂等 RefundSubscriptionPreConsume
//   - 失败走 MarkRefundIntentFailed；attempts 达上限 → dead
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

// RunRefundOutboxOnce 认领一批 pending/failed 意图并处理（worker 单次滴答）。
// 约束：REFUND_OUTBOX_ENABLED=false 时 no-op；认领用 CAS 式 status 条件更新防双工。
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

// StartRefundOutboxWorker 启动退款 outbox 后台循环，直到 stop 关闭。
//
// 行为：启动时立即 RunRefundOutboxOnce，之后每 5s 一轮。
// 约束：由 main 在系统任务上下文注入 stop；禁用开关时只打日志不启动。
// 参见：main.go systemTaskCtx
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

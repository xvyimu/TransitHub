package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/xvyimu/TransitHub/common"

	"gorm.io/gorm"
)

// Refund 意图状态常量（outbox 状态机）。
const (
	RefundIntentPending    = "pending"
	RefundIntentProcessing = "processing"
	RefundIntentSucceeded  = "succeeded"
	RefundIntentFailed     = "failed"
	RefundIntentDead       = "dead"
)

// RefundIntent 持久化退款意图，保证进程中断后仍可至少完成一次（WP-C）。
//
// 约束：
//   - IdempotencyKey 全局唯一；重复入队返回已有行
//   - TokenKey 仅服务端回写额度用，JSON 不导出；List API 必须清空
//   - wallet_done / subscription_done / token_done 标记分步进度，防重复回补
type RefundIntent struct {
	Id               int    `json:"id" gorm:"primaryKey;autoIncrement"`
	IdempotencyKey   string `json:"idempotency_key" gorm:"type:varchar(128);uniqueIndex;not null"`
	TokenId          int    `json:"token_id" gorm:"index;not null"`
	UserId           int    `json:"user_id" gorm:"index"`
	TokenQuota       int    `json:"token_quota" gorm:"not null;default:0"`
	ExtraReserved    int    `json:"extra_reserved" gorm:"not null;default:0"`
	SubscriptionId   int    `json:"subscription_id"`
	FundingSource    string `json:"funding_source" gorm:"type:varchar(32)"`
	FundingRequestId string `json:"funding_request_id" gorm:"type:varchar(128)"` // subscription request id for idempotent refund
	WalletConsumed   int    `json:"wallet_consumed" gorm:"not null;default:0"`
	TokenKey         string `json:"-" gorm:"type:varchar(128)"` // needed for quota cache; not exported in JSON
	IsPlayground     bool   `json:"is_playground" gorm:"not null;default:false"`
	WalletDone       bool   `json:"wallet_done" gorm:"not null;default:false"`
	SubscriptionDone bool   `json:"subscription_done" gorm:"not null;default:false"`
	TokenDone        bool   `json:"token_done" gorm:"not null;default:false"`
	Status           string `json:"status" gorm:"type:varchar(16);index;not null"`
	Attempts         int    `json:"attempts" gorm:"not null;default:0"`
	LastError        string `json:"last_error" gorm:"type:text"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

func (RefundIntent) TableName() string { return "refund_intents" }

// CreateRefundIntentIfAbsent 插入 pending 意图；唯一键冲突则返回已有行（created=false）。
// 约束：intent 非 nil；IdempotencyKey 必填由调用方保证。
func CreateRefundIntentIfAbsent(intent *RefundIntent) (*RefundIntent, bool, error) {
	if intent == nil {
		return nil, false, fmt.Errorf("nil refund intent")
	}
	now := time.Now().Unix()
	if intent.CreatedAt == 0 {
		intent.CreatedAt = now
	}
	intent.UpdatedAt = now
	if intent.Status == "" {
		intent.Status = RefundIntentPending
	}

	var existing RefundIntent
	err := DB.Where("idempotency_key = ?", intent.IdempotencyKey).First(&existing).Error
	if err == nil {
		return &existing, false, nil
	}
	if err := DB.Create(intent).Error; err != nil {
		// race: re-read
		if e2 := DB.Where("idempotency_key = ?", intent.IdempotencyKey).First(&existing).Error; e2 == nil {
			return &existing, false, nil
		}
		return nil, false, err
	}
	return intent, true, nil
}

// ClaimRefundIntents 将最多 limit 条 pending/failed 标为 processing 并返回（attempts+1）。
// 约束：逐行条件更新，RowsAffected=0 表示被并发抢走；非跨库事务但可防双工。
func ClaimRefundIntents(limit int) ([]*RefundIntent, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []*RefundIntent
	err := DB.Where("status IN ?", []string{RefundIntentPending, RefundIntentFailed}).
		Order("updated_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	claimed := make([]*RefundIntent, 0, len(rows))
	for _, row := range rows {
		res := DB.Model(&RefundIntent{}).
			Where("id = ? AND status IN ?", row.Id, []string{RefundIntentPending, RefundIntentFailed}).
			Updates(map[string]interface{}{
				"status":     RefundIntentProcessing,
				"attempts":   row.Attempts + 1,
				"updated_at": now,
			})
		if res.Error != nil {
			common.SysLog("claim refund intent error: " + res.Error.Error())
			continue
		}
		if res.RowsAffected == 0 {
			continue
		}
		row.Status = RefundIntentProcessing
		row.Attempts++
		row.UpdatedAt = now
		claimed = append(claimed, row)
	}
	return claimed, nil
}

// MarkRefundIntentSucceeded 标记意图成功并清空 last_error。
func MarkRefundIntentSucceeded(id int) error {
	return DB.Model(&RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     RefundIntentSucceeded,
		"last_error": "",
		"updated_at": time.Now().Unix(),
	}).Error
}

// MarkRefundIntentFailed 记录失败；attempts>=maxAttempts 时置 dead（需人工/对账）。
// 约束：本函数不修改 attempts 列；attempts 由 Claim 递增。Enqueue 直调 process 失败时 attempts 可能仍为 0。
func MarkRefundIntentFailed(id int, attempts int, maxAttempts int, errMsg string) error {
	status := RefundIntentFailed
	if attempts >= maxAttempts {
		status = RefundIntentDead
	}
	return DB.Model(&RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"last_error": errMsg,
		"updated_at": time.Now().Unix(),
	}).Error
}

// CountRefundIntentsByStatus 按状态聚合计数（运维 health / 对账）。
func CountRefundIntentsByStatus() (map[string]int64, error) {
	type row struct {
		Status string
		Cnt    int64
	}
	var rows []row
	err := DB.Model(&RefundIntent{}).Select("status, count(*) as cnt").Group("status").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]int64{}
	for _, r := range rows {
		out[r.Status] = r.Cnt
	}
	return out, nil
}

// ListRefundIntents 最近退款意图列表（管理对账）。
// 约束：limit 默认 50、上限 200；返回前清空 TokenKey，禁止密钥出站。
func ListRefundIntents(status string, limit int) ([]*RefundIntent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := DB.Model(&RefundIntent{}).Order("id DESC").Limit(limit)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []*RefundIntent
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	// Never return stored token keys in list API payloads.
	for _, r := range rows {
		if r != nil {
			r.TokenKey = ""
		}
	}
	return rows, nil
}

// ErrRefundStepAlreadyDone 表示 CAS 条件未命中（步骤已完成或意图终态）。
// 调用方应视为成功幂等，不得再次回补额度。
var ErrRefundStepAlreadyDone = errors.New("refund step already done")

func refundIntentStepCAS(tx *gorm.DB, id int, doneColumn string) error {
	now := time.Now().Unix()
	res := tx.Model(&RefundIntent{}).
		Where("id = ? AND status NOT IN ? AND "+doneColumn+" = ?",
			id,
			[]string{RefundIntentSucceeded, RefundIntentDead},
			false,
		).
		Updates(map[string]interface{}{
			doneColumn:   true,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrRefundStepAlreadyDone
	}
	return nil
}

// ApplyRefundWalletStep 在同一事务内：用户钱包额度 + wallet_done CAS。
//
// 约束：
//   - 必须同步写库（绕过 BatchUpdate），与 done 标记同事务
//   - RowsAffected=0 → ErrRefundStepAlreadyDone（崩溃恢复/并发安全）
//   - 成功后异步刷缓存（失败不影响正确性）
func ApplyRefundWalletStep(id int, userId int, quota int) error {
	if id <= 0 {
		return fmt.Errorf("invalid refund intent id")
	}
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if quota == 0 || userId <= 0 {
		return refundIntentStepCAS(DB, id, "wallet_done")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := refundIntentStepCAS(tx, id, "wallet_done"); err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", userId).
			Update("quota", gorm.Expr("quota + ?", quota)).Error
	})
	if err != nil {
		return err
	}
	gopool.Go(func() {
		if e := cacheIncrUserQuota(userId, int64(quota)); e != nil {
			common.SysLog("failed to increase user quota cache after refund: " + e.Error())
		}
	})
	return nil
}

// ApplyRefundTokenStep 在同一事务内：令牌额度回补 + token_done CAS。
// 约束同 ApplyRefundWalletStep；同步写库，禁止 batch 跳过。
func ApplyRefundTokenStep(id int, tokenId int, tokenKey string, quota int) error {
	if id <= 0 {
		return fmt.Errorf("invalid refund intent id")
	}
	if quota < 0 {
		return errors.New("quota 不能为负数！")
	}
	if quota == 0 || tokenId <= 0 {
		return refundIntentStepCAS(DB, id, "token_done")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := refundIntentStepCAS(tx, id, "token_done"); err != nil {
			return err
		}
		return tx.Model(&Token{}).Where("id = ?", tokenId).Updates(map[string]interface{}{
			"remain_quota":  gorm.Expr("remain_quota + ?", quota),
			"used_quota":    gorm.Expr("used_quota - ?", quota),
			"accessed_time": common.GetTimestamp(),
		}).Error
	})
	if err != nil {
		return err
	}
	if common.RedisEnabled && tokenKey != "" {
		key := tokenKey
		gopool.Go(func() {
			if e := cacheIncrTokenQuota(key, int64(quota)); e != nil {
				common.SysLog("failed to increase token quota cache after refund: " + e.Error())
			}
		})
	}
	return nil
}

// ApplyRefundSubscriptionExtraStep 在同一事务内：订阅额外预留回滚 + subscription_done CAS。
//
// 说明：FundingRequestId 对应的 RefundSubscriptionPreConsume 自身幂等，应在本函数之前调用。
// 本函数仅覆盖 ExtraReserved 的 AmountUsed 调整与 done 标记原子性。
func ApplyRefundSubscriptionExtraStep(id int, subscriptionId int, extraReserved int) error {
	if id <= 0 {
		return fmt.Errorf("invalid refund intent id")
	}
	if extraReserved < 0 {
		return errors.New("extraReserved 不能为负数！")
	}
	if extraReserved == 0 || subscriptionId <= 0 {
		return refundIntentStepCAS(DB, id, "subscription_done")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := refundIntentStepCAS(tx, id, "subscription_done"); err != nil {
			return err
		}
		var sub UserSubscription
		if err := lockForUpdate(tx).Where("id = ?", subscriptionId).First(&sub).Error; err != nil {
			return err
		}
		newUsed := sub.AmountUsed - int64(extraReserved)
		if newUsed < 0 {
			newUsed = 0
		}
		sub.AmountUsed = newUsed
		return tx.Save(&sub).Error
	})
}

// IsRefundStepAlreadyDone 判断是否为步骤幂等跳过。
func IsRefundStepAlreadyDone(err error) bool {
	return errors.Is(err, ErrRefundStepAlreadyDone)
}

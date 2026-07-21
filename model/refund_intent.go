package model

import (
	"fmt"
	"time"

	"github.com/xvyimu/TransitHub/common"
)

// Refund intent status values.
const (
	RefundIntentPending    = "pending"
	RefundIntentProcessing = "processing"
	RefundIntentSucceeded  = "succeeded"
	RefundIntentFailed     = "failed"
	RefundIntentDead       = "dead"
)

// RefundIntent persists a refund that must complete at least once even if the
// process restarts mid-flight (WP-C).
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

// CreateRefundIntentIfAbsent inserts a pending intent. On unique conflict returns existing row.
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

// ClaimRefundIntents marks up to limit pending/failed rows as processing and returns them.
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

func MarkRefundIntentSucceeded(id int) error {
	return DB.Model(&RefundIntent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     RefundIntentSucceeded,
		"last_error": "",
		"updated_at": time.Now().Unix(),
	}).Error
}

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

// CountRefundIntentsByStatus returns counts for ops health.
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

// ListRefundIntents returns recent intents for admin reconciliation (no token keys).
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

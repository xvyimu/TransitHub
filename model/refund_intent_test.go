package model

import (
	"testing"
	"time"

	"github.com/xvyimu/TransitHub/common"

	"github.com/stretchr/testify/require"
)

func setupRefundStepFixtures(t *testing.T) (userID, tokenID, intentID int) {
	t.Helper()
	truncateTables(t)
	require.NoError(t, DB.Exec("DELETE FROM refund_intents").Error)
	require.NoError(t, DB.Exec("DELETE FROM tokens").Error)
	require.NoError(t, DB.Exec("DELETE FROM users").Error)
	require.NoError(t, DB.Exec("DELETE FROM user_subscriptions").Error)

	oldRedis := common.RedisEnabled
	oldBatch := common.BatchUpdateEnabled
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	t.Cleanup(func() {
		common.RedisEnabled = oldRedis
		common.BatchUpdateEnabled = oldBatch
	})

	user := User{
		Username: "refund-user",
		Password: "password",
		Status:   common.UserStatusEnabled,
		Quota:    1000,
	}
	require.NoError(t, DB.Create(&user).Error)
	userID = user.Id

	token := Token{
		UserId:         userID,
		Key:            "refund-token-key",
		Status:         common.TokenStatusEnabled,
		Name:           "refund-token",
		RemainQuota:    100,
		UsedQuota:      50,
		UnlimitedQuota: false,
	}
	require.NoError(t, DB.Create(&token).Error)
	tokenID = token.Id

	now := time.Now().Unix()
	intent := RefundIntent{
		IdempotencyKey: "refund-test-wallet-token",
		TokenId:        tokenID,
		UserId:         userID,
		TokenQuota:     40,
		WalletConsumed: 30,
		TokenKey:       token.Key,
		Status:         RefundIntentProcessing,
		Attempts:       1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, DB.Create(&intent).Error)
	return userID, tokenID, intent.Id
}

func TestApplyRefundWalletStep_AtomicAndIdempotent(t *testing.T) {
	userID, _, intentID := setupRefundStepFixtures(t)

	require.NoError(t, ApplyRefundWalletStep(intentID, userID, 30))

	var user User
	require.NoError(t, DB.First(&user, userID).Error)
	require.Equal(t, 1030, user.Quota)

	var intent RefundIntent
	require.NoError(t, DB.First(&intent, intentID).Error)
	require.True(t, intent.WalletDone)

	// Second call must not double credit.
	err := ApplyRefundWalletStep(intentID, userID, 30)
	require.ErrorIs(t, err, ErrRefundStepAlreadyDone)
	require.NoError(t, DB.First(&user, userID).Error)
	require.Equal(t, 1030, user.Quota)
}

func TestApplyRefundTokenStep_AtomicAndIdempotent(t *testing.T) {
	_, tokenID, intentID := setupRefundStepFixtures(t)

	require.NoError(t, ApplyRefundTokenStep(intentID, tokenID, "refund-token-key", 40))

	var token Token
	require.NoError(t, DB.First(&token, tokenID).Error)
	require.Equal(t, 140, token.RemainQuota)
	require.Equal(t, 10, token.UsedQuota)

	var intent RefundIntent
	require.NoError(t, DB.First(&intent, intentID).Error)
	require.True(t, intent.TokenDone)

	err := ApplyRefundTokenStep(intentID, tokenID, "refund-token-key", 40)
	require.ErrorIs(t, err, ErrRefundStepAlreadyDone)
	require.NoError(t, DB.First(&token, tokenID).Error)
	require.Equal(t, 140, token.RemainQuota)
	require.Equal(t, 10, token.UsedQuota)
}

func TestApplyRefundWalletStep_RollsBackQuotaIfDoneUpdateImpossible(t *testing.T) {
	// If intent is already terminal, CAS fails and quota must not change.
	userID, _, intentID := setupRefundStepFixtures(t)
	require.NoError(t, DB.Model(&RefundIntent{}).Where("id = ?", intentID).Update("status", RefundIntentSucceeded).Error)

	err := ApplyRefundWalletStep(intentID, userID, 30)
	require.ErrorIs(t, err, ErrRefundStepAlreadyDone)

	var user User
	require.NoError(t, DB.First(&user, userID).Error)
	require.Equal(t, 1000, user.Quota)
}

func TestApplyRefundSubscriptionExtraStep_AtomicAndIdempotent(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Exec("DELETE FROM refund_intents").Error)
	require.NoError(t, DB.Exec("DELETE FROM user_subscriptions").Error)

	sub := UserSubscription{
		UserId:     1,
		PlanId:     1,
		Status:     "active",
		AmountTotal: 1000,
		AmountUsed:  200,
	}
	require.NoError(t, DB.Create(&sub).Error)

	now := time.Now().Unix()
	intent := RefundIntent{
		IdempotencyKey: "refund-test-sub-extra",
		TokenId:        1,
		UserId:         1,
		SubscriptionId: sub.Id,
		ExtraReserved:  50,
		Status:         RefundIntentProcessing,
		Attempts:       1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, DB.Create(&intent).Error)

	require.NoError(t, ApplyRefundSubscriptionExtraStep(intent.Id, sub.Id, 50))

	var got UserSubscription
	require.NoError(t, DB.First(&got, sub.Id).Error)
	require.Equal(t, int64(150), got.AmountUsed)

	var ri RefundIntent
	require.NoError(t, DB.First(&ri, intent.Id).Error)
	require.True(t, ri.SubscriptionDone)

	err := ApplyRefundSubscriptionExtraStep(intent.Id, sub.Id, 50)
	require.ErrorIs(t, err, ErrRefundStepAlreadyDone)
	require.NoError(t, DB.First(&got, sub.Id).Error)
	require.Equal(t, int64(150), got.AmountUsed)
}

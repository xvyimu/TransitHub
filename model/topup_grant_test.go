package model

import (
	"testing"

	"github.com/xvyimu/TransitHub/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRechargeWaffo_RejectsOverflowingGrant locks the P0 billing-safety
// invariant for the top-up credit path: an oversized Amount whose
// Amount*QuotaPerUnit product exceeds the int32 quota column MUST fail the
// transaction (QuotaFromDecimalStrict returns a typed clamp error) instead of
// silently saturating to MaxQuota and handing the user a near-int32-max
// credit. This protects the fix that replaced the raw
// int(decimal...IntPart()) cast in RechargeWaffo.
func TestRechargeWaffo_RejectsOverflowingGrant(t *testing.T) {
	// QuotaPerUnit is 500000; any Amount above ~4294 overflows int32 once
	// multiplied. 10000 * 500000 = 5e9 >> math.MaxInt32.
	require.Greater(t, common.QuotaPerUnit, 0.0)

	user := &User{Username: "overflow-victim", AffCode: "aff-overflow", Quota: 0, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(user).Error)

	topUp := &TopUp{
		UserId:          user.Id,
		Amount:          10000,
		Money:           10000,
		TradeNo:         "TEST-WAFFO-OVERFLOW-1",
		PaymentProvider: PaymentProviderWaffo,
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(topUp).Error)

	err := RechargeWaffo(topUp.TradeNo, "127.0.0.1")
	require.Error(t, err, "oversized top-up must fail, not silently clamp")

	// Quota must be untouched and the order must stay pending (tx rolled back).
	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	assert.Equal(t, 0, reloadedUser.Quota, "no quota may be granted on overflow")

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusPending, reloadedTopUp.Status, "order must not be marked success on failed grant")
}

// TestRechargeWaffo_GrantsInRangeAmount is the positive counterpart: a normal
// in-range top-up succeeds, credits the exact quota, and marks the order
// success. Without this the overflow test could pass against a path that
// rejects everything.
func TestRechargeWaffo_GrantsInRangeAmount(t *testing.T) {
	user := &User{Username: "normal-topup", AffCode: "aff-normal", Quota: 100, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(user).Error)

	// Amount 2 -> 2 * 500000 = 1,000,000, well within int32.
	topUp := &TopUp{
		UserId:          user.Id,
		Amount:          2,
		Money:           2,
		TradeNo:         "TEST-WAFFO-OK-1",
		PaymentProvider: PaymentProviderWaffo,
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(topUp).Error)

	err := RechargeWaffo(topUp.TradeNo, "127.0.0.1")
	require.NoError(t, err)

	expectedGrant := int(2 * common.QuotaPerUnit)
	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	assert.Equal(t, 100+expectedGrant, reloadedUser.Quota)

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusSuccess, reloadedTopUp.Status)
}

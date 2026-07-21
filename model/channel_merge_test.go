package model

import (
	"testing"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrInt64(v int64) *int64 { return &v }
func ptrUint(v uint) *uint    { return &v }
func ptrString(v string) *string {
	return &v
}

func TestNormalizeChannelHost(t *testing.T) {
	assert.Equal(t, "sub.100xlabs.space", NormalizeChannelHost("https://sub.100xlabs.space/v1"))
	assert.Equal(t, "sub.100xlabs.space", NormalizeChannelHost("http://SUB.100xlabs.space"))
	assert.Equal(t, "api.example.com", NormalizeChannelHost("api.example.com"))
	assert.Equal(t, "", NormalizeChannelHost(""))
	assert.Equal(t, "", NormalizeChannelHost("   "))
}

func TestSelectPrimaryChannel(t *testing.T) {
	a := &Channel{Id: 78, Status: common.ChannelStatusManuallyDisabled, Priority: ptrInt64(0), Weight: ptrUint(0)}
	b := &Channel{Id: 80, Status: common.ChannelStatusEnabled, Priority: ptrInt64(1), Weight: ptrUint(79)}
	c := &Channel{Id: 79, Status: common.ChannelStatusEnabled, Priority: ptrInt64(0), Weight: ptrUint(10)}
	// enabled + higher priority wins
	assert.Equal(t, 80, SelectPrimaryChannel([]*Channel{a, b, c}).Id)
	// among same priority, higher weight wins
	a2 := &Channel{Id: 78, Status: 2, Priority: ptrInt64(1), Weight: ptrUint(0)}
	b2 := &Channel{Id: 80, Status: 2, Priority: ptrInt64(1), Weight: ptrUint(79)}
	assert.Equal(t, 80, SelectPrimaryChannel([]*Channel{a2, b2}).Id)
	// same priority+weight → lower id
	a3 := &Channel{Id: 78, Status: 2, Priority: ptrInt64(1), Weight: ptrUint(10)}
	b3 := &Channel{Id: 80, Status: 2, Priority: ptrInt64(1), Weight: ptrUint(10)}
	assert.Equal(t, 78, SelectPrimaryChannel([]*Channel{a3, b3}).Id)
}

func TestUnionCSV(t *testing.T) {
	assert.Equal(t, "a,b,c", unionCSV("a,b", "b,c", "a"))
	assert.Equal(t, "gpt", unionCSV("gpt", "", "gpt,"))
}

func TestMergeKeysDedup(t *testing.T) {
	ch1 := &Channel{Key: "k1\nk2", Status: 1}
	ch2 := &Channel{Key: "k2\nk3", Status: 1}
	ch3 := &Channel{
		Key:    "k1",
		Status: common.ChannelStatusManuallyDisabled,
	}
	merged, statusList, _, _ := mergeKeys([]*Channel{ch1, ch2, ch3})
	assert.Equal(t, []string{"k1", "k2", "k3"}, merged)
	// k1 first seen from ch1 (enabled) — disabled status from ch3 must not overwrite first occurrence
	_, has := statusList[0]
	assert.False(t, has)
}

func TestChannelMerge_EndToEnd(t *testing.T) {
	truncateTables(t)

	base := "https://sub.100xlabs.space"
	chs := []*Channel{
		{
			Id: 0, Type: 1, Name: "100x", Key: "key-a", Status: common.ChannelStatusManuallyDisabled,
			BaseURL: ptrString(base), Models: "m1,m2", Group: "default",
			Priority: ptrInt64(0), Weight: ptrUint(0), UsedQuota: 10,
		},
		{
			Id: 0, Type: 1, Name: "100x", Key: "key-b", Status: common.ChannelStatusManuallyDisabled,
			BaseURL: ptrString(base), Models: "m2,m3", Group: "default,vip",
			Priority: ptrInt64(1), Weight: ptrUint(79), UsedQuota: 20,
		},
		{
			Id: 0, Type: 1, Name: "100x", Key: "key-a\nkey-c", Status: common.ChannelStatusEnabled,
			BaseURL: ptrString(base), Models: "m1", Group: "default",
			Priority: ptrInt64(0), Weight: ptrUint(5), UsedQuota: 5,
			ChannelInfo: ChannelInfo{IsMultiKey: true, MultiKeySize: 2, MultiKeyMode: constant.MultiKeyModePolling},
		},
	}
	for _, ch := range chs {
		require.NoError(t, ch.Insert())
	}

	groups, err := FindDuplicateChannelGroups()
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, 3, groups[0].Count)
	assert.Equal(t, "100x", groups[0].Name)
	assert.Equal(t, "sub.100xlabs.space", groups[0].Host)

	ids := []int{chs[0].Id, chs[1].Id, chs[2].Id}
	// primary should prefer enabled channel (chs[2])
	preview, err := PreviewChannelMerge(ids, 0)
	require.NoError(t, err)
	assert.Equal(t, chs[2].Id, preview.PrimaryId)
	assert.Equal(t, 3, preview.MergedKeyCount) // key-a, key-c, key-b
	assert.ElementsMatch(t, []int{chs[0].Id, chs[1].Id}, preview.DeleteIds)

	// override primary
	preview2, err := PreviewChannelMerge(ids, chs[1].Id)
	require.NoError(t, err)
	assert.Equal(t, chs[1].Id, preview2.PrimaryId)

	result, err := MergeChannels(ids, chs[1].Id)
	require.NoError(t, err)
	assert.Equal(t, chs[1].Id, result.PrimaryId)
	assert.Equal(t, 3, result.MergedKeyCount)
	assert.ElementsMatch(t, []int{chs[0].Id, chs[2].Id}, result.DeletedIds)

	// deleted gone
	_, err = GetChannelById(chs[0].Id, true)
	assert.Error(t, err)
	_, err = GetChannelById(chs[2].Id, true)
	assert.Error(t, err)

	// primary is multi-key with union metadata
	primary, err := GetChannelById(chs[1].Id, true)
	require.NoError(t, err)
	assert.True(t, primary.ChannelInfo.IsMultiKey)
	assert.Equal(t, 3, primary.ChannelInfo.MultiKeySize)
	assert.Equal(t, []string{"key-b", "key-a", "key-c"}, primary.GetKeys()) // primary keys first
	assert.Equal(t, "m2,m3,m1", primary.Models)
	assert.Equal(t, "default,vip", primary.Group)
	assert.EqualValues(t, 1, primary.GetPriority())
	assert.Equal(t, 79, primary.GetWeight())
	assert.Equal(t, common.ChannelStatusEnabled, primary.Status) // any enabled in set
	assert.EqualValues(t, 35, primary.UsedQuota)

	// abilities only for primary
	var abilityCount int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", primary.Id).Count(&abilityCount).Error)
	assert.Greater(t, abilityCount, int64(0))
	var orphan int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id IN ?", result.DeletedIds).Count(&orphan).Error)
	assert.EqualValues(t, 0, orphan)

	// no more duplicate groups
	groups, err = FindDuplicateChannelGroups()
	require.NoError(t, err)
	assert.Empty(t, groups)
}

func TestChannelMerge_RejectMismatch(t *testing.T) {
	truncateTables(t)

	a := &Channel{
		Type: 1, Name: "foo", Key: "k1", Status: 1,
		BaseURL: ptrString("https://a.example.com"), Models: "m1", Group: "default",
	}
	b := &Channel{
		Type: 1, Name: "foo", Key: "k2", Status: 1,
		BaseURL: ptrString("https://b.example.com"), Models: "m1", Group: "default",
	}
	require.NoError(t, a.Insert())
	require.NoError(t, b.Insert())

	_, err := PreviewChannelMerge([]int{a.Id, b.Id}, 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrChannelMergeMismatch)

	_, err = MergeChannels([]int{a.Id}, 0)
	require.ErrorIs(t, err, ErrChannelMergeTooFew)

	_, err = PreviewChannelMerge([]int{a.Id, b.Id}, 99999)
	// not found first because load checks count — primary not in set only if both load
	// With host mismatch we fail earlier; create same host for primary test
	c := &Channel{
		Type: 1, Name: "foo", Key: "k3", Status: 1,
		BaseURL: ptrString("https://a.example.com"), Models: "m1", Group: "default",
	}
	require.NoError(t, c.Insert())
	_, err = PreviewChannelMerge([]int{a.Id, c.Id}, 99999)
	require.ErrorIs(t, err, ErrChannelMergePrimary)
}

func TestChannelMerge_TypeMismatch(t *testing.T) {
	truncateTables(t)
	a := &Channel{
		Type: 1, Name: "x", Key: "k1", Status: 1,
		BaseURL: ptrString("https://h.example.com"), Models: "m", Group: "default",
	}
	b := &Channel{
		Type: 2, Name: "x", Key: "k2", Status: 1,
		BaseURL: ptrString("https://h.example.com"), Models: "m", Group: "default",
	}
	require.NoError(t, a.Insert())
	require.NoError(t, b.Insert())
	_, err := MergeChannels([]int{a.Id, b.Id}, 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrChannelMergeMismatch)
}

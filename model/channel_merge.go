package model

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/constant"

	"gorm.io/gorm"
)

// DuplicateChannelSummary is a key-free view of a channel inside a duplicate group.
type DuplicateChannelSummary struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Status      int    `json:"status"`
	Priority    int64  `json:"priority"`
	Weight      int    `json:"weight"`
	KeyCount    int    `json:"key_count"`
	ModelsCount int    `json:"models_count"`
	Group       string `json:"group"`
	BaseURL     string `json:"base_url"`
	IsMultiKey  bool   `json:"is_multi_key"`
}

// DuplicateChannelGroup is a set of channels that share name + host + type.
type DuplicateChannelGroup struct {
	GroupKey           string                    `json:"group_key"`
	Name               string                    `json:"name"`
	Host               string                    `json:"host"`
	Type               int                       `json:"type"`
	Count              int                       `json:"count"`
	SuggestedPrimaryId int                       `json:"suggested_primary_id"`
	Channels           []DuplicateChannelSummary `json:"channels"`
}

// ChannelMergePreview describes the effect of merging without applying it.
type ChannelMergePreview struct {
	GroupKey       string                    `json:"group_key"`
	Name           string                    `json:"name"`
	Host           string                    `json:"host"`
	Type           int                       `json:"type"`
	PrimaryId      int                       `json:"primary_id"`
	DeleteIds      []int                     `json:"delete_ids"`
	MergedKeyCount int                       `json:"merged_key_count"`
	ModelsCount    int                       `json:"models_count"`
	Groups         string                    `json:"groups"`
	Priority       int64                     `json:"priority"`
	Weight         int                       `json:"weight"`
	Status         int                       `json:"status"`
	Channels       []DuplicateChannelSummary `json:"channels"`
}

// ChannelMergeResult is returned after a successful merge.
type ChannelMergeResult struct {
	PrimaryId      int   `json:"primary_id"`
	MergedKeyCount int   `json:"merged_key_count"`
	DeletedIds     []int `json:"deleted_ids"`
	ModelsCount    int   `json:"models_count"`
}

// channelMergePlan is the shared intermediate for preview and apply.
type channelMergePlan struct {
	Preview    *ChannelMergePreview
	Primary    *Channel
	MergedKeys []string
	StatusList map[int]int
	ReasonList map[int]string
	TimeList   map[int]int64
	Models     string
	Groups     string
	UsedQuota  int64
}

var (
	ErrChannelMergeTooFew    = errors.New("at least two channels are required to merge")
	ErrChannelMergeNotFound  = errors.New("one or more channels were not found")
	ErrChannelMergeMismatch  = errors.New("channels must share the same name, host, and type")
	ErrChannelMergeEmptyHost = errors.New("channels without a resolvable host cannot be merged")
	ErrChannelMergePrimary   = errors.New("primary_id must be one of the merge candidates")
	ErrChannelMergeNoKeys    = errors.New("merged channel would have no keys")
)

// NormalizeChannelHost extracts a lowercase host from base_url.
// Empty host means the channel is excluded from duplicate discovery.
func NormalizeChannelHost(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "https://" + baseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(u.Host))
	if host == "" {
		host = strings.ToLower(strings.TrimSpace(u.Path))
		if i := strings.Index(host, "/"); i >= 0 {
			host = host[:i]
		}
	}
	return host
}

// ChannelDuplicateGroupKey builds the grouping key: name|host|type.
func ChannelDuplicateGroupKey(name, host string, channelType int) string {
	return strings.TrimSpace(name) + "|" + strings.ToLower(strings.TrimSpace(host)) + "|" + strconv.Itoa(channelType)
}

func channelModelsCount(models string) int {
	n := 0
	for _, m := range strings.Split(models, ",") {
		if strings.TrimSpace(m) != "" {
			n++
		}
	}
	return n
}

func channelKeyCount(ch *Channel) int {
	if ch.Key != "" {
		n := 0
		for _, k := range ch.GetKeys() {
			if strings.TrimSpace(k) != "" {
				n++
			}
		}
		return n
	}
	// Key omitted (discovery query) or empty: prefer multi-key metadata.
	if ch.ChannelInfo.IsMultiKey {
		return ch.ChannelInfo.MultiKeySize
	}
	return 1
}

func summarizeChannel(ch *Channel) DuplicateChannelSummary {
	return DuplicateChannelSummary{
		Id:          ch.Id,
		Name:        ch.Name,
		Status:      ch.Status,
		Priority:    ch.GetPriority(),
		Weight:      ch.GetWeight(),
		KeyCount:    channelKeyCount(ch),
		ModelsCount: channelModelsCount(ch.Models),
		Group:       ch.Group,
		BaseURL:     ch.GetBaseURL(),
		IsMultiKey:  ch.ChannelInfo.IsMultiKey,
	}
}

// SelectPrimaryChannel picks the default survivor:
// enabled first, then higher priority, higher weight, lower id.
func SelectPrimaryChannel(channels []*Channel) *Channel {
	if len(channels) == 0 {
		return nil
	}
	best := channels[0]
	for _, ch := range channels[1:] {
		if channelPrimaryBetter(ch, best) {
			best = ch
		}
	}
	return best
}

func channelPrimaryBetter(a, b *Channel) bool {
	aEnabled := a.Status == common.ChannelStatusEnabled
	bEnabled := b.Status == common.ChannelStatusEnabled
	if aEnabled != bEnabled {
		return aEnabled
	}
	if a.GetPriority() != b.GetPriority() {
		return a.GetPriority() > b.GetPriority()
	}
	if a.GetWeight() != b.GetWeight() {
		return a.GetWeight() > b.GetWeight()
	}
	return a.Id < b.Id
}

func unionCSV(parts ...string) string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, part := range parts {
		for _, item := range strings.Split(part, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	return strings.Join(out, ",")
}

type keyStatusEntry struct {
	status int
	reason string
	time   int64
	has    bool
}

func collectKeysWithStatus(ch *Channel) ([]string, map[string]keyStatusEntry) {
	keys := make([]string, 0)
	statusByKey := make(map[string]keyStatusEntry)
	for i, k := range ch.GetKeys() {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		keys = append(keys, k)
		entry := keyStatusEntry{status: common.ChannelStatusEnabled}
		if ch.ChannelInfo.IsMultiKey && ch.ChannelInfo.MultiKeyStatusList != nil {
			if st, ok := ch.ChannelInfo.MultiKeyStatusList[i]; ok {
				entry.status = st
				entry.has = true
				if ch.ChannelInfo.MultiKeyDisabledReason != nil {
					entry.reason = ch.ChannelInfo.MultiKeyDisabledReason[i]
				}
				if ch.ChannelInfo.MultiKeyDisabledTime != nil {
					entry.time = ch.ChannelInfo.MultiKeyDisabledTime[i]
				}
			}
		} else if ch.Status != common.ChannelStatusEnabled {
			entry.status = ch.Status
			entry.has = true
		}
		if _, exists := statusByKey[k]; !exists {
			statusByKey[k] = entry
		}
	}
	return keys, statusByKey
}

func mergeKeys(channels []*Channel) (merged []string, statusList map[int]int, reasonList map[int]string, timeList map[int]int64) {
	seen := make(map[string]struct{})
	merged = make([]string, 0)
	statusByKey := make(map[string]keyStatusEntry)

	for _, ch := range channels {
		keys, st := collectKeysWithStatus(ch)
		for _, k := range keys {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			merged = append(merged, k)
			if entry, ok := st[k]; ok {
				statusByKey[k] = entry
			}
		}
	}

	statusList = make(map[int]int)
	reasonList = make(map[int]string)
	timeList = make(map[int]int64)
	for i, k := range merged {
		entry, ok := statusByKey[k]
		if !ok || !entry.has || entry.status == common.ChannelStatusEnabled {
			continue
		}
		statusList[i] = entry.status
		if entry.reason != "" {
			reasonList[i] = entry.reason
		}
		if entry.time != 0 {
			timeList[i] = entry.time
		}
	}
	return merged, statusList, reasonList, timeList
}

func loadChannelsByIDs(ids []int) ([]*Channel, error) {
	if len(ids) == 0 {
		return nil, ErrChannelMergeTooFew
	}
	uniq := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) < 2 {
		return nil, ErrChannelMergeTooFew
	}

	channels, err := GetChannelsByIds(uniq)
	if err != nil {
		return nil, err
	}
	if len(channels) != len(uniq) {
		return nil, ErrChannelMergeNotFound
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].Id < channels[j].Id })
	return channels, nil
}

func validateMergeGroup(channels []*Channel) (name, host string, channelType int, err error) {
	if len(channels) < 2 {
		return "", "", 0, ErrChannelMergeTooFew
	}
	name = strings.TrimSpace(channels[0].Name)
	host = NormalizeChannelHost(channels[0].GetBaseURL())
	channelType = channels[0].Type
	if host == "" {
		return "", "", 0, ErrChannelMergeEmptyHost
	}
	for _, ch := range channels[1:] {
		if strings.TrimSpace(ch.Name) != name {
			return "", "", 0, fmt.Errorf("%w: name mismatch", ErrChannelMergeMismatch)
		}
		h := NormalizeChannelHost(ch.GetBaseURL())
		if h == "" {
			return "", "", 0, ErrChannelMergeEmptyHost
		}
		if h != host {
			return "", "", 0, fmt.Errorf("%w: host mismatch", ErrChannelMergeMismatch)
		}
		if ch.Type != channelType {
			return "", "", 0, fmt.Errorf("%w: type mismatch", ErrChannelMergeMismatch)
		}
	}
	return name, host, channelType, nil
}

func pickPrimary(channels []*Channel, primaryId int) (*Channel, error) {
	if primaryId == 0 {
		return SelectPrimaryChannel(channels), nil
	}
	for _, ch := range channels {
		if ch.Id == primaryId {
			return ch, nil
		}
	}
	return nil, ErrChannelMergePrimary
}

func orderChannelsPrimaryFirst(channels []*Channel, primary *Channel) []*Channel {
	out := make([]*Channel, 0, len(channels))
	out = append(out, primary)
	for _, ch := range channels {
		if ch.Id != primary.Id {
			out = append(out, ch)
		}
	}
	return out
}

func prepareChannelMerge(ids []int, primaryId int) (*channelMergePlan, error) {
	channels, err := loadChannelsByIDs(ids)
	if err != nil {
		return nil, err
	}
	primary, err := pickPrimary(channels, primaryId)
	if err != nil {
		return nil, err
	}
	name, host, channelType, err := validateMergeGroup(channels)
	if err != nil {
		return nil, err
	}

	ordered := orderChannelsPrimaryFirst(channels, primary)
	mergedKeys, statusList, reasonList, timeList := mergeKeys(ordered)
	if len(mergedKeys) == 0 {
		return nil, ErrChannelMergeNoKeys
	}

	deleteIds := make([]int, 0, len(channels)-1)
	summaries := make([]DuplicateChannelSummary, 0, len(channels))
	var maxPriority int64
	var maxWeight int
	anyEnabled := false
	modelsParts := make([]string, 0, len(channels))
	groupParts := make([]string, 0, len(channels))
	var usedQuota int64

	for _, ch := range ordered {
		summaries = append(summaries, summarizeChannel(ch))
		if ch.Id != primary.Id {
			deleteIds = append(deleteIds, ch.Id)
		}
		if ch.GetPriority() > maxPriority {
			maxPriority = ch.GetPriority()
		}
		if ch.GetWeight() > maxWeight {
			maxWeight = ch.GetWeight()
		}
		if ch.Status == common.ChannelStatusEnabled {
			anyEnabled = true
		}
		modelsParts = append(modelsParts, ch.Models)
		groupParts = append(groupParts, ch.Group)
		usedQuota += ch.UsedQuota
	}
	sort.Ints(deleteIds)

	status := primary.Status
	if anyEnabled {
		status = common.ChannelStatusEnabled
	}
	models := unionCSV(modelsParts...)
	groups := unionCSV(groupParts...)

	return &channelMergePlan{
		Preview: &ChannelMergePreview{
			GroupKey:       ChannelDuplicateGroupKey(name, host, channelType),
			Name:           name,
			Host:           host,
			Type:           channelType,
			PrimaryId:      primary.Id,
			DeleteIds:      deleteIds,
			MergedKeyCount: len(mergedKeys),
			ModelsCount:    channelModelsCount(models),
			Groups:         groups,
			Priority:       maxPriority,
			Weight:         maxWeight,
			Status:         status,
			Channels:       summaries,
		},
		Primary:    primary,
		MergedKeys: mergedKeys,
		StatusList: statusList,
		ReasonList: reasonList,
		TimeList:   timeList,
		Models:     models,
		Groups:     groups,
		UsedQuota:  usedQuota,
	}, nil
}

// FindDuplicateChannelGroups returns groups with 2+ channels sharing name+host+type.
// Keys are omitted: discovery only needs counts/metadata for the admin UI.
func FindDuplicateChannelGroups() ([]DuplicateChannelGroup, error) {
	var channels []*Channel
	if err := DB.Omit("key").Find(&channels).Error; err != nil {
		return nil, err
	}

	type bucket struct {
		name  string
		host  string
		typ   int
		items []*Channel
	}
	groups := make(map[string]*bucket)

	for _, ch := range channels {
		host := NormalizeChannelHost(ch.GetBaseURL())
		if host == "" {
			continue
		}
		name := strings.TrimSpace(ch.Name)
		key := ChannelDuplicateGroupKey(name, host, ch.Type)
		b, ok := groups[key]
		if !ok {
			b = &bucket{name: name, host: host, typ: ch.Type}
			groups[key] = b
		}
		b.items = append(b.items, ch)
	}

	out := make([]DuplicateChannelGroup, 0)
	for key, b := range groups {
		if len(b.items) < 2 {
			continue
		}
		sort.Slice(b.items, func(i, j int) bool { return b.items[i].Id < b.items[j].Id })
		primary := SelectPrimaryChannel(b.items)
		summaries := make([]DuplicateChannelSummary, 0, len(b.items))
		for _, ch := range b.items {
			summaries = append(summaries, summarizeChannel(ch))
		}
		out = append(out, DuplicateChannelGroup{
			GroupKey:           key,
			Name:               b.name,
			Host:               b.host,
			Type:               b.typ,
			Count:              len(b.items),
			SuggestedPrimaryId: primary.Id,
			Channels:           summaries,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].GroupKey < out[j].GroupKey
	})
	return out, nil
}

// PreviewChannelMerge validates ids and returns a merge plan.
func PreviewChannelMerge(ids []int, primaryId int) (*ChannelMergePreview, error) {
	plan, err := prepareChannelMerge(ids, primaryId)
	if err != nil {
		return nil, err
	}
	return plan.Preview, nil
}

// MergeChannels merges duplicate channels into a multi-key primary and deletes the rest.
func MergeChannels(ids []int, primaryId int) (*ChannelMergeResult, error) {
	plan, err := prepareChannelMerge(ids, primaryId)
	if err != nil {
		return nil, err
	}

	primary := plan.Primary
	mode := primary.ChannelInfo.MultiKeyMode
	if mode == "" {
		mode = constant.MultiKeyModeRandom
	}

	priority := plan.Preview.Priority
	weight := uint(plan.Preview.Weight)
	primary.Key = strings.Join(plan.MergedKeys, "\n")
	primary.Models = plan.Models
	primary.Group = plan.Groups
	primary.Status = plan.Preview.Status
	primary.Priority = &priority
	primary.Weight = &weight
	primary.UsedQuota = plan.UsedQuota
	primary.ChannelInfo.IsMultiKey = true
	primary.ChannelInfo.MultiKeySize = len(plan.MergedKeys)
	primary.ChannelInfo.MultiKeyMode = mode
	primary.ChannelInfo.MultiKeyStatusList = plan.StatusList
	primary.ChannelInfo.MultiKeyDisabledReason = plan.ReasonList
	primary.ChannelInfo.MultiKeyDisabledTime = plan.TimeList
	primary.ChannelInfo.MultiKeyPollingIndex = 0
	primary.Keys = nil

	deleteIds := plan.Preview.DeleteIds
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(primary).Select(
			"key", "models", "group", "status", "priority", "weight", "used_quota", "channel_info",
		).Updates(primary).Error; err != nil {
			return err
		}
		if err := primary.UpdateAbilities(tx); err != nil {
			return err
		}
		if len(deleteIds) == 0 {
			return nil
		}
		if err := tx.Where("id IN ?", deleteIds).Delete(&Channel{}).Error; err != nil {
			return err
		}
		return tx.Where("channel_id IN ?", deleteIds).Delete(&Ability{}).Error
	})
	if err != nil {
		return nil, err
	}

	return &ChannelMergeResult{
		PrimaryId:      primary.Id,
		MergedKeyCount: len(plan.MergedKeys),
		DeletedIds:     deleteIds,
		ModelsCount:    plan.Preview.ModelsCount,
	}, nil
}

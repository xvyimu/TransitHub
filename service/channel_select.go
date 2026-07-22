package service

import (
	"errors"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/constant"
	"github.com/xvyimu/TransitHub/logger"
	"github.com/xvyimu/TransitHub/model"
	"github.com/xvyimu/TransitHub/setting"
	"github.com/gin-gonic/gin"
)

type RetryParam struct {
	Ctx          *gin.Context
	TokenGroup   string
	ModelName    string
	RequestPath  string
	Retry        *int
	resetNextTry bool
}

func (p *RetryParam) GetRetry() int {
	if p.Retry == nil {
		return 0
	}
	return *p.Retry
}

func (p *RetryParam) SetRetry(retry int) {
	p.Retry = &retry
}

func (p *RetryParam) IncreaseRetry() {
	if p.resetNextTry {
		p.resetNextTry = false
		return
	}
	if p.Retry == nil {
		p.Retry = new(int)
	}
	*p.Retry++
}

func (p *RetryParam) ResetRetryNextTry() {
	p.resetNextTry = true
}

// CacheGetRandomSatisfiedChannel 选路统一入口（中继重试环应只调此函数）。
//
// 行为：
//   - 自适应开启或 Shadow 开启 → AdaptiveSelectChannel
//   - 否则 → 旧版随机/auto 分组选路
//
// 约束：
//   - AdaptiveSelectChannel 内部回退禁止再调用本函数（防无限递归），只许调
//     cacheGetRandomSatisfiedChannelLegacy
//   - Shadow 模式下真实返回值仍是 legacy 渠道（见 AdaptiveSelectChannel）
//
// auto 分组跨组重试细节见 cacheGetRandomSatisfiedChannelLegacy 实现注释。
func CacheGetRandomSatisfiedChannel(param *RetryParam) (*model.Channel, string, error) {
	// Adaptive entry. AdaptiveSelectChannel must only call
	// cacheGetRandomSatisfiedChannelLegacy — never this function — or flags
	// cause infinite recursion / stack overflow.
	if constant.AdaptiveBalanceEnabled || constant.AdaptiveBalanceShadowMode {
		return AdaptiveSelectChannel(param)
	}
	return cacheGetRandomSatisfiedChannelLegacy(param)
}

// cacheGetRandomSatisfiedChannelLegacy 旧版随机 / auto 分组选路（无动态评分）。
//
// 约束：自适应回退与候选收集必须调用本函数，禁止调用 CacheGetRandomSatisfiedChannel。
//
// auto + 跨组重试：每组耗尽优先级后再切下一组；用 ContextKeyAutoGroupIndex /
// AutoGroupRetryIndex 跟踪；priorityRetry 表示组内优先级层级。
func cacheGetRandomSatisfiedChannelLegacy(param *RetryParam) (*model.Channel, string, error) {
	var channel *model.Channel
	var err error
	selectGroup := param.TokenGroup
	userGroup := common.GetContextKeyString(param.Ctx, constant.ContextKeyUserGroup)

	if param.TokenGroup == "auto" {
		if len(setting.GetAutoGroups()) == 0 {
			return nil, selectGroup, errors.New("auto groups is not enabled")
		}
		autoGroups := GetUserAutoGroup(userGroup)

		// startGroupIndex: the group index to start searching from
		// startGroupIndex: 开始搜索的分组索引
		startGroupIndex := 0
		crossGroupRetry := common.GetContextKeyBool(param.Ctx, constant.ContextKeyTokenCrossGroupRetry)

		if lastGroupIndex, exists := common.GetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex); exists {
			if idx, ok := lastGroupIndex.(int); ok {
				startGroupIndex = idx
			}
		}

		for i := startGroupIndex; i < len(autoGroups); i++ {
			autoGroup := autoGroups[i]
			// Calculate priorityRetry for current group
			// 计算当前分组的 priorityRetry
			priorityRetry := param.GetRetry()
			// If moved to a new group, reset priorityRetry and update startRetryIndex
			// 如果切换到新分组，重置 priorityRetry 并更新 startRetryIndex
			if i > startGroupIndex {
				priorityRetry = 0
			}
			logger.LogDebug(param.Ctx, "Auto selecting group: %s, priorityRetry: %d", autoGroup, priorityRetry)

			channel, _ = model.GetRandomSatisfiedChannelExcluding(autoGroup, param.ModelName, priorityRetry, param.RequestPath, adaptiveUsedChannelSet(param.Ctx))
			if channel == nil {
				// Current group has no available channel for this model, try next group
				// 当前分组没有该模型的可用渠道，尝试下一个分组
				logger.LogDebug(param.Ctx, "No available channel in group %s for model %s at priorityRetry %d, trying next group", autoGroup, param.ModelName, priorityRetry)
				// 重置状态以尝试下一个分组
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i+1)
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupRetryIndex, 0)
				// Reset retry counter so outer loop can continue for next group
				// 重置重试计数器，以便外层循环可以为下一个分组继续
				param.SetRetry(0)
				continue
			}
			common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroup, autoGroup)
			selectGroup = autoGroup
			logger.LogDebug(param.Ctx, "Auto selected group: %s", autoGroup)

			// Prepare state for next retry
			// 为下一次重试准备状态
			if crossGroupRetry && priorityRetry >= common.RetryTimes {
				// Current group has exhausted all retries, prepare to switch to next group
				// This request still uses current group, but next retry will use next group
				// 当前分组已用完所有重试次数，准备切换到下一个分组
				// 本次请求仍使用当前分组，但下次重试将使用下一个分组
				logger.LogDebug(param.Ctx, "Current group %s retries exhausted (priorityRetry=%d >= RetryTimes=%d), preparing switch to next group for next retry", autoGroup, priorityRetry, common.RetryTimes)
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i+1)
				// Reset retry counter so outer loop can continue for next group
				// 重置重试计数器，以便外层循环可以为下一个分组继续
				param.SetRetry(0)
				param.ResetRetryNextTry()
			} else {
				// Stay in current group, save current state
				// 保持在当前分组，保存当前状态
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i)
			}
			break
		}
	} else {
		channel, err = model.GetRandomSatisfiedChannelExcluding(param.TokenGroup, param.ModelName, param.GetRetry(), param.RequestPath, adaptiveUsedChannelSet(param.Ctx))
		if err != nil {
			return nil, param.TokenGroup, err
		}
	}
	return channel, selectGroup, nil
}

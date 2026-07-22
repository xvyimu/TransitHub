package service

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/constant"
	"github.com/xvyimu/TransitHub/logger"
	"github.com/xvyimu/TransitHub/model"
	"github.com/gin-gonic/gin"
)

var adaptiveLogSample = 0.05 // shadow-mode log sample rate (Info-level)

// 请求上下文 key
type adaptiveContextKey string

const (
	ctxKeyAdaptiveUsedChannels  adaptiveContextKey = "adaptive_used_channels"
	ctxKeyAdaptiveGroup         adaptiveContextKey = "adaptive_group"
	ctxKeyAdaptiveModel         adaptiveContextKey = "adaptive_model"
	ctxKeyAdaptiveSelected      adaptiveContextKey = "adaptive_selected"
	ctxKeyAdaptiveScores        adaptiveContextKey = "adaptive_scores"
	ctxKeyAdaptiveCircuitPermit adaptiveContextKey = "adaptive_circuit_permit"
)

// AdaptiveSelectChannel 动态评分选路主入口。
//
// 行为：
//   - 关闭自适应且非 Shadow → 直接 legacy
//   - 否则：拉候选 → ScoreCandidates → 过滤熔断/已用 → TopK 加权
//   - Shadow：仍返回 legacy 选中渠道，仅 RecordShadowCompare + 采样日志
//   - 真切流：返回评分选中渠道，并持有 half-open CircuitPermit
//
// 约束：
//   - 所有回退只许 cacheGetRandomSatisfiedChannelLegacy（禁止再进 CacheGetRandomSatisfiedChannel）
//   - Shadow 不得 AcquireCircuitPermit，也不得改变真实路由
//   - 关 Shadow 需观察证据 + 人类书面批准（运营硬顶，非本函数参数）
//
// 参见：channel_select.go、channel_circuit.go、channel_score.go、controller/relay.go
func AdaptiveSelectChannel(param *RetryParam) (*model.Channel, string, error) {
	ctx := param.Ctx

	// Enter scoring when fully enabled OR shadow-only (A/B compare logs).
	// Shadow still returns legacy routing below — it must not skip scoring entirely.
	if !constant.AdaptiveBalanceEnabled && !constant.AdaptiveBalanceShadowMode {
		return cacheGetRandomSatisfiedChannelLegacy(param)
	}

	// 提取 group 和 model
	group := common.GetContextKeyString(ctx, constant.ContextKeyUsingGroup)
	if group == "" {
		group = param.TokenGroup
	}
	modelName := param.ModelName

	// 获取该 group+model 下的可用渠道
	channels, err := getCandidateChannels(group, modelName, param)
	if err != nil {
		return nil, group, err
	}
	if len(channels) == 0 {
		return cacheGetRandomSatisfiedChannelLegacy(param)
	}

	// 获取亲和偏好 channel
	preferredID := getPreferredChannelID(ctx, modelName, group)

	// 评分
	candidates := ScoreCandidates(channels, group, modelName, preferredID)

	usedIDs := getAdaptiveUsedChannels(ctx)
	filtered, permits := filterAdaptiveCandidates(
		candidates, group, modelName, preferredID, usedIDs, constant.AdaptiveBalanceShadowMode,
	)

	if len(filtered) == 0 {
		if shouldFallbackToLegacy(len(candidates), len(filtered), constant.AdaptiveBalanceShadowMode) {
			return cacheGetRandomSatisfiedChannelLegacy(param)
		}
		return nil, group, fmt.Errorf("adaptive: no available channels after circuit and retry filtering")
	}

	// topK 加权随机选择
	selected := SelectTopKWeighted(filtered, 3)
	if selected == nil {
		releaseUnselectedCircuitPermits(permits, 0)
		return nil, group, fmt.Errorf("adaptive: failed to select an eligible channel")
	}

	// Shadow Mode：选择仍走旧逻辑，仅记录对比
	if constant.AdaptiveBalanceShadowMode {
		oldCh, oldGroup, oldErr := cacheGetRandomSatisfiedChannelLegacy(param)
		agree := oldCh != nil && selected != nil && selected.Channel != nil && oldCh.Id == selected.Channel.Id
		RecordShadowCompare(agree)

		// 采样日志
		if randFloat64() < adaptiveLogSample {
			logAdaptiveCompare(ctx, modelName, group, selected, oldCh)
		}

		// shadow mode never changes routing or acquires half-open permits.
		if oldCh != nil {
			addAdaptiveUsedChannel(ctx, oldCh.Id)
			storeAdaptiveSelection(ctx, selected.Channel, group, candidates)
		}
		return oldCh, oldGroup, oldErr
	}

	// 正常模式：使用动态选择的渠道
	selectGroup := group
	ch := selected.Channel
	permit := permits[ch.Id]
	releaseUnselectedCircuitPermits(permits, ch.Id)

	addAdaptiveUsedChannel(ctx, ch.Id)
	storeAdaptiveSelection(ctx, ch, group, candidates)
	ctx.Set(string(ctxKeyAdaptiveCircuitPermit), permit)

	logger.LogDebug(ctx, "adaptive selected channel #%d (score=%.3f) for group=%s model=%s",
		ch.Id, selected.Score, group, modelName)

	return ch, selectGroup, nil
}

func filterAdaptiveCandidates(
	candidates []CandidateScore,
	group string,
	modelName string,
	preferredID int,
	usedIDs []int,
	shadowMode bool,
) ([]CandidateScore, map[int]CircuitPermit) {
	filtered := make([]CandidateScore, 0, len(candidates))
	permits := make(map[int]CircuitPermit, len(candidates))
	for _, candidate := range candidates {
		channelID := candidate.Channel.Id
		if containsInt(usedIDs, channelID) {
			continue
		}
		if shadowMode {
			if IsCircuitOpen(channelID) || candidate.Score <= 0 {
				continue
			}
			filtered = append(filtered, candidate)
			continue
		}

		permit, ok := AcquireCircuitPermit(channelID)
		if !ok {
			continue
		}
		if permit.HalfOpen {
			candidate = scoreCandidate(candidate.Channel, group, modelName, preferredID, 0.5)
		}
		if candidate.Score <= 0 {
			ReleaseCircuitPermit(permit)
			continue
		}
		permits[channelID] = permit
		filtered = append(filtered, candidate)
	}
	return filtered, permits
}

// ReleaseAdaptiveCircuitPermit 释放请求上下文中持有的熔断探测许可。
//
// 行为：channelID>0 时要求与 permit 渠道一致；channelID<=0 释放当前持有的任意 permit
// （客户端取消时尚不知渠道 id 时使用）。
//
// 约束：仅真切流会写入 ctxKeyAdaptiveCircuitPermit；Shadow 路径不应持有 permit。
func ReleaseAdaptiveCircuitPermit(c *gin.Context, channelID int) {
	if c == nil {
		return
	}
	permitAny, ok := c.Get(string(ctxKeyAdaptiveCircuitPermit))
	permit, permitOK := permitAny.(CircuitPermit)
	if !ok || !permitOK || permit.ChannelID <= 0 {
		return
	}
	if channelID > 0 && permit.ChannelID != channelID {
		return
	}
	ReleaseCircuitPermit(permit)
	c.Set(string(ctxKeyAdaptiveCircuitPermit), CircuitPermit{})
}

func shouldFallbackToLegacy(candidateCount, filteredCount int, shadowMode bool) bool {
	return candidateCount == 0 || (shadowMode && filteredCount == 0)
}

func releaseUnselectedCircuitPermits(permits map[int]CircuitPermit, selectedChannelID int) {
	for channelID, permit := range permits {
		if channelID != selectedChannelID {
			ReleaseCircuitPermit(permit)
		}
	}
}

// getCandidateChannels 获取 group+model 全部候选（非单渠道路由）
func getCandidateChannels(group, modelName string, param *RetryParam) ([]*model.Channel, error) {
	// auto 分组：优先用上下文已解析的 auto group，否则 legacy 解析一次
	if group == "auto" || param.TokenGroup == "auto" {
		if g := common.GetContextKeyString(param.Ctx, constant.ContextKeyAutoGroup); g != "" {
			group = g
		} else {
			// 用 legacy 解析 auto → 具体 group，再拉全量候选
			ch, selectGroup, err := cacheGetRandomSatisfiedChannelLegacy(param)
			if err != nil {
				return nil, err
			}
			if ch == nil {
				return nil, nil
			}
			if selectGroup != "" {
				group = selectGroup
			}
			// 继续用解析后的 group 拉全量；若失败至少返回当前渠道
			list, listErr := model.GetSatisfiedChannels(group, modelName, param.RequestPath)
			if listErr != nil {
				return []*model.Channel{ch}, nil
			}
			if len(list) == 0 {
				return []*model.Channel{ch}, nil
			}
			return list, nil
		}
	}

	return model.GetSatisfiedChannels(group, modelName, param.RequestPath)
}

// getPreferredChannelID 读取亲和偏好（如果有）
func getPreferredChannelID(ctx *gin.Context, modelName, group string) int {
	if !common.MemoryCacheEnabled {
		return 0
	}
	id, found := GetPreferredChannelByAffinity(ctx, modelName, group)
	if found {
		return id
	}
	return 0
}

// getAdaptiveUsedChannels 获取本次请求已用过的渠道 ID 列表
func getAdaptiveUsedChannels(c *gin.Context) []int {
	v, ok := c.Get(string(ctxKeyAdaptiveUsedChannels))
	if !ok {
		return nil
	}
	ids, _ := v.([]int)
	return ids
}

// addAdaptiveUsedChannel 记录本次请求使用过的渠道
func addAdaptiveUsedChannel(c *gin.Context, channelID int) {
	existing := getAdaptiveUsedChannels(c)
	existing = append(existing, channelID)
	c.Set(string(ctxKeyAdaptiveUsedChannels), existing)
}

func MarkChannelUsed(c *gin.Context, channelID int) {
	if c == nil || channelID <= 0 || containsInt(getAdaptiveUsedChannels(c), channelID) {
		return
	}
	addAdaptiveUsedChannel(c, channelID)
}

func adaptiveUsedChannelSet(c *gin.Context) map[int]struct{} {
	used := getAdaptiveUsedChannels(c)
	if len(used) == 0 {
		return nil
	}
	excluded := make(map[int]struct{}, len(used))
	for _, channelID := range used {
		excluded[channelID] = struct{}{}
	}
	return excluded
}

// storeAdaptiveSelection 保存本次选择结果到上下文（供失败回写用）
func storeAdaptiveSelection(c *gin.Context, ch *model.Channel, group string, candidates []CandidateScore) {
	c.Set(string(ctxKeyAdaptiveSelected), ch.Id)
	c.Set(string(ctxKeyAdaptiveGroup), group)
	if len(candidates) > 0 {
		c.Set(string(ctxKeyAdaptiveScores), candidates)
	}
}

// logAdaptiveCompare shadow mode 日志
func logAdaptiveCompare(c *gin.Context, modelName, group string, selected *CandidateScore, oldCh *model.Channel) {
	oldID := 0
	if oldCh != nil {
		oldID = oldCh.Id
	}
	// Info so production shadow observation is visible without DEBUG=true.
	logger.LogInfo(c, fmt.Sprintf("[shadow] model=%s group=%s adaptive=#%d(%.3f) orig=#%d",
		modelName, group, selected.Channel.Id, selected.Score, oldID))
}

// RecordAdaptiveResult 单次中继尝试结束后的自适应回写：EWMA 指标 +（真切流）熔断。
//
// 行为：
//   - 自适应或 Shadow 开启时写入 ObserveSuccess/Failure
//   - Shadow：只观测，**立即返回**，不写熔断状态
//   - 真切流：按 ctx 中 CircuitPermit 写成功/失败；4xx 客户端错误按成功释放 half-open，避免探针卡死
//
// 约束：Shadow 期失败不得污染熔断，以便日后真切流从干净状态启动。
// 调用方：controller/relay.go 每次 attempt 后。
func RecordAdaptiveResult(c *gin.Context, channelID int, group, modelName string, statusCode int, latency time.Duration, err error) {
	// Observe when adaptive is on OR shadow-only (shadow needs metrics for A/B).
	if !constant.AdaptiveBalanceEnabled && !constant.AdaptiveBalanceShadowMode {
		return
	}
	if channelID <= 0 {
		return
	}

	succeeded := err == nil && statusCode < 400
	if succeeded {
		ObserveSuccess(channelID, group, modelName, latency)
	} else {
		ObserveFailure(channelID, group, modelName, statusCode, latency)
	}

	// Shadow：只观测，不写熔断（无 permit 回写）。
	if constant.AdaptiveBalanceShadowMode {
		return
	}
	if !constant.AdaptiveBalanceEnabled {
		return
	}
	permitAny, ok := c.Get(string(ctxKeyAdaptiveCircuitPermit))
	permit, permitOK := permitAny.(CircuitPermit)
	if !ok || !permitOK || permit.ChannelID != channelID {
		return
	}
	if succeeded {
		RecordCircuitSuccessWithPermit(permit)
	} else if statusCode >= 500 || statusCode == 429 {
		RecordCircuitFailureWithPermit(permit, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		// A client/input error still proves the upstream is reachable. Do not
		// leave a half-open permit stuck or preserve an old failure streak.
		RecordCircuitSuccessWithPermit(permit)
	}
}

// containsInt 检查 int 切片是否包含某值
func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// randFloat64 生成 [0,1) 随机数
var randFloat64 = func() float64 {
	return rand.Float64()
}

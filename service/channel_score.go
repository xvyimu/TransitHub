package service

import (
	"github.com/xvyimu/TransitHub/constant"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/xvyimu/TransitHub/model"
)

// CandidateScore 单渠道动态评分结果（含因子明细，供日志/观测）。
type CandidateScore struct {
	Channel *model.Channel
	Score   float64

	// 各因子明细（供日志/观测用）
	BaseWeight        float64 `json:"base_weight"`
	SuccessFactor     float64 `json:"success_factor"`
	LatencyFactor     float64 `json:"latency_factor"`
	RateLimitFactor   float64 `json:"rate_limit_factor"`
	ConcurrencyFactor float64 `json:"concurrency_factor"`
	CircuitFactor     float64 `json:"circuit_factor"`
	AffinityFactor    float64 `json:"affinity_factor"`
}

// ScoreCandidates 对候选渠道动态评分并按分数降序返回。
//
// 行为：综合权重、成功率、延迟、429 率、并发、熔断因子、亲和；附带轻微抖动防饿死。
// 约束：open 熔断 circuitFactor=0；half-open=0.5。Shadow 仍会评分（用于对比），但真实路由不采用结果。
// group/model 用于指标分桶；缺失时回退渠道级指标语义由 GetMetrics 决定。
func ScoreCandidates(channels []*model.Channel, group, model string, preferredChannelID int) []CandidateScore {
	if len(channels) == 0 {
		return nil
	}

	candidates := make([]CandidateScore, 0, len(channels))

	for _, ch := range channels {
		circuitFactor := 1.0
		if constant.ChannelCircuitBreakerEnabled {
			// One state read is enough for scoring; open -> 0, half-open -> 0.5.
			state, _, _ := GetCircuitState(ch.Id)
			switch state {
			case CircuitOpen:
				circuitFactor = 0.0
			case CircuitHalfOpen:
				circuitFactor = 0.5
			}
		}

		candidates = append(candidates, scoreCandidate(ch, group, model, preferredChannelID, circuitFactor))
	}

	// 按分数降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates
}

func scoreCandidate(ch *model.Channel, group, model string, preferredChannelID int, circuitFactor float64) CandidateScore {
	metrics := GetMetrics(ch.Id, group, model)
	baseWeight := float64(ch.GetWeight())
	if baseWeight <= 0 {
		baseWeight = 1.0
	}
	successFactor := math.Pow(metrics.SuccessRate, 2)
	latencyMs := float64(metrics.AvgLatency) / float64(time.Millisecond)
	latencyFactor := latencyToScore(latencyMs)
	rateLimitFactor := math.Max(0, 1.0-metrics.RateLimitRate)

	currentConcurrency := GetChannelConcurrency(ch.Id)
	maxConcurrency := int64(constant.MaxChannelConcurrency)
	concurrencyFactor := 1.0
	if maxConcurrency > 0 && currentConcurrency >= maxConcurrency {
		concurrencyFactor = 0.1
	} else if maxConcurrency > 0 {
		concurrencyFactor = 1.0 - float64(currentConcurrency)/float64(maxConcurrency)*0.5
	}

	affinityFactor := 1.0
	if preferredChannelID > 0 && ch.Id == preferredChannelID {
		affinityFactor = 1.5
	}
	score := baseWeight * successFactor * latencyFactor * rateLimitFactor *
		concurrencyFactor * circuitFactor * affinityFactor
	score *= 0.95 + rand.Float64()*0.1

	return CandidateScore{
		Channel:           ch,
		Score:             score,
		BaseWeight:        baseWeight,
		SuccessFactor:     successFactor,
		LatencyFactor:     latencyFactor,
		RateLimitFactor:   rateLimitFactor,
		ConcurrencyFactor: concurrencyFactor,
		CircuitFactor:     circuitFactor,
		AffinityFactor:    affinityFactor,
	}
}

// SelectTopKWeighted 取分数最高的 topK，再按 score 加权随机选一个。
// 约束：score≤0 不参与权重；全为 0 时在 top 内均匀随机。k≤0 时默认 3。
func SelectTopKWeighted(candidates []CandidateScore, k int) *CandidateScore {
	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) == 1 {
		return &candidates[0]
	}

	// 取 topK
	if k <= 0 {
		k = 3
	}
	if k > len(candidates) {
		k = len(candidates)
	}
	top := candidates[:k]

	// 加权随机
	var totalWeight float64
	for _, c := range top {
		if c.Score > 0 {
			totalWeight += c.Score
		}
	}

	if totalWeight <= 0 {
		// 所有分数为 0，均匀随机
		idx := rand.Intn(len(top))
		return &top[idx]
	}

	r := rand.Float64() * totalWeight
	var cumulative float64
	for i, c := range top {
		cumulative += c.Score
		if r < cumulative {
			return &top[i]
		}
	}

	return &top[len(top)-1]
}

// latencyToScore 将延迟（毫秒）映射到 [0, 1] 分数
// 500ms → 1.0, 1s → 0.8, 2s → 0.5, 5s → 0.2, 10s+ → 0.05
func latencyToScore(ms float64) float64 {
	switch {
	case ms <= 0:
		return 1.0
	case ms <= 500:
		return 1.0
	case ms <= 1000:
		return 0.8
	case ms <= 2000:
		return 0.5
	case ms <= 5000:
		return 0.2
	case ms <= 10000:
		return 0.1
	default:
		return 0.05
	}
}

package service

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/model"
)

// In-process ops health metrics (WP-D). Reset on process restart unless noted.

type channelErrorSample struct {
	ChannelID int
	Status    int
	Model     string
	At        time.Time
}

var (
	healthMu sync.Mutex

	relaySuccess atomic.Int64
	relayFail    atomic.Int64

	retryHist   = map[int]int64{}
	errorRing   = make([]channelErrorSample, 0, 256)
	errorByCh   = map[int]int64{}
	shadowAgree atomic.Int64
	shadowTotal atomic.Int64

	refundMetrics = map[string]int64{}
)

// RecordRelayAttempt records a finished relay attempt for ops health.
func RecordRelayAttempt(channelID int, modelName string, retryIndex int, statusCode int, err error) {
	if err == nil && statusCode < 400 {
		relaySuccess.Add(1)
	} else {
		relayFail.Add(1)
		healthMu.Lock()
		errorByCh[channelID]++
		errorRing = append(errorRing, channelErrorSample{
			ChannelID: channelID,
			Status:    statusCode,
			Model:     modelName,
			At:        time.Now(),
		})
		if len(errorRing) > 200 {
			errorRing = errorRing[len(errorRing)-200:]
		}
		healthMu.Unlock()
	}
	healthMu.Lock()
	retryHist[retryIndex]++
	healthMu.Unlock()
}

// RecordShadowCompare records whether adaptive choice matched legacy.
func RecordShadowCompare(agree bool) {
	shadowTotal.Add(1)
	if agree {
		shadowAgree.Add(1)
	}
}

// RecordRefundIntentMetric increments refund intent status counters.
func RecordRefundIntentMetric(status string) {
	if status == "" {
		return
	}
	healthMu.Lock()
	refundMetrics[status]++
	healthMu.Unlock()
}

// HealthSnapshot is returned by the ops health API.
type HealthSnapshot struct {
	GeneratedAt      int64              `json:"generated_at"`
	RelaySuccess     int64              `json:"relay_success"`
	RelayFail        int64              `json:"relay_fail"`
	RetryHistogram   map[string]int64   `json:"retry_histogram"`
	TopErrorChannels []HealthErrorRow   `json:"top_error_channels"`
	Circuits         []HealthCircuitRow `json:"circuits"`
	Shadow           HealthShadow       `json:"shadow"`
	RefundIntents    map[string]int64   `json:"refund_intents"`
	Notes            string             `json:"notes"`
}

type HealthErrorRow struct {
	ChannelID int   `json:"channel_id"`
	Count     int64 `json:"count"`
}

type HealthCircuitRow struct {
	ChannelID          int    `json:"channel_id"`
	State              string `json:"state"`
	ConsecutiveFailure int    `json:"consecutive_failure"`
	OpenUntilUnix      int64  `json:"open_until_unix"`
	LastError          string `json:"last_error"`
}

type HealthShadow struct {
	Samples   int64   `json:"samples"`
	Agree     int64   `json:"agree"`
	AgreeRate float64 `json:"agree_rate"`
}

// SnapshotChannelHealth builds an in-process health snapshot.
func SnapshotChannelHealth() HealthSnapshot {
	healthMu.Lock()
	defer healthMu.Unlock()

	type kv struct {
		id  int
		cnt int64
	}
	list := make([]kv, 0, len(errorByCh))
	for id, cnt := range errorByCh {
		list = append(list, kv{id, cnt})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].cnt > list[j].cnt })
	top := make([]HealthErrorRow, 0, 10)
	for i := 0; i < len(list) && i < 10; i++ {
		top = append(top, HealthErrorRow{ChannelID: list[i].id, Count: list[i].cnt})
	}

	rh := map[string]int64{}
	for k, v := range retryHist {
		rh[strconv.Itoa(k)] = v
	}

	st := shadowTotal.Load()
	sa := shadowAgree.Load()
	rate := 0.0
	if st > 0 {
		rate = float64(sa) / float64(st)
	}

	refundCopy := map[string]int64{}
	for k, v := range refundMetrics {
		refundCopy[k] = v
	}
	if dbCounts, err := model.CountRefundIntentsByStatus(); err == nil {
		for k, v := range dbCounts {
			refundCopy["db_"+k] = v
		}
	}

	return HealthSnapshot{
		GeneratedAt:      time.Now().Unix(),
		RelaySuccess:     relaySuccess.Load(),
		RelayFail:        relayFail.Load(),
		RetryHistogram:   rh,
		TopErrorChannels: top,
		Circuits:         ListCircuitStates(),
		Shadow: HealthShadow{
			Samples:   st,
			Agree:     sa,
			AgreeRate: rate,
		},
		RefundIntents: refundCopy,
		Notes:         "in-process metrics; reset on restart unless db_* refund counts",
	}
}

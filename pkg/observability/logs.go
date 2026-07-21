package observability

import (
	"context"
	"sync"

	"github.com/xvyimu/TransitHub/common"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// OTEL Phase 2 — logs correlation.
//
// This file intentionally adds NO new module dependency: it only reads the
// active OTEL SpanContext (already available via go.opentelemetry.io/otel/trace,
// pulled in by Phase 1 traces.go) so logs can carry the W3C trace_id / span_id
// that Jaeger uses. Full OTLP log push (otel/sdk/log) is deliberately deferred:
// Loki ingestion is done Collector-side (filelog receiver reading the on-disk
// log file), keeping the production binary free of alpha-grade log SDK deps.
//
// Field names are fixed and distinct from the AxonHub business "trace_id"
// (middleware/trace.go) so the two never collide in a log line:
//   - otel_trace_id : W3C 128-bit trace id (hex) — matches Jaeger
//   - otel_span_id  : W3C 64-bit span id (hex)
//
// Gate: OTEL_LOGS_ENABLED (default false). Even when enabled, fields are omitted
// unless a valid sampled span is in ctx (which only exists when traces are also
// enabled and the request was sampled). When either gate is off, callers omit
// the fields entirely (no empty noise in logs).

var (
	logsEnabledOnce sync.Once
	logsEnabled     bool
)

// EnabledLogs reports whether OTEL log correlation fields are opted in.
// Independent of EnabledTraces so operators can run traces without stamping
// every log line (and vice versa: logs on with traces off is a no-op).
func EnabledLogs() bool {
	logsEnabledOnce.Do(func() {
		logsEnabled = common.GetEnvOrDefaultBool("OTEL_LOGS_ENABLED", false)
	})
	return logsEnabled
}

// resetEnabledLogsForTest clears the cached EnabledLogs result (tests only).
func resetEnabledLogsForTest() {
	logsEnabledOnce = sync.Once{}
	logsEnabled = false
}

// TraceIDsFromContext returns the active OTEL (traceID, spanID) as lowercase hex
// strings, or ("", "") when logs correlation is disabled or there is no valid
// sampled span in ctx.
//
// Gated by EnabledLogs(): when logs are opt-out (default), this is a cheap
// no-op so logging hot paths pay nothing.
func TraceIDsFromContext(ctx context.Context) (traceID string, spanID string) {
	if ctx == nil || !EnabledLogs() {
		return "", ""
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return "", ""
	}
	return sc.TraceID().String(), sc.SpanID().String()
}

// TraceIDsFromGin is a gin.Context convenience wrapper around
// TraceIDsFromContext using the request context.
func TraceIDsFromGin(c *gin.Context) (traceID string, spanID string) {
	if c == nil || c.Request == nil {
		return "", ""
	}
	return TraceIDsFromContext(c.Request.Context())
}

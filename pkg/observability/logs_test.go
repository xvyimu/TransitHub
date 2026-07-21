package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)
	hex16 = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

// When OTEL_LOGS_ENABLED is opt-out (default), the helper is a cheap no-op
// returning empties so the logging hot path pays nothing.
func TestTraceIDsFromContextDisabledReturnsEmpty(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "")
	resetEnabledLogsForTest()

	tid, sid := TraceIDsFromContext(context.Background())
	require.Empty(t, tid)
	require.Empty(t, sid)
}

// Explicit false also yields empties.
func TestTraceIDsFromContextLogsFalseReturnsEmpty(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "false")
	resetEnabledLogsForTest()
	require.False(t, EnabledLogs())

	tid, sid := TraceIDsFromContext(context.Background())
	require.Empty(t, tid)
	require.Empty(t, sid)
}

// A nil context must never panic and yields empty ids.
func TestTraceIDsFromContextNilContext(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	resetEnabledLogsForTest()

	require.NotPanics(t, func() {
		tid, sid := TraceIDsFromContext(nil)
		require.Empty(t, tid)
		require.Empty(t, sid)
	})
}

// Enabled but no span in context (unsampled / no active span) yields empties,
// so callers omit the fields entirely rather than emitting all-zero ids.
func TestTraceIDsFromContextEnabledNoSpan(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	resetEnabledLogsForTest()

	tid, sid := TraceIDsFromContext(context.Background())
	require.Empty(t, tid)
	require.Empty(t, sid)
}

// With a valid sampled span installed, the helper returns the same W3C
// trace_id / span_id hex that Jaeger uses — this is the correlation contract.
func TestTraceIDsFromContextEnabledWithSpan(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	resetEnabledLogsForTest()

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "unit")
	defer span.End()

	tid, sid := TraceIDsFromContext(ctx)
	require.Regexp(t, hex32, tid)
	require.Regexp(t, hex16, sid)

	// Must match exactly what the span reports (what Jaeger will index).
	sc := trace.SpanContextFromContext(ctx)
	require.Equal(t, sc.TraceID().String(), tid)
	require.Equal(t, sc.SpanID().String(), sid)
}

// Logs enabled but traces-off still works IF a span is manually put in ctx
// (e.g. unit tests); production without traces has no span so fields stay empty.
func TestTraceIDsFromContextLogsOnTracesOffNoSpan(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	t.Setenv("OTEL_TRACES_ENABLED", "false")
	resetEnabledLogsForTest()
	resetEnabledTracesForTest()

	tid, sid := TraceIDsFromContext(context.Background())
	require.Empty(t, tid)
	require.Empty(t, sid)
}

// TraceIDsFromGin returns empties (no panic) when the gin.Context has no request.
func TestTraceIDsFromGinNilRequest(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	resetEnabledLogsForTest()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	require.NotPanics(t, func() {
		tid, sid := TraceIDsFromGin(c)
		require.Empty(t, tid)
		require.Empty(t, sid)
	})
}

// End-to-end through the HTTP middleware: when traces+logs are enabled the
// request context carries a valid span whose ids the helper can read inside a handler.
func TestTraceIDsFromGinWithSpan(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")
	t.Setenv("OTEL_TRACES_ENABLED", "true")
	resetEnabledLogsForTest()
	resetEnabledTracesForTest()

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	InitTracesForTest(tp)
	t.Cleanup(func() { _ = ShutdownTraces(context.Background()) })

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(HTTPTraceMiddleware())

	var gotTID, gotSID string
	engine.GET("/ping", func(c *gin.Context) {
		gotTID, gotSID = TraceIDsFromGin(c)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Regexp(t, hex32, gotTID)
	require.Regexp(t, hex16, gotSID)
}

func TestEnabledLogsDefaultOff(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "")
	resetEnabledLogsForTest()
	require.False(t, EnabledLogs())
}

package logger

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// captureDefaultWriter swaps gin.DefaultWriter/ErrorWriter for a buffer for the
// duration of fn, restoring them after. logHelper writes through these, guarded
// by common.LogWriterMu.
func captureDefaultWriter(fn func()) string {
	var buf bytes.Buffer
	oldOut, oldErr := gin.DefaultWriter, gin.DefaultErrorWriter
	gin.DefaultWriter = io.Writer(&buf)
	gin.DefaultErrorWriter = io.Writer(&buf)
	defer func() {
		gin.DefaultWriter = oldOut
		gin.DefaultErrorWriter = oldErr
	}()
	fn()
	return buf.String()
}

// TestLoggerOTELCorrelation asserts logs gain otel_trace_id/otel_span_id fields
// only when OTEL_LOGS_ENABLED=true AND a valid span is in context.
//
// The env is set before the first EnabledLogs() call in this package's test
// binary (sync.Once cache).
func TestLoggerOTELCorrelation(t *testing.T) {
	t.Setenv("OTEL_LOGS_ENABLED", "true")

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "unit")
	defer span.End()

	// With a valid sampled span, the log line must carry the correlation fields.
	out := captureDefaultWriter(func() { LogInfo(ctx, "hello with span") })
	require.Contains(t, out, "hello with span")
	require.Contains(t, out, "otel_trace_id=")
	require.Contains(t, out, "otel_span_id=")

	// A background context (no span) must NOT add empty/zero fields — the line
	// stays exactly as before, so existing log parsers are unaffected.
	out = captureDefaultWriter(func() { LogInfo(context.Background(), "hello no span") })
	require.Contains(t, out, "hello no span")
	require.NotContains(t, out, "otel_trace_id=")
	require.NotContains(t, out, "otel_span_id=")

	// The emitted id must be the real hex from the span (what Jaeger indexes),
	// never the all-zero id.
	out = captureDefaultWriter(func() { LogInfo(ctx, "check id value") })
	require.NotContains(t, out, "otel_trace_id=00000000000000000000000000000000")
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "check id value") {
			require.Contains(t, line, span.SpanContext().TraceID().String())
		}
	}
}

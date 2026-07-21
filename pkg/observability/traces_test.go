package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xvyimu/TransitHub/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestEnabledTracesDefaultOff(t *testing.T) {
	t.Setenv("OTEL_TRACES_ENABLED", "")
	resetEnabledTracesForTest()
	require.False(t, EnabledTraces())
}

func TestInitTracesDisabledNoExporter(t *testing.T) {
	t.Setenv("OTEL_TRACES_ENABLED", "false")
	resetEnabledTracesForTest()

	shutdown, err := InitTraces(context.Background())
	require.NoError(t, err)
	require.Nil(t, shutdown)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(HTTPTraceMiddleware())
	engine.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	require.NotPanics(t, func() {
		engine.ServeHTTP(rec, req)
	})
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestInitTracesEnabledBadEndpointNoPanic(t *testing.T) {
	t.Setenv("OTEL_TRACES_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "127.0.0.1:1")
	t.Setenv("OTEL_TRACES_INSECURE", "true")
	t.Setenv("OTEL_TRACES_SAMPLER_RATIO", "1.0")
	resetEnabledTracesForTest()

	var shutdown func(context.Context) error
	require.NotPanics(t, func() {
		var err error
		shutdown, err = InitTraces(context.Background())
		// fail-open: may return err or succeed with lazy dial
		_ = err
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(HTTPTraceMiddleware())
	engine.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	require.NotPanics(t, func() {
		engine.ServeHTTP(rec, req)
	})
	require.Equal(t, http.StatusOK, rec.Code)

	if shutdown != nil {
		require.NotPanics(t, func() {
			_ = shutdown(context.Background())
		})
	}
	// Ensure double shutdown is safe
	require.NotPanics(t, func() {
		_ = ShutdownTraces(context.Background())
	})
}

func TestHTTPTraceMiddlewareSetsSpanAttributes(t *testing.T) {
	t.Setenv("OTEL_TRACES_ENABLED", "true")
	resetEnabledTracesForTest()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	InitTracesForTest(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		resetEnabledTracesForTest()
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(common.RequestIdKey, "req-test-1")
		c.Set("trace_id", "axon-trace-1")
		c.Set("thread_id", "axon-thread-1")
		c.Next()
	})
	engine.Use(HTTPTraceMiddleware())
	engine.GET("/v1/models", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// Force flush
	require.NoError(t, tp.ForceFlush(context.Background()))
	spans := exporter.GetSpans()
	require.NotEmpty(t, spans)

	var found *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name == "HTTP GET" {
			found = &spans[i]
			break
		}
	}
	require.NotNil(t, found, "expected HTTP GET span")

	attrs := attrMap(found.Attributes)
	require.Equal(t, "GET", attrs["http.request.method"])
	require.Equal(t, "/v1/models", attrs["http.route"])
	require.Equal(t, int64(200), attrs["http.response.status_code"])
	require.Equal(t, "req-test-1", attrs["request_id"])
	require.Equal(t, "axon-trace-1", attrs["axon.trace_id"])
	require.Equal(t, "axon-thread-1", attrs["axon.thread_id"])
	// Forbidden attributes must not appear
	_, hasAuth := attrs["Authorization"]
	require.False(t, hasAuth)
	_, hasBody := attrs["http.request.body"]
	require.False(t, hasBody)
}

func TestSamplerRatioClamp(t *testing.T) {
	require.InDelta(t, 0.1, parseSamplerRatio("", 0.1), 1e-9)
	require.InDelta(t, 0.5, parseSamplerRatio("0.5", 0.1), 1e-9)
	require.InDelta(t, 1.0, parseSamplerRatio("1.0", 0.1), 1e-9)
	require.InDelta(t, 0.0, parseSamplerRatio("0", 0.1), 1e-9)
	require.InDelta(t, 0.1, parseSamplerRatio("2", 0.1), 1e-9)
	require.InDelta(t, 0.1, parseSamplerRatio("-0.1", 0.1), 1e-9)
	require.InDelta(t, 0.1, parseSamplerRatio("not-a-number", 0.1), 1e-9)
}

func TestNormalizeOTLPEndpoint(t *testing.T) {
	require.Equal(t, "localhost:4317", normalizeOTLPEndpoint(""))
	require.Equal(t, "collector:4317", normalizeOTLPEndpoint("collector:4317"))
	require.Equal(t, "localhost:4317", normalizeOTLPEndpoint("http://localhost:4317"))
	require.Equal(t, "localhost:4317", normalizeOTLPEndpoint("https://localhost:4317"))
}

func attrMap(attrs []attribute.KeyValue) map[string]any {
	out := make(map[string]any, len(attrs))
	for _, a := range attrs {
		key := string(a.Key)
		switch a.Value.Type() {
		case attribute.STRING:
			out[key] = a.Value.AsString()
		case attribute.INT64:
			out[key] = a.Value.AsInt64()
		case attribute.BOOL:
			out[key] = a.Value.AsBool()
		case attribute.FLOAT64:
			out[key] = a.Value.AsFloat64()
		default:
			out[key] = a.Value.Emit()
		}
	}
	return out
}

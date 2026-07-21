package observability

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/xvyimu/TransitHub/common"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName           = "github.com/xvyimu/TransitHub"
	defaultServiceName   = "new-api"
	defaultOTLPEndpoint  = "localhost:4317"
	defaultSamplerRatio  = 0.1
	spanNameHTTPPrefix   = "HTTP "
	spanNameRelayAttempt = "relay.attempt"
	spanNameRelayUpstream = "relay.upstream"
)

var (
	tracesEnabledOnce sync.Once
	tracesEnabled     bool

	tracesMu       sync.Mutex
	tracerProvider *sdktrace.TracerProvider
	shutdownOnce   sync.Once
)

// EnabledTraces reports whether OTEL traces are opted in.
func EnabledTraces() bool {
	tracesEnabledOnce.Do(func() {
		tracesEnabled = common.GetEnvOrDefaultBool("OTEL_TRACES_ENABLED", false)
	})
	return tracesEnabled
}

// resetEnabledTracesForTest clears the cached EnabledTraces result (tests only).
func resetEnabledTracesForTest() {
	tracesEnabledOnce = sync.Once{}
	tracesEnabled = false
}

// parseSamplerRatio parses a head-sampling ratio in [0,1]. Invalid values fall back
// to defaultRatio and log once via common.SysError.
func parseSamplerRatio(raw string, defaultRatio float64) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultRatio
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || f < 0 || f > 1 {
		common.SysError(fmt.Sprintf("invalid OTEL_TRACES_SAMPLER_RATIO %q, using default %.1f", raw, defaultRatio))
		return defaultRatio
	}
	return f
}

// normalizeOTLPEndpoint strips http(s):// schemes for gRPC dial targets.
func normalizeOTLPEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultOTLPEndpoint
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "https://") {
		raw = raw[len("https://"):]
	} else if strings.HasPrefix(lower, "http://") {
		raw = raw[len("http://"):]
	}
	// Drop trailing path if accidentally included after host:port
	if i := strings.Index(raw, "/"); i >= 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultOTLPEndpoint
	}
	return raw
}

// InitTraces installs global TracerProvider when enabled.
// When disabled: no-op, returns (nil, nil).
// When enabled but exporter setup fails: log error, leave noop provider, return (nil, err)
// so main can continue serving (fail-open). Never panic.
func InitTraces(ctx context.Context) (func(context.Context) error, error) {
	if !EnabledTraces() {
		return nil, nil
	}

	endpoint := normalizeOTLPEndpoint(common.GetEnvOrDefaultString("OTEL_EXPORTER_OTLP_ENDPOINT", defaultOTLPEndpoint))
	serviceName := common.GetEnvOrDefaultString("OTEL_SERVICE_NAME", defaultServiceName)
	ratio := parseSamplerRatio(os.Getenv("OTEL_TRACES_SAMPLER_RATIO"), defaultSamplerRatio)
	insecureDial := common.GetEnvOrDefaultBool("OTEL_TRACES_INSECURE", true)

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if insecureDial {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		common.SysError(fmt.Sprintf("otel traces exporter init failed (fail-open): %v", err))
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(common.Version),
		),
	)
	if err != nil {
		_ = exporter.Shutdown(ctx)
		common.SysError(fmt.Sprintf("otel traces resource init failed (fail-open): %v", err))
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if err != nil {
			common.SysError(fmt.Sprintf("otel error: %v", err))
		}
	}))

	tracesMu.Lock()
	tracerProvider = tp
	shutdownOnce = sync.Once{}
	tracesMu.Unlock()

	common.SysLog(fmt.Sprintf("otel traces enabled: endpoint=%s service=%s sampler_ratio=%.3f insecure=%v",
		endpoint, serviceName, ratio, insecureDial))

	return tp.Shutdown, nil
}

// InitTracesForTest installs a custom TracerProvider for unit tests and marks traces enabled.
// Caller must restore via ShutdownTraces / otel.SetTracerProvider when done.
func InitTracesForTest(tp *sdktrace.TracerProvider) {
	tracesEnabledOnce.Do(func() {})
	tracesEnabled = true
	if tp != nil {
		otel.SetTracerProvider(tp)
		tracesMu.Lock()
		tracerProvider = tp
		shutdownOnce = sync.Once{}
		tracesMu.Unlock()
	}
}

// ShutdownTraces flushes and shuts down the provider if InitTraces installed one.
// Safe if never initialized or already shut down.
func ShutdownTraces(ctx context.Context) error {
	var err error
	shutdownOnce.Do(func() {
		tracesMu.Lock()
		tp := tracerProvider
		tracerProvider = nil
		tracesMu.Unlock()
		if tp == nil {
			return
		}
		err = tp.Shutdown(ctx)
		if err != nil {
			common.SysError(fmt.Sprintf("otel traces shutdown: %v", err))
		}
	})
	return err
}

// Tracer returns a named tracer; always safe (uses global otel.GetTracerProvider()).
func Tracer(name string) trace.Tracer {
	if name == "" {
		name = tracerName
	}
	return otel.Tracer(name)
}

// StartSpan starts a span from ctx; returns (ctx, span). Safe with noop provider.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer(tracerName).Start(ctx, name, opts...)
}

// SpanFromGin returns context for span ops from gin.Context request context.
func SpanFromGin(c *gin.Context) context.Context {
	if c == nil || c.Request == nil {
		return context.Background()
	}
	return c.Request.Context()
}

// HTTPTraceMiddleware returns Gin middleware that creates a server span for the request.
func HTTPTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !EnabledTraces() {
			c.Next()
			return
		}

		method := c.Request.Method
		spanName := spanNameHTTPPrefix + method
		ctx, span := StartSpan(c.Request.Context(), spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.request.method", method),
				attribute.String("url.path", c.Request.URL.Path),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := c.Writer.Status()
		attrs := []attribute.KeyValue{
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
		}
		if rid := c.GetString(common.RequestIdKey); rid != "" {
			attrs = append(attrs, attribute.String("request_id", rid))
		}
		if tid := c.GetString("trace_id"); tid != "" {
			attrs = append(attrs, attribute.String("axon.trace_id", tid))
		}
		if thid := c.GetString("thread_id"); thid != "" {
			attrs = append(attrs, attribute.String("axon.thread_id", thid))
		}
		span.SetAttributes(attrs...)

		if status >= 500 {
			span.SetStatus(codes.Error, httpStatusText(status))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

func httpStatusText(code int) string {
	// Generic status class only — no response body.
	switch {
	case code >= 500:
		return "server_error"
	case code >= 400:
		return "client_error"
	default:
		return ""
	}
}

// SpanNameRelayAttempt is the child span name for relay retry attempts.
func SpanNameRelayAttempt() string { return spanNameRelayAttempt }

// SpanNameRelayUpstream is the child span name for upstream HTTP doRequest.
func SpanNameRelayUpstream() string { return spanNameRelayUpstream }

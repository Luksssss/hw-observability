package fibonacci

import (
	"context"
	"time"

	"git.ozon.dev/3.10-observability/hw6-grpc-fib/pkg/fibproto"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// МЕТРИКИ
var (
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ozon_school",
		Subsystem: "grpc",
		Name:      "requests_total",
	}, []string{"method"})

	requestDurationSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "ozon_school",
		Subsystem:  "grpc",
		Name:       "requests_duration_seconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		MaxAge:     30 * time.Second,
	}, []string{"method"})

	requestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ozon",
		Subsystem: "http",
		Name:      "requests_duration_histogram_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.00001, 2, 64),
	}, []string{"method"})

	inflightQueries = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "ozon_school",
		Subsystem: "grpc",
		Name:      "requests_inflight",
	})
)

func getTrace(ctx context.Context, m string) (opentracing.Span, context.Context) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	spanctx, _ := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(md),
	)

	span, ctx := opentracing.StartSpanFromContext(ctx, m, ext.RPCServerOption(spanctx))

	if spanctx, ok := span.Context().(jaeger.SpanContext); ok {
		md := metadata.Pairs("x-trace-id", spanctx.TraceID().String())
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return span, ctx
}

// функция-прослойка на все методы
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	var (
		h   interface{}
		err error
	)
	conn := info.Server.(*Conn)
	method := info.FullMethod

	// Tracing
	span, ctx := getTrace(ctx, method)
	defer span.Finish()

	// разбивка по методу
	// requestCounter.Inc()
	requestCounter.WithLabelValues(method).Inc()

	traceID := getTraceID(ctx)

	start := time.Now()
	switch method {
	case "/Fibonacci/Fib":
		r := req.(*fibproto.NumReq)
		h, err = middlewareFib(ctx, r, handler, conn, method, traceID)
	case "/Fibonacci/Sqr":
		r := req.(*fibproto.NumReq)
		h, err = middlewareSqr(ctx, r, handler, conn, method, traceID)
	default:
		conn.Logger.Error("RPC method not found", zap.String("method", method))
	}
	total := time.Since(start)
	requestDurationSummary.WithLabelValues(method).Observe(float64(total) / float64(time.Second))
	requestDurationHistogram.WithLabelValues(method).Observe(float64(total) / float64(time.Second))

	return h, err

}

func middlewareFib(ctx context.Context, req *fibproto.NumReq, handler grpc.UnaryHandler, conn *Conn, method string, traceID string) (interface{}, error) {

	// запрос
	conn.Logger.Debug("request", zap.String("method", method), zap.Int64("n", req.N), zap.String("x-trace-id", traceID))
	if req.N < 0 {
		conn.Logger.Error("negative number", zap.Int64("n", req.N))
	}

	h, err := handler(ctx, req)
	if err != nil {
		conn.Logger.Error("RPC Fib failed with error", zap.Error(err))
		return h, err
	}

	// ответ
	resp := h.(*fibproto.NumResp)
	conn.Logger.Debug("response", zap.String("method", method), zap.Int64("nFib", resp.NFib), zap.String("x-trace-id", traceID))

	return h, err

}

func middlewareSqr(ctx context.Context, req *fibproto.NumReq, handler grpc.UnaryHandler, conn *Conn, method string, traceID string) (interface{}, error) {

	// запрос
	conn.Logger.Debug("request", zap.String("method", method), zap.Int64("n", req.N), zap.String("x-trace-id", traceID))

	h, err := handler(ctx, req)
	if err != nil {
		conn.Logger.Error("RPC Fib failed with error", zap.Error(err))
		return h, err
	}

	// ответ
	resp := h.(*fibproto.NumRespQ)
	conn.Logger.Debug("response", zap.String("method", method), zap.Int64("nSqr", resp.NSqr), zap.String("x-trace-id", traceID))

	return h, err

}

func getTraceID(ctx context.Context) string {

	var traceID string

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		traceID = md["x-trace-id"][0]
	}

	return traceID

}

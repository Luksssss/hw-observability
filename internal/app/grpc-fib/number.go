package fibonacci

import (
	"context"

	"git.ozon.dev/3.10-observability/hw6-grpc-fib/pkg/fibproto"
	"github.com/opentracing/opentracing-go"
)

// Fib grpc
func (c *Conn) Fib(ctx context.Context, req *fibproto.NumReq) (*fibproto.NumResp, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "funcFib")
	span.SetTag("n", req.N)
	defer span.Finish()

	if req.N < 0 {
		return &fibproto.NumResp{
			NFib: -1,
		}, nil
	}

	return &fibproto.NumResp{
		NFib: Get(ctx, req.N),
	}, nil
}

// Get мат. логика по Фибоначчи
func Get(ctx context.Context, n int64) int64 {

	span, ctx := opentracing.StartSpanFromContext(ctx, "funcGet")
	span.SetTag("n", n)
	defer span.Finish()

	switch {
	case n == 0:
		return 0
	case n == 1:
		return 1
	default:
		return Get(ctx, n-1) + Get(ctx, n-2)
	}
}

// Sqr grpc
func (c *Conn) Sqr(ctx context.Context, req *fibproto.NumReq) (*fibproto.NumRespQ, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "funcSqr")
	span.SetTag("n", req.N)
	defer span.Finish()

	return &fibproto.NumRespQ{
		NSqr: GetQ(ctx, req.N),
	}, nil
}

// GetQ мат. логика квадрата
func GetQ(ctx context.Context, n int64) int64 {

	span, ctx := opentracing.StartSpanFromContext(ctx, "funcGetQ")
	span.SetTag("n", n)
	defer span.Finish()

	return n * n
}

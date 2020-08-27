package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"git.ozon.dev/3.10-observability/hw6-grpc-fib/pkg/fibproto"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	envConf "git.ozon.dev/3.10-observability/hw6-grpc-fib/internal/app/config"
)

var (
	n = flag.Int64("n", 5, "fibonacci number")
)

func main() {
	flag.Parse()

	globCfg := envConf.Config{}
	err := envConf.ReadConfig(&globCfg)
	if err != nil {
		log.Println(err)
	}

	tcfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  globCfg.TrConfType,
			Param: globCfg.TrConfParam,
		},
	}

	closer, err := tcfg.InitGlobalTracer("client-grpc")
	if err != nil {
		log.Println(err)
	}
	defer closer.Close()

	span, ctx := opentracing.StartSpanFromContext(context.Background(), "Client-grpc")
	defer span.Finish()

	// связываем в единый стэктрейс
	md := make(metadata.MD)
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(md),
	)

	for k, v := range md {
		ctx = metadata.AppendToOutgoingContext(ctx, k, v[0])
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", globCfg.Site, globCfg.Port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := fibproto.NewFibonacciClient(conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	req := fibproto.NumReq{N: *n}
	c.Fib(ctx, &req)
}

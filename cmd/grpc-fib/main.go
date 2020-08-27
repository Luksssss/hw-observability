package main

import (
	"log"

	fibserver "git.ozon.dev/3.10-observability/hw6-grpc-fib/internal/app/grpc-fib"
	"go.uber.org/zap"
)

func main() {

	conn, err := fibserver.New()
	if err != nil {
		log.Fatal(err)
	}

	err = conn.RunServer()
	if err != nil {
		conn.Logger.Fatal("failed to start server", zap.Error(err))
	}

	defer conn.Closer.Close()

}

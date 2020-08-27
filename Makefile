build:
	mkdir -p bin
	go build -o -mod=vendor ./bin/grpc-fib ./cmd/grpc-fib/

run:
	go run ./cmd/grpc-fib/ --port=8080

run2:
	go run ./cmd/grpc-fib/ --port=8090 --log.lvl=info

run-client:
	go run ./cmd/grpc-fib/ --port=8085 --port.prom=8086 2>&1 | tee ../logs/logs/file.log

run-log:
	go run ./cmd/grpc-fib/ 2>&1 | tee ./logs/logs/file.log

test:
	go test ./...

testv:
	go test -v ./...

gen-rpc: 
	protoc -I ./api/ --go_out=plugins=grpc:./pkg/ ./api/fib.proto

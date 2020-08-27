package fibonacci

import (
	"fmt"
	"io"
	"net"
	"net/http"

	envConf "git.ozon.dev/3.10-observability/hw6-grpc-fib/internal/app/config"
	"git.ozon.dev/3.10-observability/hw6-grpc-fib/pkg/fibproto"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/grpc"
)

// Conn структура для коннекта
type Conn struct {
	App      *grpc.Server
	Logger   *zap.Logger
	metrics  *http.Server
	Closer   io.Closer
	globConf *envConf.Config
}

// New инициализируем стартовые параметры (коннект, логи).
func New() (*Conn, error) {

	conn := Conn{}
	globCfg := envConf.Config{}

	err := envConf.ReadConfig(&globCfg)
	if err != nil {
		return &conn, err
	}
	conn.globConf = &globCfg

	logger, err := conn.createLogger()
	if err != nil {
		return &conn, err
	}
	conn.Logger = logger

	closer, err := conn.createTrace()
	if err != nil {
		return &conn, err
	}
	conn.Closer = closer

	conn.App = grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
	)

	conn.metrics = conn.listenMetrics()

	return &conn, nil

}

func (c *Conn) createTrace() (io.Closer, error) {

	tcfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  c.globConf.TrConfType,
			Param: c.globConf.TrConfParam,
		},
	}

	closer, err := tcfg.InitGlobalTracer("fib")

	return closer, err
}

func (c *Conn) createLogger() (*zap.Logger, error) {

	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// cfg.DisableStacktrace = true
	if c.globConf.Development {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := cfg.Build()

	if err != nil {
		return logger, err
	}

	return logger, nil

}

// listenMetrics пишем метрики.
func (c *Conn) listenMetrics() *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", c.globConf.PortProm),
		Handler: routing(),
	}
}

// RunServer запуск сервера.
func (c *Conn) RunServer() error {

	fibproto.RegisterFibonacciServer(c.App, c)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", c.globConf.Port))
	if err != nil {
		return err
	}

	ch := make(chan error)

	go func() {
		ch <- c.metrics.ListenAndServe()
	}()

	c.Logger.Debug("Listening metrics...", zap.Int("port metrics", c.globConf.PortProm))

	go func() {
		ch <- c.App.Serve(listener)
	}()

	c.Logger.Debug("Listening grpc...", zap.Int("port grpc", c.globConf.Port))

	// можно заменить на ErrGroup()
	return <-ch

}

func routing() *mux.Router {
	// InstrumentHandlerInFlight - обвязка для inflightQueries
	mux := mux.NewRouter()
	mux.Handle("/metrics", promhttp.InstrumentHandlerInFlight(inflightQueries, promhttp.Handler())).Methods("GET")

	return mux
}

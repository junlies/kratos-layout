package server

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos-layout/internal/service"
	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, hs []HttpService, logger log.Logger, meter metric.Meter, tp trace.TracerProvider) *http.Server {
	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		panic(err)
	}
	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		panic(err)
	}

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.ProtoValidate(),
			tracing.Server(tracing.WithTracerProvider(tp)),
			logging.Server(logger),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(seconds)),
			metadata.Server(),
		),
	}
	c := bc.GetServer()
	if c.Http.GetNetwork() != "" {
		opts = append(opts, http.Network(c.Http.GetNetwork()))
	}
	if c.Http.GetAddr() != "" {
		opts = append(opts, http.Address(c.Http.GetAddr()))
	}
	if c.Http.GetTimeout() != nil {
		opts = append(opts, http.Timeout(c.Http.GetTimeout().AsDuration()))
	}
	srv := http.NewServer(opts...)
	srv.HandlePrefix("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	for _, h := range hs {
		h.RegisterHttpServer(srv)
	}
	return srv
}

type HttpService interface {
	RegisterHttpServer(*http.Server)
}

func NewHTTPServiceSet(greeter *service.GreeterService) []HttpService {
	return []HttpService{
		greeter,
	}
}

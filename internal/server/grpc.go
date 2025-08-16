package server

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos-layout/internal/service"
	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *conf.Bootstrap, gs []GrpcService, logger log.Logger, meter metric.Meter, tp trace.TracerProvider) *grpc.Server {
	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		panic(err)
	}
	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		panic(err)
	}
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(tracing.WithTracerProvider(tp)),
			validate.ProtoValidate(),
			logging.Server(logger),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(seconds)),
			metadata.Server(),
		),
	}
	s := bc.GetServer()
	if s.Grpc.Network != "" {
		opts = append(opts, grpc.Network(s.Grpc.Network))
	}
	if s.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(s.Grpc.Addr))
	}
	if s.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(s.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	for _, g := range gs {
		g.RegisterServer(srv)
	}
	return srv
}

type GrpcService interface {
	RegisterServer(*grpc.Server)
}

func NewGRPCServiceSet(greeter *service.GreeterService) []GrpcService {
	return []GrpcService{
		greeter,
	}
}

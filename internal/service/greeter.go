package service

import (
	"context"
	"time"

	v1 "github.com/go-kratos/kratos-layout/api/helloworld/v1"
	"github.com/go-kratos/kratos-layout/internal/biz"
	"github.com/go-kratos/kratos-layout/internal/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

func (s *GreeterService) RegisterServer(srv *grpc.Server) {
	v1.RegisterGreeterServer(srv, s)
}

func (s *GreeterService) RegisterHttpServer(srv *http.Server) {
	v1.RegisterGreeterHTTPServer(srv, s)
}

func (s *GreeterService) RegisterLimiter() []*middleware.LimiterConfig {
	return []*middleware.LimiterConfig{
		{
			O: v1.OperationGreeterSayHello,
			R: 2000,
			B: 2000,
			T: time.Second * 3,
		},
	}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}

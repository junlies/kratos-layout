package server

import (
	"context"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
)

type PprofServer struct {
	started uint32

	conf *conf.Server_Pprof

	listener net.Listener
	srv      *http.Server

	log *log.Helper
}

func NewPprof(bc *conf.Bootstrap, logger log.Logger) (*PprofServer, error) {
	s := bc.Server
	pprof := &PprofServer{
		conf: s.GetPprof(),
		log:  log.NewHelper(logger, log.WithMessageKey("pprof")),
	}
	addr := s.GetPprof().GetAddr()
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		pprof.listener = listener
		pprof.srv = &http.Server{}
	}
	return pprof, nil
}

func (s *PprofServer) Start(ctx context.Context) error {
	if s.listener == nil {
		return nil
	}
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return nil
	}
	defer atomic.CompareAndSwapUint32(&s.started, 1, 0)
	s.log.WithContext(ctx).Infof("[Pprof] server listening at %s", s.listener.Addr().String())
	err := s.srv.Serve(s.listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *PprofServer) Stop(ctx context.Context) error {
	s.log.WithContext(ctx).Info("[Pprof] server stopped")
	if atomic.LoadUint32(&s.started) == 1 {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

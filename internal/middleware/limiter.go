package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/aegis/ratelimit"
	"github.com/go-kratos/aegis/ratelimit/bbr"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/time/rate"
)

type (
	// LimiterConfig 限速器配置
	//
	// O 接口的唯一标识，常使用Protobuf自带的Operation
	//
	// R 接口限速 r/s(每秒处理r次请求)
	//
	// B 并发限制 并发处理b个请求
	//
	// T 请求等待时间 未处理的请求会进入等待，等待时间为t
	LimiterConfig struct {
		O string
		R rate.Limit
		B int
		T time.Duration
	}

	// limiter 业务limiter
	limiter struct {
		limiter *rate.Limiter
		t       time.Duration
	}

	// Option limiter middleware option
	Option  func(*options)
	options struct {
		bbrConfig *conf.BBR
	}
)

var (
	// ErrLimitExceed is service unavailable due to rate limit exceeded.
	ErrLimitExceed = errors.New(http.StatusTooManyRequests, "RATE_LIMIT", "service unavailable due to rate limit exceeded")
	// ErrLimitUnknown is service unavailable due to unknown rate limit
	ErrLimitUnknown = errors.New(http.StatusInternalServerError, "RATE_LIMIT", "service unavailable due to unknown rate limit")
)

var (
	// 业务limiter
	limiters = sync.Map{}
	// bbr
	globalLimiter ratelimit.Limiter
	windowSize    = time.Second * 10
	bucketNum     = 100
	cpuThreshold  = int64(800)
)

func WithBBR(bbrConfig *conf.BBR) Option {
	return func(o *options) {
		o.bbrConfig = bbrConfig
	}
}

func Limiter(ls []*LimiterConfig, opts ...Option) middleware.Middleware {
	op := options{
		bbrConfig: nil,
	}
	for _, o := range opts {
		o(&op)
	}

	// bbr
	if op.bbrConfig != nil && globalLimiter == nil {
		windowSize = op.bbrConfig.GetWindowSize().AsDuration()
		bucketNum = int(op.bbrConfig.GetBucket())
		cpuThreshold = op.bbrConfig.GetCpuThreshold()

		bbrOptions := []bbr.Option{
			bbr.WithWindow(windowSize),
			bbr.WithBucket(bucketNum),
			bbr.WithCPUThreshold(cpuThreshold),
		}
		globalLimiter = bbr.NewLimiter(bbrOptions...)
	}

	// 业务limiter
	for _, l := range ls {
		limiters.Store(l.O, &limiter{
			limiter: rate.NewLimiter(l.R, l.B),
			t:       l.T,
		})
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// bbr
			if globalLimiter != nil {
				done, e := globalLimiter.Allow()
				if e != nil {
					// rejected
					return nil, ErrLimitExceed
				}
				defer func() {
					done(ratelimit.DoneInfo{Err: err})
				}()
			}

			// 业务limiter
			var o string
			if info, ok := transport.FromServerContext(ctx); ok {
				o = info.Operation()
			}

			// 获取限速器
			v, ok := limiters.Load(o)
			if !ok || v == nil {
				return nil, ErrLimitUnknown
			}
			l, ok2 := v.(*limiter)
			if !ok2 {
				return nil, ErrLimitUnknown
			}

			//设置超时时间
			waitCtx, cancel := context.WithTimeout(ctx, l.t)
			defer cancel()

			if err1 := l.limiter.Wait(waitCtx); err1 != nil {
				return nil, ErrLimitExceed
			}

			// allowed
			return handler(ctx, req)
		}
	}
}

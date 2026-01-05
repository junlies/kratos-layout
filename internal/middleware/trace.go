package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
)

const traceIdKey = "X-Trace-Id"

func TraceMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			reply, err = handler(ctx, req)

			span := trace.SpanFromContext(ctx)
			sc := span.SpanContext()
			if !sc.IsValid() {
				return reply, err
			}

			if tr, ok := transport.FromServerContext(ctx); ok {
				tr.ReplyHeader().Set(traceIdKey, sc.TraceID().String())
			}

			return reply, err
		}
	}
}

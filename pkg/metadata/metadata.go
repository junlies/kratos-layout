package metadata

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"io"
)

func Set(ctx context.Context, key, value string) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	for _, k := range Keys {
		if key == k {
			return nil, errors.New("illegal metadata key")
		}
	}
	return metadata.AppendToClientContext(ctx, key, value), nil
}

func Get(ctx context.Context, key string) (result string, err error) {
	if ctx == nil {
		return "", errors.New("ctx is nil")
	}

	md, ok := metadata.FromServerContext(ctx)
	if ok {
		result = md.Get(key)
	} else {
		err = errors.New("no metadata id in context")
	}

	return
}

func GetHeader(ctx context.Context) (transport.Header, bool) {
	if ctx == nil {
		return nil, false
	}

	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return nil, false
	}
	header := tr.RequestHeader()
	return header, true
}

func GetHttpRequestBody(ctx context.Context) ([]byte, bool) {
	if ctx == nil {
		return nil, false
	}

	r, ok := http.RequestFromServerContext(ctx)
	if !ok {
		return nil, false
	}

	var bodyByte []byte
	if r.Body != nil {
		bodyByte, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyByte))
	}

	return bodyByte, true
}

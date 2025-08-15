package data

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	db  DbClient
	rdb RedisClient
}

// NewData .
func NewData(c *conf.Bootstrap, provider trace.TracerProvider, meterProvider metric.MeterProvider, logger log.Logger) (*Data, func(), error) {
	dbClient, dbClean, err := newDB(c, provider)
	if err != nil {
		panic(err)
	}
	redisClient, rdbClean, err := newRedis(c, provider, meterProvider)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		dbClean()
		rdbClean()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:  dbClient,
		rdb: redisClient,
	}, cleanup, nil
}

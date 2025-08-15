package data

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	Shard       int32
	RedisClient map[string]map[Shard]*redis.Client
)

func newRedis(c *conf.Bootstrap, provider trace.TracerProvider, meterProvider metric.MeterProvider) (RedisClient, func(), error) {
	redisClient := make(map[string]map[Shard]*redis.Client)
	for alias, rds := range c.GetData().GetRedis() {
		redisClient[alias] = make(map[Shard]*redis.Client)
		for _, shard := range rds.GetShards() {
			client := redis.NewClient(&redis.Options{
				Addr:         rds.GetAddr(),
				Password:     rds.GetPassword(),
				DB:           int(shard),
				DialTimeout:  rds.GetReadTimeout().AsDuration(),
				ReadTimeout:  rds.GetReadTimeout().AsDuration(),
				WriteTimeout: rds.GetWriteTimeout().AsDuration(),
			})

			redisClient[alias][Shard(shard)] = client

			// Enable tracing instrumentation.
			if err := redisotel.InstrumentTracing(client,
				redisotel.WithTracerProvider(provider),
				redisotel.WithAttributes(semconv.DBSystemRedis),
			); err != nil {
				return nil, nil, err
			}

			// Enable metrics instrumentation.
			if err := redisotel.InstrumentMetrics(client,
				redisotel.WithMeterProvider(meterProvider),
				redisotel.WithAttributes(semconv.DBSystemRedis),
			); err != nil {
				return nil, nil, err
			}
		}
	}
	return redisClient, func() {
		for _, rds := range redisClient {
			for _, client := range rds {
				if err := client.Close(); err != nil {

				}
			}
		}
	}, nil
}

func (c RedisClient) GetRdbClient(alias string, shard Shard) *redis.Client {
	return c[alias][shard]
}

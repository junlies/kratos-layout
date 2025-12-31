package data

import (
	"time"

	"github.com/go-kratos/kratos-layout/internal/conf"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

type (
	DbClient map[string]*gorm.DB
)

func newDB(c *conf.Bootstrap, provider trace.TracerProvider) (DbClient, func(), error) {
	dbClient := make(map[string]*gorm.DB)
	for alias, cfg := range c.GetData().GetDatabase() {
		switch cfg.Driver {
		case mysql.DefaultDriverName:
			client, err := newMySQL(cfg, provider)
			if err != nil {
				panic(err)
			}
			dbClient[alias] = client
		case PostgreSQLDriverName:
			client, err := newPostgreSQL(cfg, provider)
			if err != nil {
				panic(err)
			}
			dbClient[alias] = client
		default:
			panic("unexpected driver: " + cfg.Driver)
		}
	}
	return dbClient, func() {
		for _, db := range dbClient {
			d, _ := db.DB()
			d.Close()
		}
	}, nil
}

func newMySQL(cfg *conf.Data_Database, provider trace.TracerProvider) (*gorm.DB, error) {
	conn, err := gorm.Open(mysql.Open(cfg.GetSource()))
	if err != nil {
		panic(err)
	}
	if !cfg.GetConnMaxLifetime().IsValid() {
		cfg.ConnMaxLifetime = durationpb.New(30 * time.Minute)
	}
	if cfg.GetMaxOpenConn() == 0 {
		cfg.MaxOpenConn = 20
	}
	if cfg.GetMaxIdleConn() == 0 {
		cfg.MaxIdleConn = 10
	}

	sqlDb, _ := conn.DB()
	sqlDb.SetConnMaxLifetime(cfg.GetConnMaxLifetime().AsDuration())
	sqlDb.SetMaxOpenConns(int(cfg.GetMaxOpenConn()))
	sqlDb.SetMaxIdleConns(int(cfg.GetMaxIdleConn()))
	if tcErr := conn.Use(tracing.NewPlugin(tracing.WithTracerProvider(provider))); tcErr != nil {
		return nil, tcErr
	}

	return conn, nil
}

const PostgreSQLDriverName = "postgresql"

func newPostgreSQL(cfg *conf.Data_Database, provider trace.TracerProvider) (*gorm.DB, error) {
	conn, err := gorm.Open(postgres.Open(cfg.GetSource()))
	if err != nil {
		panic(err)
	}
	if !cfg.GetConnMaxLifetime().IsValid() {
		cfg.ConnMaxLifetime = durationpb.New(30 * time.Minute)
	}
	if cfg.GetMaxOpenConn() == 0 {
		cfg.MaxOpenConn = 20
	}
	if cfg.GetMaxIdleConn() == 0 {
		cfg.MaxIdleConn = 10
	}

	sqlDb, _ := conn.DB()
	sqlDb.SetConnMaxLifetime(cfg.GetConnMaxLifetime().AsDuration())
	sqlDb.SetMaxOpenConns(int(cfg.GetMaxOpenConn()))
	sqlDb.SetMaxIdleConns(int(cfg.GetMaxIdleConn()))
	if tcErr := conn.Use(tracing.NewPlugin(tracing.WithTracerProvider(provider))); tcErr != nil {
		return nil, tcErr
	}

	return conn, nil
}

func (c DbClient) GetDbClient(alias string) *gorm.DB {
	return c[alias]
}

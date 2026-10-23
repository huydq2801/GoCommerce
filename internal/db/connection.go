package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gocommerce/internal/config"
)

func Open(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	conf.MaxConns = 10
	conf.MinConns = 1
	conf.MaxConnLifetime = 30 * time.Minute
	conf.MaxConnIdleTime = 5 * time.Minute
	conf.ConnConfig.ConnectTimeout = 5 * time.Second
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(pingCtx, conf)
	if err != nil {
		return nil, errors.New("could not create database pool")
	}
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, errors.New("database unavailable: check DB_* settings and PostgreSQL readiness")
	}
	return pool, nil
}

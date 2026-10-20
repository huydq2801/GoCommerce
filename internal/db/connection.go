package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"huydq2801.com/golearn/internal/config"
	"huydq2801.com/golearn/internal/db/sqlc"
)

var DB *sqlc.Queries

func InitDataBase() error {
	connStr := config.NewConfig().DNS()
	conf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("error parsing database connection string: %v", err)
	}
	conf.MaxConns = 50
	conf.MinConns = 5
	conf.MaxConnLifetime = 30 * time.Minute
	conf.MaxConnIdleTime = 5 * time.Minute
	conf.HealthCheckPeriod = 1 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	DBPool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("error creating database pool: %v", err)
	}
	DB = sqlc.New(DBPool)
	if err := DBPool.Ping(ctx); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}
	log.Println("Database connection successful")
	return nil
}

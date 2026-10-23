package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gocommerce/internal/db/sqlc"
)

type ShoppingStore struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewShoppingStore(pool *pgxpool.Pool) *ShoppingStore {
	return &ShoppingStore{pool: pool, queries: sqlc.New(pool)}
}
func (s *ShoppingStore) Queries() *sqlc.Queries { return s.queries }
func (s *ShoppingStore) Transact(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	if err = fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

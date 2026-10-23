package repository

import (
	"context"

	"github.com/google/uuid"

	"gocommerce/internal/db/sqlc"
)

type SQLUserRepository struct{ db sqlc.Querier }

func NewSQLUserRepository(db sqlc.Querier) UserRepository { return &SQLUserRepository{db: db} }
func (r *SQLUserRepository) Create(ctx context.Context, p sqlc.CreateUserParams) (sqlc.User, error) {
	return r.db.CreateUser(ctx, p)
}
func (r *SQLUserRepository) List(ctx context.Context, p sqlc.ListUsersParams) ([]sqlc.User, error) {
	return r.db.ListUsers(ctx, p)
}
func (r *SQLUserRepository) Count(ctx context.Context, search string) (int64, error) {
	return r.db.CountUsers(ctx, search)
}
func (r *SQLUserRepository) FindByUUID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.db.GetUser(ctx, id)
}
func (r *SQLUserRepository) Update(ctx context.Context, p sqlc.UpdateUserParams) (sqlc.User, error) {
	return r.db.UpdateUser(ctx, p)
}
func (r *SQLUserRepository) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.db.SoftDeleteUser(ctx, id)
}

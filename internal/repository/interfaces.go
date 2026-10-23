package repository

import (
	"context"

	"github.com/google/uuid"

	"gocommerce/internal/db/sqlc"
)

type UserRepository interface {
	Create(context.Context, sqlc.CreateUserParams) (sqlc.User, error)
	List(context.Context, sqlc.ListUsersParams) ([]sqlc.User, error)
	Count(context.Context, string) (int64, error)
	FindByUUID(context.Context, uuid.UUID) (sqlc.User, error)
	Update(context.Context, sqlc.UpdateUserParams) (sqlc.User, error)
	Delete(context.Context, uuid.UUID) (int64, error)
}

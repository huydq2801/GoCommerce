package v1service

import (
	"context"

	"github.com/google/uuid"

	"huydq2801.com/golearn/internal/db/sqlc"
	"huydq2801.com/golearn/internal/utils"
)

type UserService interface {
	CreateUser(context.Context, sqlc.CreateUserParams) (sqlc.User, error)
	GetAllUsers(context.Context, utils.Page) ([]sqlc.User, int64, error)
	GetUserByUUID(context.Context, uuid.UUID) (sqlc.User, error)
	UpdateUser(context.Context, sqlc.UpdateUserParams) (sqlc.User, error)
	DeleteUser(context.Context, uuid.UUID) error
}

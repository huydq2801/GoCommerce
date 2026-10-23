package v1service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"gocommerce/internal/db/sqlc"
	"gocommerce/internal/repository"
	"gocommerce/internal/utils"
)

type userService struct{ repo repository.UserRepository }

func NewUserService(repo repository.UserRepository) UserService { return &userService{repo: repo} }
func hashPassword(value string) (string, error) {
	if len([]byte(value)) > 72 {
		return "", utils.NewError("Password must not exceed 72 bytes", utils.ErrCodeBadRequest)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", utils.WrapError("Could not hash password", utils.ErrCodeInternal, err)
	}
	return string(hash), nil
}
func (s *userService) CreateUser(ctx context.Context, p sqlc.CreateUserParams) (sqlc.User, error) {
	p.UserEmail = utils.NormalizeString(p.UserEmail)
	p.UserFullname = strings.TrimSpace(p.UserFullname)
	if p.UserFullname == "" {
		return sqlc.User{}, utils.NewError("Name must not be blank", utils.ErrCodeBadRequest)
	}
	// Registration never accepts privilege or status assignments from a caller.
	p.UserLevel = 3
	p.UserStatus = 1
	hash, err := hashPassword(p.UserPassword)
	if err != nil {
		return sqlc.User{}, err
	}
	p.UserPassword = hash
	u, err := s.repo.Create(ctx, p)
	return u, utils.DatabaseError(err)
}
func (s *userService) GetAllUsers(ctx context.Context, p utils.Page) ([]sqlc.User, int64, error) {
	p.Defaults()
	items, err := s.repo.List(ctx, sqlc.ListUsersParams{Search: strings.TrimSpace(p.Search), PageLimit: int32(p.Limit), PageOffset: p.Offset()})
	if err != nil {
		return nil, 0, utils.DatabaseError(err)
	}
	count, err := s.repo.Count(ctx, strings.TrimSpace(p.Search))
	return items, count, utils.DatabaseError(err)
}
func (s *userService) GetUserByUUID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	u, err := s.repo.FindByUUID(ctx, id)
	return u, utils.DatabaseError(err)
}
func (s *userService) UpdateUser(ctx context.Context, p sqlc.UpdateUserParams) (sqlc.User, error) {
	if p.UserEmail != nil {
		email := utils.NormalizeString(*p.UserEmail)
		p.UserEmail = &email
	}
	if p.UserFullname != nil {
		name := strings.TrimSpace(*p.UserFullname)
		if name == "" {
			return sqlc.User{}, utils.NewError("Name must not be blank", utils.ErrCodeBadRequest)
		}
		p.UserFullname = &name
	}
	if p.UserPassword != nil {
		hash, err := hashPassword(*p.UserPassword)
		if err != nil {
			return sqlc.User{}, err
		}
		p.UserPassword = &hash
	}
	u, err := s.repo.Update(ctx, p)
	return u, utils.DatabaseError(err)
}
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	rows, err := s.repo.Delete(ctx, id)
	if err != nil {
		return utils.DatabaseError(err)
	}
	if rows == 0 {
		return utils.NewError("User not found", utils.ErrCodeNotFound)
	}
	return nil
}

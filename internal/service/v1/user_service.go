package v1service

import (
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"huydq2801.com/golearn/internal/models"
	"huydq2801.com/golearn/internal/repository"
	"huydq2801.com/golearn/internal/utils"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (us *userService) GetAllUsers(search string, page, limit int) ([]models.User, error) {
	users, err := us.repo.FindAll()
	if err != nil {
		return nil, utils.NewError("can't get users", utils.ErrCodeInternal)
	}
	var filteredUsers []models.User
	if search != "" {
		search = strings.ToLower(search)
		for _, user := range users {
			name := strings.ToLower(user.Name)
			email := strings.ToLower(user.Email)
			if strings.Contains(name, search) || strings.Contains(email, search) {
				filteredUsers = append(filteredUsers, user)
			}
		}
	} else {
		filteredUsers = users
	}
	start := (page - 1) * limit
	if start >= len(filteredUsers) {
		return []models.User{}, nil
	}
	end := start + limit
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}
	return filteredUsers[start:end], nil
}

func (us *userService) CreateUser(user models.User) (models.User, error) {
	user.Email = utils.NormalizeString(user.Email)
	if _, exists := us.repo.FindByEmail(user.Email); exists {
		return models.User{}, utils.NewError("email already exists", utils.ErrCodeConflict)
	}
	user.UUID = uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, utils.WrapError("can't hash password", utils.ErrCodeInternal, err)
	}
	user.Password = string(hashedPassword)
	if err := us.repo.Create(user); err != nil {
		return models.User{}, utils.WrapError("can't create user", utils.ErrCodeInternal, err)
	}
	return user, nil
}

func (us *userService) GetUserByUUID(uuid string) (models.User, error) {
	user, ok := us.repo.FindByUUID(uuid)
	if !ok {
		return models.User{}, utils.NewError("user not found", utils.ErrCodeNotFound)
	}
	return user, nil
}

func (us *userService) UpdateUser(uuid string, user models.User) (models.User, error) {
	user.Email = utils.NormalizeString(user.Email)
	if u, exists := us.repo.FindByEmail(user.Email); exists && u.UUID != uuid {
		return models.User{}, utils.NewError("email already exists", utils.ErrCodeConflict)
	}
	currentUser, found := us.repo.FindByUUID(uuid)
	if !found {
		return models.User{}, utils.NewError("user not found", utils.ErrCodeNotFound)
	}
	if user.Name != "" {
		currentUser.Name = user.Name
	}
	if user.Email != "" {
		currentUser.Email = user.Email
	}
	if user.Age > 0 {
		currentUser.Age = user.Age
	}
	if user.Level > 0 {
		currentUser.Level = user.Level
	}
	if user.Status > 0 {
		currentUser.Status = user.Status
	}
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, utils.WrapError("can't hash password", utils.ErrCodeInternal, err)
		}
		currentUser.Password = string(hashedPassword)
	}
	if err := us.repo.Update(uuid, currentUser); err != nil {
		return models.User{}, utils.WrapError("can't update user", utils.ErrCodeInternal, err)
	}
	return currentUser, nil
}

func (us *userService) DeleteUser(uuid string) error {
	err := us.repo.Delete(uuid)
	if err != nil {
		return utils.WrapError("can't delete user", utils.ErrCodeInternal, err)
	}
	return nil
}

package v1dto

import (
	"huydq2801.com/golearn/internal/models"
)

type UserDTO struct {
	UUID   string `json:"uuid"`
	Name   string `json:"full_name"`
	Email  string `json:"email_address"`
	Age    int    `json:"age"`
	Level  string `json:"level"`
	Status string `json:"status"`
}

type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,email_advanced"`
	Age      int    `json:"age" binding:"required,gt=0"`
	Password string `json:"password" binding:"required,min=8,password_strong"`
	Status   int    `json:"status" binding:"required,oneof=1 2"`
	Level    int    `json:"level" binding:"required,oneof=1 2"`
}

type UpdateUserInput struct {
	Name     string `json:"name" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email,email_advanced"`
	Age      int    `json:"age" binding:"omitempty,gt=0"`
	Password string `json:"password" binding:"omitempty,min=8,password_strong"`
	Status   int    `json:"status" binding:"omitempty,oneof=1 2"`
	Level    int    `json:"level" binding:"omitempty,oneof=1 2"`
}

func (input *CreateUserInput) MapCreateInputToModel() models.User {
	return models.User{
		Name:     input.Name,
		Email:    input.Email,
		Age:      input.Age,
		Password: input.Password,
		Status:   input.Status,
		Level:    input.Level,
	}
}
func (input *UpdateUserInput) MapUpdateInputToModel() models.User {
	return models.User{
		Name:     input.Name,
		Email:    input.Email,
		Age:      input.Age,
		Password: input.Password,
		Status:   input.Status,
		Level:    input.Level,
	}
}
func MapUserToDTO(user models.User) *UserDTO {
	return &UserDTO{
		UUID:   user.UUID,
		Name:   user.Name,
		Email:  user.Email,
		Age:    int(user.Age),
		Level:  mapLevelText(user.Level),
		Status: mapStatusText(user.Status),
	}
}
func MapUsersToDTOs(users []models.User) []UserDTO {
	dtoList := make([]UserDTO, 0, len(users))
	for _, user := range users {
		dtoList = append(dtoList, *MapUserToDTO(user))
	}
	return dtoList
}
func mapStatusText(status int) string {
	switch status {
	case 1:
		return "Show"
	case 2:
		return "Hide"
	}
	return "None"
}

func mapLevelText(level int) string {
	switch level {
	case 1:
		return "Admin"
	case 2:
		return "User"
	}
	return "None"
}

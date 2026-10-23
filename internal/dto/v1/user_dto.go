package v1dto

import (
	"huydq2801.com/golearn/internal/db/sqlc"
	"huydq2801.com/golearn/internal/utils"
)

type UserDTO struct {
	UUID   string `json:"uuid"`
	Name   string `json:"full_name"`
	Email  string `json:"email_address"`
	Age    *int32 `json:"age"`
	Level  string `json:"level"`
	Status string `json:"status"`
}
type CreateUserInput struct {
	Name     string `json:"name" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,email,max=150"`
	Age      int    `json:"age" binding:"required,min=1,max=150"`
	Password string `json:"password" binding:"required,min=8,password_strong"`
}
type UpdateUserInput struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=100"`
	Email    *string `json:"email" binding:"omitempty,email,max=150"`
	Age      *int32  `json:"age" binding:"omitempty,min=1,max=150"`
	Password *string `json:"password" binding:"omitempty,min=8,password_strong"`
	Status   *int32  `json:"status" binding:"omitempty,oneof=1 2 3"`
	Level    *int32  `json:"level" binding:"omitempty,oneof=1 2 3"`
}

func (i *CreateUserInput) MapCreateInputToModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{UserEmail: i.Email, UserPassword: i.Password, UserFullname: i.Name, UserAge: utils.ConverInt32Pointer(i.Age), UserStatus: 1, UserLevel: 3}
}
func MapUserToDTO(u sqlc.User) *UserDTO {
	levels := map[int32]string{1: "Admin", 2: "Moderator", 3: "Member"}
	statuses := map[int32]string{1: "Active", 2: "Inactive", 3: "Banned"}
	return &UserDTO{UUID: u.Uuid.String(), Name: u.UserFullname, Email: u.UserEmail, Age: u.UserAge, Level: levels[u.UserLevel], Status: statuses[u.UserStatus]}
}
func MapUsersToDTOs(users []sqlc.User) []UserDTO {
	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, *MapUserToDTO(u))
	}
	return out
}

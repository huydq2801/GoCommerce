package v1dto



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

func (input *CreateUserInput) MapCreateInputToModel() {
	
}
func (input *UpdateUserInput) MapUpdateInputToModel(){
	
}
func MapUserToDTO() *UserDTO {
	return &UserDTO{
	}
}
func MapUsersToDTOs() []UserDTO {
	dtoList := make([]UserDTO, 0)
	
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

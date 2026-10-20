package v1service


type UserService interface {
	GetAllUsers()
	CreateUser()
	GetUserByUUID()
	UpdateUser()
	DeleteUser()
}

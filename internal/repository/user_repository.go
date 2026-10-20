package repository

type SQLUserRepository struct {
}

func NewSQLUserRepository() UserRepository {
	return &SQLUserRepository{}
}

func (ur *SQLUserRepository) FindAll() {
}

func (ur *SQLUserRepository) Create() {

}

func (ur *SQLUserRepository) FindByUUID() {

}

func (ur *SQLUserRepository) Update() {

}

func (ur *SQLUserRepository) Delete() {

}

func (ur *SQLUserRepository) FindByEmail() {

}

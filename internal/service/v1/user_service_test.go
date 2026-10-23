package v1service

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"huydq2801.com/golearn/internal/db/sqlc"
	"huydq2801.com/golearn/internal/repository"
)

type userRepoStub struct {
	repository.UserRepository
	captured sqlc.CreateUserParams
	calls    int
}

func (r *userRepoStub) Create(ctx context.Context, p sqlc.CreateUserParams) (sqlc.User, error) {
	r.captured = p
	r.calls++
	return sqlc.User{UserEmail: p.UserEmail, UserPassword: p.UserPassword, UserLevel: p.UserLevel}, nil
}
func TestRegistrationNormalizesHashesAndForcesMember(t *testing.T) {
	repo := &userRepoStub{}
	service := NewUserService(repo)
	password := "GoodPass1!"
	_, err := service.CreateUser(context.Background(), sqlc.CreateUserParams{UserEmail: "  Huy@Example.com ", UserPassword: password, UserFullname: " Huy ", UserLevel: 1, UserStatus: 3})
	if err != nil {
		t.Fatal(err)
	}
	p := repo.captured
	if p.UserEmail != "huy@example.com" || p.UserFullname != "Huy" || p.UserLevel != 3 || p.UserStatus != 1 {
		t.Fatalf("bad normalized values: email=%s role=%d", p.UserEmail, p.UserLevel)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(p.UserPassword), []byte(password)); err != nil {
		t.Fatal("password was not hashed correctly")
	}
}
func TestTooLongPasswordRejectedBeforeDatabase(t *testing.T) {
	repo := &userRepoStub{}
	_, err := NewUserService(repo).CreateUser(context.Background(), sqlc.CreateUserParams{UserFullname: "Huy", UserPassword: strings.Repeat("a", 73)})
	if err == nil || repo.calls != 0 {
		t.Fatal("oversized password reached database")
	}
}

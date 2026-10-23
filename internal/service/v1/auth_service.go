package v1service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gocommerce/internal/db/sqlc"
	"gocommerce/internal/utils"
)

type AuthRepository interface {
	GetUserByEmail(context.Context, string) (sqlc.User, error)
}
type AuthService struct {
	repo   AuthRepository
	secret []byte
}

func NewAuthService(repo AuthRepository, secret string) *AuthService {
	return &AuthService{repo: repo, secret: []byte(secret)}
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, utils.NormalizeString(email))
	if err != nil || user.UserStatus != 1 || bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(password)) != nil {
		return LoginResult{}, utils.NewError("Invalid email or password", utils.ErrCodeUnauthorized)
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{Issuer: "gocommerce", Audience: jwt.ClaimStrings{"gocommerce-api"}, Subject: user.Uuid.String(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute))}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return LoginResult{}, utils.WrapError("Could not create access token", utils.ErrCodeInternal, err)
	}
	return LoginResult{AccessToken: token, TokenType: "Bearer", ExpiresIn: 900}, nil
}

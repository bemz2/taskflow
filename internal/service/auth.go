package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	userService  *UserService
	tokenService *TokenService
}

func NewAuthService(userService *UserService, tokenService *TokenService) *AuthService {
	return &AuthService{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	user, err := s.userService.CreateUser(ctx, email, password)
	if err != nil {
		return "", err
	}

	return s.tokenService.Issue(user.ID)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokenService.Issue(user.ID)
}

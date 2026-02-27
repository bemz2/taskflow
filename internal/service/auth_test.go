package service

import (
	"context"
	"testing"

	"taskflow/internal/domain"
	"taskflow/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceRegisterReturnsToken(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)
	tokenService := NewTokenService("test-secret", mockTTL())
	authService := NewAuthService(userService, tokenService)
	ctx := context.Background()
	createdUser := domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hash",
	}

	repo.
		On("Create", ctx, createdUserMatcher("user@example.com", "secret")).
		Return(createdUser, nil).
		Once()

	token, err := authService.Register(ctx, "user@example.com", "secret")

	require.NoError(t, err)

	parsedUserID, err := tokenService.Parse(token)
	require.NoError(t, err)
	require.Equal(t, createdUser.ID, parsedUserID)
	repo.AssertExpectations(t)
}

func TestAuthServiceRegisterPropagatesCreateUserError(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)
	tokenService := NewTokenService("test-secret", mockTTL())
	authService := NewAuthService(userService, tokenService)

	_, err := authService.Register(context.Background(), "user@example.com", "")

	require.ErrorIs(t, err, domain.ErrEmptyPassword)
}

func TestAuthServiceLoginReturnsToken(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)
	tokenService := NewTokenService("test-secret", mockTTL())
	authService := NewAuthService(userService, tokenService)
	ctx := context.Background()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := domain.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: string(passwordHash),
	}

	repo.
		On("GetByEmail", ctx, "user@example.com").
		Return(user, nil).
		Once()

	token, err := authService.Login(ctx, "user@example.com", "secret")

	require.NoError(t, err)

	parsedUserID, err := tokenService.Parse(token)
	require.NoError(t, err)
	require.Equal(t, user.ID, parsedUserID)
	repo.AssertExpectations(t)
}

func TestAuthServiceLoginRejectsUnknownUser(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)
	tokenService := NewTokenService("test-secret", mockTTL())
	authService := NewAuthService(userService, tokenService)
	ctx := context.Background()

	repo.
		On("GetByEmail", ctx, "user@example.com").
		Return(domain.User{}, assertErrUserNotFound()).
		Once()

	_, err := authService.Login(ctx, "user@example.com", "secret")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	repo.AssertExpectations(t)
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	userService := NewUserService(repo)
	tokenService := NewTokenService("test-secret", mockTTL())
	authService := NewAuthService(userService, tokenService)
	ctx := context.Background()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.
		On("GetByEmail", ctx, "user@example.com").
		Return(domain.User{
			ID:           uuid.New(),
			Email:        "user@example.com",
			PasswordHash: string(passwordHash),
		}, nil).
		Once()

	_, err = authService.Login(ctx, "user@example.com", "wrong-password")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	repo.AssertExpectations(t)
}

package service

import (
	"context"
	"errors"
	"testing"

	"taskflow/internal/domain"
	"taskflow/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceCreateUserRejectsEmptyPassword(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), "user@example.com", "")

	require.ErrorIs(t, err, domain.ErrEmptyPassword)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestUserServiceCreateUserHashesPassword(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.
		On("Create", ctx, mock.MatchedBy(func(user domain.User) bool {
			return user.Email == "user@example.com" &&
				user.ID != uuid.Nil &&
				bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret")) == nil
		})).
		Return(domain.User{ID: uuid.New(), Email: "user@example.com"}, nil).
		Once()

	user, err := svc.CreateUser(ctx, "  USER@example.com  ", "secret")

	require.NoError(t, err)
	require.Equal(t, "user@example.com", user.Email)
	repo.AssertExpectations(t)
}

func TestUserServiceGetUserMapsRepositoryError(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	svc := NewUserService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.
		On("Get", ctx, userID).
		Return(domain.User{}, errors.New("db error")).
		Once()

	_, err := svc.GetUser(ctx, userID)

	require.ErrorIs(t, err, ErrUserNotFound)
	repo.AssertExpectations(t)
}

func TestUserServiceEnsureDevUser(t *testing.T) {
	t.Parallel()

	repo := mocks.NewUserRepository(t)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.
		On("Ensure", ctx, mock.MatchedBy(func(user domain.User) bool {
			return user.ID == devUserID &&
				user.Email == "dev@taskflow.local" &&
				bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("dev-password")) == nil
		})).
		Return(nil).
		Once()

	err := svc.EnsureDevUser(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

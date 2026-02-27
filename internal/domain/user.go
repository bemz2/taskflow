package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidUserEmail  = errors.New("invalid user email")
	ErrEmptyPassword     = errors.New("password is empty")
	ErrEmptyPasswordHash = errors.New("password hash is empty")
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewUser(email, passwordHash string) (User, error) {
	return NewUserWithID(uuid.New(), email, passwordHash)
}

func NewUserWithID(id uuid.UUID, email, passwordHash string) (User, error) {
	email = NormalizeUserEmail(email)
	passwordHash = strings.TrimSpace(passwordHash)

	if id == uuid.Nil {
		return User{}, ErrInvalidUserID
	}
	if email == "" {
		return User{}, ErrInvalidUserEmail
	}
	if passwordHash == "" {
		return User{}, ErrEmptyPasswordHash
	}

	return User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}

func NewUserFromStorage(id uuid.UUID, email, passwordHash string, createdAt time.Time) (User, error) {
	user, err := NewUserWithID(id, email, passwordHash)
	if err != nil {
		return User{}, err
	}

	user.CreatedAt = createdAt
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUserEmail(email string) string {
	return normalizeEmail(email)
}

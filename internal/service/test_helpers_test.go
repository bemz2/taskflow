package service

import (
	"errors"
	"time"

	"taskflow/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func mockTime() time.Time {
	return time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
}

func mockTTL() time.Duration {
	return time.Minute
}

func createdUserMatcher(email, password string) interface{} {
	return mock.MatchedBy(func(user domain.User) bool {
		return user.Email == email &&
			user.ID != uuid.Nil &&
			bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
	})
}

func assertErrUserNotFound() error {
	return errors.New("user not found in storage")
}

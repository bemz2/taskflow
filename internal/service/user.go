package service

import (
	"context"
	"errors"
	"taskflow/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

var devUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Ensure(ctx context.Context, user domain.User) error
}

type UserService struct {
	UserRepository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
		UserRepository: repository,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, password string) (domain.User, error) {
	if password == "" {
		return domain.User{}, domain.ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	user, err := domain.NewUser(email, string(hash))
	if err != nil {
		return domain.User{}, err
	}

	return s.UserRepository.Create(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	user, err := s.UserRepository.Get(ctx, userID)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := s.UserRepository.GetByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) EnsureDevUser(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("dev-password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user, err := domain.NewUserWithID(devUserID, "dev@taskflow.local", string(hash))
	if err != nil {
		return err
	}

	return s.UserRepository.Ensure(ctx, user)
}

package user

import (
	"context"
	"errors"
	"taskflow/internal/domain"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserModel struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	m := toModel(user)

	query, args, err := sq.
		Insert("users").
		Columns("id", "email", "password_hash").
		Values(m.ID, m.Email, m.PasswordHash).
		Suffix("RETURNING id, email, password_hash, created_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, err
	}

	var created UserModel
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.Email,
		&created.PasswordHash,
		&created.CreatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	return toDomain(created)
}

func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	query, args, err := sq.
		Select("id", "email", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, err
	}

	var m UserModel
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&m.ID,
		&m.Email,
		&m.PasswordHash,
		&m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return toDomain(m)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query, args, err := sq.
		Select("id", "email", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"email": domain.NormalizeUserEmail(email)}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return domain.User{}, err
	}

	var m UserModel
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&m.ID,
		&m.Email,
		&m.PasswordHash,
		&m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	return toDomain(m)
}

func (r *UserRepository) Ensure(ctx context.Context, user domain.User) error {
	m := toModel(user)
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, m.ID, m.Email, m.PasswordHash)
	return err
}

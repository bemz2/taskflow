package analytics

import (
	"context"
	"errors"
	"taskflow/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.TaskAnalytics, error) {
	var analytics domain.TaskAnalytics

	err := r.db.QueryRow(ctx, `
		SELECT tasks_created, tasks_completed, updated_at
		FROM task_analytics
		WHERE user_id = $1
	`, userID).Scan(
		&analytics.TasksCreated,
		&analytics.TasksCompleted,
		&analytics.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TaskAnalytics{}, nil
	}
	if err != nil {
		return domain.TaskAnalytics{}, err
	}

	return analytics, nil
}

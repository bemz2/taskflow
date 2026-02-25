package repository

import (
	"context"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskModel struct {
	ID          uuid.UUID  `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Status      bool       `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

type TaskRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewTaskRepository(db *pgxpool.Pool, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context, t TaskModel) (TaskModel, error) {
	const contextKey = "TaskRepositroy.CreateTask"

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	query, args, err := sq.
		Insert("tasks").
		Columns("id", "user_id", "title", "description", "status", "completed_at").
		Values(t.ID, t.UserID, t.Title, t.Description, t.Status, t.CompletedAt).
		Suffix("RETURNING id, user_id, title, description, status, created_at, completed_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		r.logger.Error(contextKey, "failed to build query: %w", err)
		return TaskModel{}, err
	}

	var created TaskModel

	err = r.db.QueryRow(ctx, query, args...).
		Scan(
			&created.ID,
			&created.UserID,
			&created.Title,
			&created.Description,
			&created.Status,
			&created.CreatedAt,
			&created.CompletedAt,
		)
	if err != nil {
		r.logger.Error(contextKey, "failed to QueryRow: %w", err)
		return TaskModel{}, err
	}

	return created, nil
}

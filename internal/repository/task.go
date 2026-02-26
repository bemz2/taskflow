package repository

import (
	"context"
	"fmt"
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
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) CreateTask(ctx context.Context, t TaskModel) (TaskModel, error) {
	const contextKey = "TaskRepository.CreateTask"

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
		return TaskModel{}, fmt.Errorf("%s: failed to build sql: %w", contextKey, err)
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
		return TaskModel{}, fmt.Errorf("%s: failed to query row: %w", contextKey, err)
	}

	return created, err
}

func (r *TaskRepository) ListTasks(ctx context.Context, userID uuid.UUID) ([]TaskModel, error) {
	const contextKey = "TaskRepository.ListTasks"

	query := sq.
		Select("id", "user_id", "title", "description", "status", "created_at", "completed_at").
		From("tasks").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql %w", contextKey, err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query: %w", contextKey, err)
	}
	defer rows.Close()

	var tasks []TaskModel

	for rows.Next() {
		var t TaskModel
		err = rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.CreatedAt,
			&t.CompletedAt,
		)
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed scan: %w", contextKey, err)
	}
	return tasks, err
}

func (r *TaskRepository) GetTask(ctx context.Context) {

}

func (r *TaskRepository) DeleteTask(ctx context.Context) {

}

func (r *TaskRepository) UpdateTask(ctx context.Context) {

}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context) {

}

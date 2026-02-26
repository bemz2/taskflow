package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTaskNotFound = errors.New("task not found")

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

	var task TaskModel

	err = r.db.QueryRow(ctx, query, args...).
		Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.CompletedAt,
		)
	if err != nil {
		return TaskModel{}, fmt.Errorf("%s: failed to query row: %w", contextKey, err)
	}

	return task, nil
}

func (r *TaskRepository) ListTasks(ctx context.Context, userID uuid.UUID) ([]TaskModel, error) {
	const contextKey = "TaskRepository.ListTasks"

	query, args, err := sq.
		Select("id", "user_id", "title", "description", "status", "created_at", "completed_at").
		From("tasks").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return nil, fmt.Errorf("%s: failed to build sql %w", contextKey, err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query: %w", contextKey, err)
	}
	defer rows.Close()

	var tasks []TaskModel

	for rows.Next() {
		var task TaskModel
		err = rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.CompletedAt,
		)
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed scan: %w", contextKey, err)
	}
	return tasks, nil
}

func (r *TaskRepository) GetTask(ctx context.Context, id, userID uuid.UUID) (TaskModel, error) {
	const contextKey = "TaskRepository.GetTask"

	query := sq.
		Select("id", "user_id", "title", "description", "status", "created_at", "completed_at").
		From("tasks").
		Where(sq.Eq{
			"id":      id,
			"user_id": userID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return TaskModel{}, fmt.Errorf("%s: failed to build sql: %w", contextKey, err)
	}

	var task TaskModel

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskModel{}, ErrTaskNotFound
	}

	if err != nil {
		return TaskModel{}, fmt.Errorf("%s: failed to query row: %w", contextKey, err)
	}

	return task, nil
}

func (r *TaskRepository) DeleteTask(ctx context.Context, id, userID uuid.UUID) error {
	const contextKey = "TaskRepository.DeleteTask"

	query, args, err := sq.Delete("tasks").
		Where(sq.Eq{
			"id":      id,
			"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("%s: failed to build sql: %w", contextKey, err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: exec delete: %w", contextKey, err)
	}
	affected := result.RowsAffected()
	if affected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) UpdateTask(ctx context.Context, t TaskModel) (TaskModel, error) {
	const contextKey = "TaskRepository.UpdateTask"

	query, args, err := sq.
		Update("tasks").
		Set("title", t.Title).
		Set("description", t.Description).
		Set("status", t.Status).
		Set("completed_at", t.CompletedAt).
		Where(sq.Eq{
			"id":      t.ID,
			"user_id": t.UserID,
		}).
		Suffix("RETURNING id, user_id, title, description, status, created_at, completed_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return TaskModel{}, fmt.Errorf("%s: build sql: %w", contextKey, err)
	}

	var updated TaskModel

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Title,
		&updated.Description,
		&updated.Status,
		&updated.CreatedAt,
		&updated.CompletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return TaskModel{}, ErrTaskNotFound
	}

	if err != nil {
		return TaskModel{}, fmt.Errorf("%s: scan: %w", contextKey, err)
	}

	return updated, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, id, userID uuid.UUID, status string) error {
	const contextKey = "TaskRepository.UpdateTaskStatus"

	query, args, err := sq.Update("tasks").
		Set("status", status).
		Where(sq.Eq{
			"id":      id,
			"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: build sql: %w", contextKey, err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: exec update task status: %w", contextKey, err)
	}
	affected := result.RowsAffected()
	if affected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

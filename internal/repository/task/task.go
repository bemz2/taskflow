package task

import (
	"context"
	"errors"
	"fmt"
	"taskflow/internal/domain"
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

var allowedSortColumns = map[string]string{
	"created_at":   "created_at",
	"title":        "title",
	"status":       "status",
	"completed_at": "completed_at",
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) Create(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {

	m := toModel(task)

	query, args, err := sq.
		Insert("tasks").
		Columns("id", "user_id", "title", "description", "status", "completed_at").
		Values(m.ID, m.UserID, m.Title, m.Description, m.Status, m.CompletedAt).
		Suffix("RETURNING id, user_id, title, description, status, created_at, completed_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return domain.Task{}, err
	}

	var created TaskModel

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.UserID,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.CreatedAt,
		&created.CompletedAt,
	)

	if err != nil {
		return domain.Task{}, err
	}

	return toDomain(created)
}

func (r *TaskRepository) Get(
	ctx context.Context,
	id, userID uuid.UUID,
) (domain.Task, error) {

	query, args, err := sq.
		Select("id", "user_id", "title", "description", "status", "created_at", "completed_at").
		From("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return domain.Task{}, err
	}

	var m TaskModel

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&m.ID,
		&m.UserID,
		&m.Title,
		&m.Description,
		&m.Status,
		&m.CreatedAt,
		&m.CompletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}

	return toDomain(m)
}

func (r *TaskRepository) Update(
	ctx context.Context,
	task domain.Task,
) (domain.Task, error) {

	m := toModel(task)

	query, args, err := sq.
		Update("tasks").
		Set("title", m.Title).
		Set("description", m.Description).
		Set("status", m.Status).
		Set("completed_at", m.CompletedAt).
		Where(sq.Eq{"id": m.ID, "user_id": m.UserID}).
		Suffix("RETURNING id, user_id, title, description, status, created_at, completed_at").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return domain.Task{}, err
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
		return domain.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}

	return toDomain(updated)
}

func (r *TaskRepository) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {

	query, args, err := sq.
		Delete("tasks").
		Where(sq.Eq{"id": id, "user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	filter domain.TaskFilter,
) ([]domain.Task, error) {

	filter.Normalize()

	builder := sq.
		Select("id", "user_id", "title", "description", "status", "created_at", "completed_at").
		From("tasks").
		Where(sq.Eq{"user_id": userID}).
		Limit(uint64(filter.Limit)).
		Offset(uint64(filter.Offset)).
		PlaceholderFormat(sq.Dollar)

	if filter.Status != nil {
		builder = builder.Where(sq.Eq{"status": string(*filter.Status)})
	}

	if filter.Search != nil {
		builder = builder.Where("title ILIKE ?", "%"+*filter.Search+"%")
	}

	sortColumn := allowedSortColumns[filter.SortBy]
	if sortColumn == "" {
		sortColumn = allowedSortColumns["created_at"]
	}

	builder = builder.OrderBy(fmt.Sprintf("%s %s", sortColumn, filter.SortDir))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Task

	for rows.Next() {
		var m TaskModel
		if err := rows.Scan(
			&m.ID,
			&m.UserID,
			&m.Title,
			&m.Description,
			&m.Status,
			&m.CreatedAt,
			&m.CompletedAt,
		); err != nil {
			return nil, err
		}

		task, err := toDomain(m)
		if err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	return result, rows.Err()
}

package service

import (
	"context"
	"taskflow/internal/repository"

	"github.com/google/uuid"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, t repository.TaskModel) (repository.TaskModel, error)
	GetTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (repository.TaskModel, error)
	ListTasks(ctx context.Context, userID uuid.UUID) ([]repository.TaskModel, error)
	UpdateTask(ctx context.Context, t repository.TaskModel) (repository.TaskModel, error)
	DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string) error
}

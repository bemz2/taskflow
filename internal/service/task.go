package service

import (
	"context"
	"errors"
	"taskflow/internal/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("forbidden")
)

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	Get(ctx context.Context, id, userID uuid.UUID) (domain.Task, error)
	List(ctx context.Context, userID uuid.UUID, filter domain.TaskFilter) ([]domain.Task, error)
	Update(ctx context.Context, task domain.Task) (domain.Task, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type TaskService struct {
	TaskRepository TaskRepository
}

func NewTaskService(repository TaskRepository) *TaskService {
	return &TaskService{
		TaskRepository: repository,
	}
}

func (s *TaskService) CreateTask(
	ctx context.Context,
	userID uuid.UUID,
	title, description string,
) (domain.Task, error) {
	task, err := domain.NewTask(userID, title, description)
	if err != nil {
		return domain.Task{}, err
	}

	return s.TaskRepository.Create(ctx, task)
}

func (s *TaskService) GetTask(
	ctx context.Context,
	userID, taskID uuid.UUID,
) (domain.Task, error) {

	task, err := s.TaskRepository.Get(ctx, taskID, userID)
	if err != nil {
		return domain.Task{}, ErrTaskNotFound
	}

	return task, nil
}

func (s *TaskService) ChangeStatus(
	ctx context.Context,
	userID, taskID uuid.UUID,
	status domain.Status,
) error {

	task, err := s.TaskRepository.Get(ctx, taskID, userID)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := task.ChangeStatus(status, time.Now()); err != nil {
		return err
	}

	_, err = s.TaskRepository.Update(ctx, task)
	return err
}

func (s *TaskService) UpdateTask(
	ctx context.Context,
	userID, taskID uuid.UUID,
	title, description *string,
) (domain.Task, error) {

	task, err := s.TaskRepository.Get(ctx, taskID, userID)
	if err != nil {
		return domain.Task{}, ErrTaskNotFound
	}

	if title != nil {
		if err := task.Rename(*title); err != nil {
			return domain.Task{}, err
		}
	}

	if description != nil {
		task.ChangeDescription(*description)
	}

	return s.TaskRepository.Update(ctx, task)
}

func (s *TaskService) DeleteTask(
	ctx context.Context,
	userID, taskID uuid.UUID,
) error {
	return s.TaskRepository.Delete(ctx, taskID, userID)
}
func (s *TaskService) ListTasks(
	ctx context.Context,
	userID uuid.UUID,
	filter domain.TaskFilter,
) ([]domain.Task, error) {
	return s.TaskRepository.List(ctx, userID, filter)
}

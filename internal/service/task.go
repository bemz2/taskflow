package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"taskflow/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("forbidden")
)

const taskCacheTTL = 5 * time.Minute

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	Get(ctx context.Context, id, userID uuid.UUID) (domain.Task, error)
	List(ctx context.Context, userID uuid.UUID, filter domain.TaskFilter) ([]domain.Task, error)
	Update(ctx context.Context, task domain.Task) (domain.Task, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type TaskCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type redisTaskCache struct {
	client redis.Cmdable
}

type TaskService struct {
	TaskRepository TaskRepository
	TaskCache      TaskCache
}

func NewRedisTaskCache(client redis.Cmdable) TaskCache {
	if client == nil {
		return nil
	}

	return &redisTaskCache{client: client}
}

func (c *redisTaskCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisTaskCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisTaskCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func NewTaskService(repository TaskRepository, cache TaskCache) *TaskService {
	return &TaskService{
		TaskRepository: repository,
		TaskCache:      cache,
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

	createdTask, err := s.TaskRepository.Create(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	s.cacheTask(ctx, createdTask)

	return createdTask, nil
}

func (s *TaskService) GetTask(
	ctx context.Context,
	userID, taskID uuid.UUID,
) (domain.Task, error) {
	if task, ok := s.getCachedTask(ctx, userID, taskID); ok {
		return task, nil
	}

	task, err := s.TaskRepository.Get(ctx, taskID, userID)
	if err != nil {
		return domain.Task{}, ErrTaskNotFound
	}

	s.cacheTask(ctx, task)

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

	updatedTask, err := s.TaskRepository.Update(ctx, task)
	if err != nil {
		return err
	}

	s.cacheTask(ctx, updatedTask)

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

	updatedTask, err := s.TaskRepository.Update(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	s.cacheTask(ctx, updatedTask)

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(
	ctx context.Context,
	userID, taskID uuid.UUID,
) error {
	if err := s.TaskRepository.Delete(ctx, taskID, userID); err != nil {
		return err
	}

	s.deleteCachedTask(ctx, userID, taskID)

	return nil
}
func (s *TaskService) ListTasks(
	ctx context.Context,
	userID uuid.UUID,
	filter domain.TaskFilter,
) ([]domain.Task, error) {
	return s.TaskRepository.List(ctx, userID, filter)
}

func (s *TaskService) taskCacheKey(userID, taskID uuid.UUID) string {
	return fmt.Sprintf("task:%s:%s", userID.String(), taskID.String())
}

func (s *TaskService) getCachedTask(ctx context.Context, userID, taskID uuid.UUID) (domain.Task, bool) {
	if s.TaskCache == nil {
		return domain.Task{}, false
	}

	payload, err := s.TaskCache.Get(ctx, s.taskCacheKey(userID, taskID))
	if err != nil {
		return domain.Task{}, false
	}

	var task domain.Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return domain.Task{}, false
	}

	return task, true
}

func (s *TaskService) cacheTask(ctx context.Context, task domain.Task) {
	if s.TaskCache == nil {
		return
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return
	}

	_ = s.TaskCache.Set(ctx, s.taskCacheKey(task.UserID, task.ID), string(payload), taskCacheTTL)
}

func (s *TaskService) deleteCachedTask(ctx context.Context, userID, taskID uuid.UUID) {
	if s.TaskCache == nil {
		return
	}

	_ = s.TaskCache.Delete(ctx, s.taskCacheKey(userID, taskID))
}

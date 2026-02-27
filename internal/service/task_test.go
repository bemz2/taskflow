package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"taskflow/internal/domain"
	"taskflow/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTaskServiceCreateTask(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()

	repo.
		On("Create", ctx, mock.MatchedBy(func(task domain.Task) bool {
			return task.UserID == userID &&
				task.Title == "Title" &&
				task.Description == "Description" &&
				task.Status == domain.StatusPending &&
				task.ID != uuid.Nil &&
				!task.CreatedAt.IsZero()
		})).
		Return(domain.Task{ID: uuid.New(), UserID: userID, Title: "Title"}, nil).
		Once()

	task, err := svc.CreateTask(ctx, userID, "  Title  ", "  Description  ")

	require.NoError(t, err)
	require.Equal(t, userID, task.UserID)
	repo.AssertExpectations(t)
}

func TestTaskServiceGetTaskReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	repo.
		On("Get", ctx, taskID, userID).
		Return(domain.Task{}, errors.New("db error")).
		Once()

	_, err := svc.GetTask(ctx, userID, taskID)

	require.ErrorIs(t, err, ErrTaskNotFound)
	repo.AssertExpectations(t)
}

func TestTaskServiceGetTaskReturnsTask(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()
	expected := domain.Task{
		ID:        taskID,
		UserID:    userID,
		Title:     "Task",
		Status:    domain.StatusPending,
		CreatedAt: mockTime(),
	}

	repo.
		On("Get", ctx, taskID, userID).
		Return(expected, nil).
		Once()

	task, err := svc.GetTask(ctx, userID, taskID)

	require.NoError(t, err)
	require.Equal(t, expected, task)
	repo.AssertExpectations(t)
}

func TestTaskServiceChangeStatusUpdatesTask(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := domain.Task{
		ID:        taskID,
		UserID:    userID,
		Title:     "Task",
		Status:    domain.StatusPending,
		CreatedAt: mockTime(),
	}

	repo.
		On("Get", ctx, taskID, userID).
		Return(task, nil).
		Once()
	repo.
		On("Update", ctx, mock.MatchedBy(func(updated domain.Task) bool {
			return updated.ID == taskID &&
				updated.Status == domain.StatusDone &&
				updated.CompletedAt != nil
		})).
		Return(task, nil).
		Once()

	err := svc.ChangeStatus(ctx, userID, taskID, domain.StatusDone)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTaskServiceChangeStatusReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	repo.
		On("Get", ctx, taskID, userID).
		Return(domain.Task{}, errors.New("db error")).
		Once()

	err := svc.ChangeStatus(ctx, userID, taskID, domain.StatusDone)

	require.ErrorIs(t, err, ErrTaskNotFound)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestTaskServiceChangeStatusRejectsInvalidTransition(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := domain.Task{
		ID:        taskID,
		UserID:    userID,
		Title:     "Task",
		Status:    domain.StatusDone,
		CreatedAt: mockTime(),
		CompletedAt: func() *time.Time {
			tm := mockTime()
			return &tm
		}(),
	}

	repo.
		On("Get", ctx, taskID, userID).
		Return(task, nil).
		Once()

	err := svc.ChangeStatus(ctx, userID, taskID, domain.StatusCancelled)

	require.ErrorIs(t, err, domain.ErrInvalidTransition)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestTaskServiceUpdateTaskRejectsEmptyTitle(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()
	emptyTitle := "   "

	task := domain.Task{
		ID:        taskID,
		UserID:    userID,
		Title:     "Task",
		Status:    domain.StatusPending,
		CreatedAt: mockTime(),
	}

	repo.
		On("Get", ctx, taskID, userID).
		Return(task, nil).
		Once()

	_, err := svc.UpdateTask(ctx, userID, taskID, &emptyTitle, nil)

	require.ErrorIs(t, err, domain.ErrEmptyTitle)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestTaskServiceUpdateTaskUpdatesFields(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()
	title := "  Renamed  "
	description := "  Updated description  "

	task := domain.Task{
		ID:          taskID,
		UserID:      userID,
		Title:       "Task",
		Description: "Initial",
		Status:      domain.StatusPending,
		CreatedAt:   mockTime(),
	}

	repo.
		On("Get", ctx, taskID, userID).
		Return(task, nil).
		Once()
	repo.
		On("Update", ctx, mock.MatchedBy(func(updated domain.Task) bool {
			return updated.ID == taskID &&
				updated.Title == "Renamed" &&
				updated.Description == "Updated description"
		})).
		Return(domain.Task{
			ID:          taskID,
			UserID:      userID,
			Title:       "Renamed",
			Description: "Updated description",
			Status:      domain.StatusPending,
			CreatedAt:   mockTime(),
		}, nil).
		Once()

	updated, err := svc.UpdateTask(ctx, userID, taskID, &title, &description)

	require.NoError(t, err)
	require.Equal(t, "Renamed", updated.Title)
	require.Equal(t, "Updated description", updated.Description)
	repo.AssertExpectations(t)
}

func TestTaskServiceDeleteTaskDelegatesToRepository(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	repo.
		On("Delete", ctx, taskID, userID).
		Return(nil).
		Once()

	err := svc.DeleteTask(ctx, userID, taskID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTaskServiceListTasksDelegatesToRepository(t *testing.T) {
	t.Parallel()

	repo := mocks.NewTaskRepository(t)
	svc := NewTaskService(repo)
	ctx := context.Background()
	userID := uuid.New()
	filter := domain.TaskFilter{Limit: 10}
	expected := []domain.Task{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Task",
			Status:    domain.StatusPending,
			CreatedAt: mockTime(),
		},
	}

	repo.
		On("List", ctx, userID, filter).
		Return(expected, nil).
		Once()

	tasks, err := svc.ListTasks(ctx, userID, filter)

	require.NoError(t, err)
	require.Equal(t, expected, tasks)
	repo.AssertExpectations(t)
}

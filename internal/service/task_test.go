package service

import (
	"context"
	"errors"
	"testing"

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

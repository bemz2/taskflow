package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

var (
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrEmptyTitle        = errors.New("title is empty")
	ErrInvalidTaskOwner  = errors.New("invalid task owner")
)

type Task struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Status      Status
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type TaskFilter struct {
	Limit   int
	Offset  int
	Status  *Status
	Search  *string
	SortBy  string
	SortDir string
}

func (f *TaskFilter) Normalize() {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortDir != "asc" && f.SortDir != "desc" {
		f.SortDir = "desc"
	}
}

func NewTask(userID uuid.UUID, title, description string) (Task, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if userID == uuid.Nil {
		return Task{}, ErrInvalidTaskOwner
	}

	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	createdAt := time.Now()

	return Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   createdAt,
	}, nil
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone, StatusCancelled:
		return true
	default:
		return false
	}
}

func NewTaskFromStorage(id, userID uuid.UUID, title, description string, status Status, createdAt time.Time, completedAt *time.Time) (Task, error) {
	if !status.IsValid() {
		return Task{}, ErrInvalidStatus
	}

	if userID == uuid.Nil {
		return Task{}, ErrInvalidTaskOwner
	}

	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	t := Task{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   createdAt,
		CompletedAt: completedAt,
	}

	if err := t.checkInvariants(); err != nil {
		return Task{}, err
	}

	return t, nil
}

func (t *Task) Rename(newTitle string) error {
	newTitle = strings.TrimSpace(newTitle)
	if newTitle == "" {
		return ErrEmptyTitle
	}
	t.Title = newTitle
	return nil
}

func (t *Task) ChangeStatus(to Status, now time.Time) error {
	if !to.IsValid() {
		return ErrInvalidStatus
	}
	if !isAllowedTransition(t.Status, to) {
		return ErrInvalidTransition
	}

	t.Status = to

	if to == StatusDone {
		n := now.UTC()
		t.CompletedAt = &n
	} else {
		t.CompletedAt = nil
	}

	return t.checkInvariants()
}

func (t *Task) ChangeDescription(desc string) {
	desc = strings.TrimSpace(desc)
	t.Description = desc
}

func isAllowedTransition(from, to Status) bool {
	switch from {
	case StatusPending:
		return to == StatusInProgress || to == StatusCancelled || to == StatusDone
	case StatusInProgress:
		return to == StatusDone || to == StatusCancelled
	case StatusDone, StatusCancelled:
		return false
	default:
		return false
	}
}

func (t *Task) checkInvariants() error {
	if t.Status == StatusDone && t.CompletedAt == nil {
		return ErrInvalidTransition
	}
	if t.Status != StatusDone && t.CompletedAt != nil {
		return ErrInvalidTransition
	}
	return nil
}

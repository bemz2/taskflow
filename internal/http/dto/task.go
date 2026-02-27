package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type ChangeStatusRequest struct {
	Status string `json:"status"`
}

type TaskResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

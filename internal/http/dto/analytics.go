package dto

import "time"

type TaskAnalyticsResponse struct {
	TasksCreated   int64     `json:"tasks_created"`
	TasksCompleted int64     `json:"tasks_completed"`
	TasksOpen      int64     `json:"tasks_open"`
	CompletionRate float64   `json:"completion_rate"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
}

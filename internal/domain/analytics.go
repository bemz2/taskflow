package domain

import "time"

type TaskAnalytics struct {
	TasksCreated   int64
	TasksCompleted int64
	UpdatedAt      time.Time
}

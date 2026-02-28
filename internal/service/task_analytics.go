package service

import (
	"context"
	"taskflow/internal/domain"

	"github.com/google/uuid"
)

type TaskAnalyticsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.TaskAnalytics, error)
}

type TaskAnalyticsService struct {
	repository TaskAnalyticsRepository
}

func NewTaskAnalyticsService(repository TaskAnalyticsRepository) *TaskAnalyticsService {
	return &TaskAnalyticsService{repository: repository}
}

func (s *TaskAnalyticsService) GetByUserID(ctx context.Context, userID uuid.UUID) (domain.TaskAnalytics, error) {
	return s.repository.GetByUserID(ctx, userID)
}

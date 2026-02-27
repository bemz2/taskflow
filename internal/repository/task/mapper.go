package task

import (
	"taskflow/internal/domain"
)

func toModel(t domain.Task) TaskModel {
	return TaskModel{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}
func toDomain(m TaskModel) (domain.Task, error) {
	return domain.NewTaskFromStorage(
		m.ID,
		m.UserID,
		m.Title,
		m.Description,
		domain.Status(m.Status),
		m.CreatedAt,
		m.CompletedAt,
	)
}

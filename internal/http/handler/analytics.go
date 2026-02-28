package handler

import (
	"net/http"
	"taskflow/internal/domain"
	"taskflow/internal/http/dto"
	middleware2 "taskflow/internal/http/middleware"
	"taskflow/internal/service"

	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	service *service.TaskAnalyticsService
}

func NewAnalyticsHandler(service *service.TaskAnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// Get godoc
// @Summary Get task analytics
// @Description Returns per-user task analytics produced by the Kafka worker and stored in task_analytics.
// @Tags analytics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.TaskAnalyticsResponse
// @Failure 401 {string} string "missing token, invalid token, or invalid auth context"
// @Failure 500 {string} string "unexpected server error"
// @Router /analytics [get]
func (h *AnalyticsHandler) Get(c echo.Context) error {
	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	analytics, err := h.service.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, toTaskAnalyticsResponse(analytics))
}

func toTaskAnalyticsResponse(analytics domain.TaskAnalytics) dto.TaskAnalyticsResponse {
	tasksOpen := analytics.TasksCreated - analytics.TasksCompleted
	if tasksOpen < 0 {
		tasksOpen = 0
	}

	completionRate := 0.0
	if analytics.TasksCreated > 0 {
		completionRate = float64(analytics.TasksCompleted) / float64(analytics.TasksCreated)
	}

	return dto.TaskAnalyticsResponse{
		TasksCreated:   analytics.TasksCreated,
		TasksCompleted: analytics.TasksCompleted,
		TasksOpen:      tasksOpen,
		CompletionRate: completionRate,
		LastUpdatedAt:  analytics.UpdatedAt,
	}
}

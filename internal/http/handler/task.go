package handler

import (
	"errors"
	"net/http"
	"strconv"
	middleware2 "taskflow/internal/http/middleware"
	"taskflow/internal/service"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"taskflow/internal/domain"
	"taskflow/internal/http/dto"
)

type TaskHandler struct {
	service   *service.TaskService
	analytics service.AnalyticsPublisher
}

func NewTaskHandler(taskService *service.TaskService, analytics service.AnalyticsPublisher) *TaskHandler {
	if analytics == nil {
		analytics = service.NewNoopAnalyticsPublisher()
	}

	return &TaskHandler{
		service:   taskService,
		analytics: analytics,
	}
}

// Create godoc
// @Summary Create task
// @Description Creates a task for the authenticated user.
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTaskRequest true "Task creation payload"
// @Success 201 {object} dto.TaskResponse
// @Failure 400 {string} string "invalid request or validation error"
// @Failure 401 {string} string "missing or invalid token"
// @Router /task [post]
func (h *TaskHandler) Create(c echo.Context) error {
	var req dto.CreateTaskRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	task, err := h.service.CreateTask(
		c.Request().Context(),
		userID,
		req.Title,
		req.Description,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	_ = h.analytics.PublishTaskEvent(c.Request().Context(), service.TaskEvent{
		Type:      service.TaskEventCreated,
		UserID:    userID,
		TaskID:    task.ID,
		CreatedAt: time.Now().UTC(),
	})

	return c.JSON(http.StatusCreated, toResponse(task))
}

// Get godoc
// @Summary Get task by ID
// @Description Returns a single task that belongs to the authenticated user. The service checks Redis before querying PostgreSQL.
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {string} string "invalid task id"
// @Failure 401 {string} string "missing or invalid token"
// @Failure 404 {string} string "task not found"
// @Router /task/{id} [get]
func (h *TaskHandler) Get(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid id")
	}

	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	task, err := h.service.GetTask(
		c.Request().Context(),
		userID,
		taskID,
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, "task not found")
	}

	return c.JSON(http.StatusOK, toResponse(task))
}

// ChangeStatus godoc
// @Summary Change task status
// @Description Updates the status of a task that belongs to the authenticated user.
// @Tags tasks
// @Accept json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param request body dto.ChangeStatusRequest true "Status update payload"
// @Success 204 "No Content"
// @Failure 400 {string} string "invalid request, invalid id, or invalid transition"
// @Failure 401 {string} string "missing or invalid token"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{id}/status [patch]
func (h *TaskHandler) ChangeStatus(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid id")
	}

	var req dto.ChangeStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	status := domain.NormalizeStatus(domain.Status(req.Status))

	if err := h.service.ChangeStatus(
		c.Request().Context(),
		userID,
		taskID,
		status,
	); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if status == domain.StatusDone {
		_ = h.analytics.PublishTaskEvent(c.Request().Context(), service.TaskEvent{
			Type:      service.TaskEventCompleted,
			UserID:    userID,
			TaskID:    taskID,
			CreatedAt: time.Now().UTC(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// Delete godoc
// @Summary Delete task
// @Description Deletes a task that belongs to the authenticated user.
// @Tags tasks
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {string} string "invalid task id"
// @Failure 401 {string} string "missing or invalid token"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid id")
	}

	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	if err := h.service.DeleteTask(c.Request().Context(), userID, taskID); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, "task not found")
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	_ = h.analytics.PublishTaskEvent(c.Request().Context(), service.TaskEvent{
		Type:      service.TaskEventDeleted,
		UserID:    userID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	})

	return c.NoContent(http.StatusNoContent)
}

// List godoc
// @Summary List tasks
// @Description Returns tasks for the authenticated user with optional pagination, filtering, search, and sorting.
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of tasks to return"
// @Param offset query int false "Pagination offset"
// @Param status query string false "Task status filter" Enums(pending,in_progress,done,canceled)
// @Param search query string false "Case-insensitive title search"
// @Param sort_by query string false "Sort column"
// @Param sort_dir query string false "Sort direction" Enums(asc,desc)
// @Success 200 {array} dto.TaskResponse
// @Failure 401 {string} string "missing or invalid token"
// @Failure 500 {string} string "unexpected server error"
// @Router /tasks [get]
func (h *TaskHandler) List(c echo.Context) error {
	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

	var filter domain.TaskFilter

	// limit
	if limit := c.QueryParam("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			filter.Limit = v
		}
	}

	// offset
	if offset := c.QueryParam("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil {
			filter.Offset = v
		}
	}

	// status
	if status := c.QueryParam("status"); status != "" {
		s := domain.NormalizeStatus(domain.Status(status))
		filter.Status = &s
	}

	// search
	if search := c.QueryParam("search"); search != "" {
		filter.Search = &search
	}

	filter.SortBy = c.QueryParam("sort_by")
	filter.SortDir = c.QueryParam("sort_dir")

	tasks, err := h.service.ListTasks(
		c.Request().Context(),
		userID,
		filter,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	resp := make([]dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, toResponse(t))
	}

	return c.JSON(http.StatusOK, resp)
}

func toResponse(t domain.Task) dto.TaskResponse {
	return dto.TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}

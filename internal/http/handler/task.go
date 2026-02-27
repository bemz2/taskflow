package handler

import (
	"net/http"
	"strconv"
	"taskflow/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"taskflow/internal/domain"
	"taskflow/internal/http/dto"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) Create(c echo.Context) error {
	var req dto.CreateTaskRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	userID := c.Get("userID").(uuid.UUID)

	task, err := h.service.CreateTask(
		c.Request().Context(),
		userID,
		req.Title,
		req.Description,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, toResponse(task))
}

func (h *TaskHandler) Get(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid id")
	}

	userID := c.Get("userID").(uuid.UUID)

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

	userID := c.Get("userID").(uuid.UUID)

	status := domain.Status(req.Status)

	if err := h.service.ChangeStatus(
		c.Request().Context(),
		userID,
		taskID,
		status,
	); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
func (h *TaskHandler) List(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

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
		s := domain.Status(status)
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

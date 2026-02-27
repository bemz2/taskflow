package handler

import (
	"net/http"
	"taskflow/internal/domain"
	"taskflow/internal/http/dto"
	"taskflow/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	user, err := h.service.CreateUser(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) Me(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	user, err := h.service.GetUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

func toUserResponse(user domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

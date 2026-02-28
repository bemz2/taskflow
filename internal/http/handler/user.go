package handler

import (
	"net/http"
	"taskflow/internal/domain"
	"taskflow/internal/http/dto"
	middleware2 "taskflow/internal/http/middleware"
	"taskflow/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Create godoc
// @Summary Create user
// @Description Creates a new user account without issuing a token.
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User creation payload"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {string} string "invalid request or validation error"
// @Router /users [post]
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

// Me godoc
// @Summary Get current user
// @Description Returns the currently authenticated user.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {string} string "missing or invalid token"
// @Failure 404 {string} string "user not found"
// @Router /me [get]
func (h *UserHandler) Me(c echo.Context) error {
	userID, ok := middleware2.UserIDFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "invalid auth context")
	}

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

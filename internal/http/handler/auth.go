package handler

import (
	"net/http"
	"taskflow/internal/http/dto"
	"taskflow/internal/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(authService *service.AuthService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	return h.authenticate(c, true)
}

func (h *AuthHandler) Login(c echo.Context) error {
	return h.authenticate(c, false)
}

func (h *AuthHandler) authenticate(c echo.Context, register bool) error {
	var req dto.AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request")
	}

	var (
		token string
		err   error
	)

	if register {
		token, err = h.authService.Register(c.Request().Context(), req.Email, req.Password)
	} else {
		token, err = h.authService.Login(c.Request().Context(), req.Email, req.Password)
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	user, err := h.userService.GetUserByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	})
}

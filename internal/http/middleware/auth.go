package middleware

import (
	"net/http"
	"strings"
	"taskflow/internal/service"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(tokenService *service.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, "missing bearer token")
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			userID, err := tokenService.Parse(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, "invalid token")
			}

			c.Set("userID", userID)
			return next(c)
		}
	}
}

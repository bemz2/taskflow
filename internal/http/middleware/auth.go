package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func DevUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		c.Set("userID", userID)
		return next(c)
	}
}

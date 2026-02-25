package application

import (
	"log/slog"
	"taskflow/internal"

	"github.com/labstack/echo/v5"
)

const APIV1Version = "/api/v1"

type PublicServer struct {
	cfg    internal.AppConfig
	echo   *echo.Echo
	logger slog.Logger
}

func NewPublicServer(cfg)

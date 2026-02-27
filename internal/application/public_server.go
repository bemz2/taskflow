package application

import (
	"context"
	"errors"
	"fmt"
	"taskflow/internal"
	middleware2 "taskflow/internal/http/middleware"
	"taskflow/internal/lib/logger/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const APIV1Version = "/api/v1"

type PublicServer struct {
	cfg    internal.AppConfig
	echo   *echo.Echo
	logger logger.Logger
}

func NewPublicServer(cfg internal.AppConfig, logger logger.Logger) *PublicServer {
	return &PublicServer{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *PublicServer) Configure(container *Container) (*PublicServer, error) {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogError:    true,
		LogLatency:  true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			s.logger.Info("http_request,",
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"error", v.Error,
			)
			return nil
		},
	}))
	s.echo = e
	s.v1(container)

	return s, nil
}

func (s *PublicServer) Start() error {
	if s.echo == nil {
		return errors.New("echo is not initialized")
	}
	address := fmt.Sprintf(":%s", s.cfg.PublicServerConfig.Port)
	return s.echo.Start(address)
}

func (s *PublicServer) ShutDown(ctx context.Context) error {
	if s.echo == nil {
		return errors.New("echo is not initialized")
	}
	return s.echo.Shutdown(ctx)
}

func (s *PublicServer) Echo() *echo.Echo {
	return s.echo
}

func (s *PublicServer) v1(container *Container) {
	v1 := s.echo.Group(APIV1Version)

	registerHandler := container.AuthHandler.Register
	loginHandler := container.AuthHandler.Login
	createUserHandler := container.UserHandler.Create
	meHandler := container.UserHandler.Me
	listTaskHandler := container.TaskHandler.List
	createTaskHandler := container.TaskHandler.Create
	getTaskHandler := container.TaskHandler.Get
	changeTaskStatusHandler := container.TaskHandler.ChangeStatus
	deleteTaskHandler := container.TaskHandler.Delete

	authM := middleware2.AuthMiddleware(container.TokenService)

	v1.POST("/auth/register", registerHandler)
	v1.POST("/auth/login", loginHandler)
	v1.POST("/users", createUserHandler)
	v1.GET("/me", meHandler, authM)
	v1.GET("/tasks", listTaskHandler, authM)
	v1.POST("/task", createTaskHandler, authM)
	v1.GET("/task/:id", getTaskHandler, authM)
	v1.PATCH("/tasks/:id/status", changeTaskStatusHandler, authM)
	v1.DELETE("/tasks/:id", deleteTaskHandler, authM)
}

package application

import (
	"context"
	"taskflow/internal"
	"taskflow/internal/client/postgres"
	"taskflow/internal/http/handler"
	"taskflow/internal/lib/logger/logger"
	"taskflow/internal/repository/task"
	userrepo "taskflow/internal/repository/user"
	"taskflow/internal/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Ctx    context.Context
	Config *internal.AppConfig
	Logger logger.Logger

	Pool *pgxpool.Pool

	TokenService *service.TokenService

	UserRepo    *userrepo.UserRepository
	UserService *service.UserService
	UserHandler *handler.UserHandler

	AuthService *service.AuthService
	AuthHandler *handler.AuthHandler

	TaskRepo    *task.TaskRepository
	TaskService *service.TaskService
	TaskHandler *handler.TaskHandler
}

func NewContainer(ctx context.Context, config internal.AppConfig) *Container {
	return &Container{
		Ctx:    ctx,
		Config: &config,
	}
}

func (c *Container) Init(ctx context.Context) (*Container, error) {
	c.Logger = logger.NewSlogLogger()
	pool, err := postgres.NewPool(ctx, c.Config.PostgresConfig)
	if err != nil {
		return c, err
	}
	c.Pool = pool
	c.TokenService = service.NewTokenService(
		c.Config.AuthConfig.JWTSecret,
		time.Duration(c.Config.AuthConfig.JWTExpirationHours)*time.Hour,
	)

	c.UserRepo = userrepo.NewUserRepository(c.Pool)
	c.UserService = service.NewUserService(c.UserRepo)
	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.AuthService = service.NewAuthService(c.UserService, c.TokenService)
	c.AuthHandler = handler.NewAuthHandler(c.AuthService, c.UserService)
	if err := c.UserService.EnsureDevUser(ctx); err != nil {
		return c, err
	}

	c.TaskRepo = task.NewTaskRepository(c.Pool)
	c.TaskService = service.NewTaskService(c.TaskRepo)
	c.TaskHandler = handler.NewTaskHandler(c.TaskService)

	return c, nil
}

func (c *Container) Close() error {
	if c.Pool != nil {
		c.Pool.Close()
	}
	return nil
}

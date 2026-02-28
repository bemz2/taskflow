package application

import (
	"context"
	"taskflow/internal"
	kafka2 "taskflow/internal/client/kafka"
	"taskflow/internal/client/postgres"
	redis2 "taskflow/internal/client/redis"
	"taskflow/internal/http/handler"
	"taskflow/internal/lib/logger/logger"
	analyticsrepo "taskflow/internal/repository/analytics"
	"taskflow/internal/repository/task"
	userrepo "taskflow/internal/repository/user"
	"taskflow/internal/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	Ctx    context.Context
	Config *internal.AppConfig
	Logger logger.Logger

	Pool  *pgxpool.Pool
	Redis *redis.Client

	TokenService  *service.TokenService
	Analytics     service.AnalyticsPublisher
	AnalyticsRepo *analyticsrepo.Repository

	UserRepo    *userrepo.UserRepository
	UserService *service.UserService
	UserHandler *handler.UserHandler

	AuthService *service.AuthService
	AuthHandler *handler.AuthHandler

	TaskRepo    *task.TaskRepository
	TaskService *service.TaskService
	TaskHandler *handler.TaskHandler

	TaskAnalyticsService *service.TaskAnalyticsService
	AnalyticsHandler     *handler.AnalyticsHandler
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

	rdb, err := redis2.NewClient(c.Config.RedisConfig)
	if err != nil {
		return nil, err
	}
	c.Redis = rdb

	c.TokenService = service.NewTokenService(
		c.Config.AuthConfig.JWTSecret,
		time.Duration(c.Config.AuthConfig.JWTExpirationHours)*time.Hour,
	)
	c.Analytics = service.NewKafkaAnalyticsPublisher(kafka2.NewWriter(c.Config.KafkaConfig))

	c.UserRepo = userrepo.NewUserRepository(c.Pool)
	c.UserService = service.NewUserService(c.UserRepo)
	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.AuthService = service.NewAuthService(c.UserService, c.TokenService)
	c.AuthHandler = handler.NewAuthHandler(c.AuthService, c.UserService)
	if err := c.UserService.EnsureDevUser(ctx); err != nil {
		return c, err
	}

	c.TaskRepo = task.NewTaskRepository(c.Pool)
	c.TaskService = service.NewTaskService(c.TaskRepo, service.NewRedisTaskCache(c.Redis))
	c.TaskHandler = handler.NewTaskHandler(c.TaskService, c.Analytics)
	c.AnalyticsRepo = analyticsrepo.NewRepository(c.Pool)
	c.TaskAnalyticsService = service.NewTaskAnalyticsService(c.AnalyticsRepo)
	c.AnalyticsHandler = handler.NewAnalyticsHandler(c.TaskAnalyticsService)

	return c, nil
}

func (c *Container) Close() error {
	if c.Pool != nil {
		c.Pool.Close()
	}

	if c.Redis != nil {
		c.Redis.Close()
	}
	if c.Analytics != nil {
		_ = c.Analytics.Close()
	}
	return nil
}

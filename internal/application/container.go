package application

import (
	"context"
	"taskflow/internal"
	"taskflow/internal/client/postgres"
	"taskflow/internal/lib/logger/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Ctx    context.Context
	Config *internal.AppConfig
	Logger logger.Logger

	Pool *pgxpool.Pool
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

	return c, nil
}

func (c *Container) Close() error {
	if c.Pool != nil {
		c.Pool.Close()
	}
	return nil
}

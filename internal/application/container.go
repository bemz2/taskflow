package application

import (
	"context"
	"taskflow/internal"
	"taskflow/internal/client/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Config *internal.AppConfig

	Pool *pgxpool.Pool
}

func NewContainer(config *internal.AppConfig) *Container {
	return &Container{
		Config: config,
	}
}

func (c *Container) Init(ctx context.Context) error {
	pool, err := postgres.NewPool(ctx, c.Config.PostgresConfig)
	if err != nil {
		return err
	}
	c.Pool = pool

	return nil
}

func (c *Container) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}

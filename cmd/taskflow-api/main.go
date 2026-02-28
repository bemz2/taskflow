package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"taskflow/internal"
	"taskflow/internal/application"
	"time"

	_ "taskflow/docs"
)

// @title Taskflow API
// @version 1.0
// @description Task management HTTP API with JWT authentication, PostgreSQL persistence, and Redis-backed task caching.
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use the format: Bearer <token>
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := internal.NewConfig[internal.AppConfig](".env")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	container, err := application.NewContainer(ctx, cfg).Init(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to init container"))
	}

	publicServer, err := application.NewPublicServer(cfg, container.Logger).Configure(container)
	if err != nil {
		panic(fmt.Errorf("failed to configure public server: %w", err))
	}

	app := application.NewApp(publicServer, container)

	if err = app.Run(ctx); err != nil {
		container.Logger.Error(fmt.Sprintf("failed to run app: %v", err))
	}
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = app.ShutDown(ctx); err != nil {
		container.Logger.Error(fmt.Sprintf("failed to shutdown app: %v", err))
	}

}

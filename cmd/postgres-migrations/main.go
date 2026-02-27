package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appinternal "taskflow/internal"
)

func main() {
	up := flag.Bool("up", false, "run up migrations")
	flag.Parse()

	if *up {
		if err := runMigrations(); err != nil {
			log.Fatal(err)
		}
	}
}

func runMigrations() error {
	cfg, err := appinternal.NewConfig[appinternal.AppConfig](".env")
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresConfig.DBAdapter,
		cfg.PostgresConfig.DBUser,
		cfg.PostgresConfig.DBPassword,
		cfg.PostgresConfig.DBHost,
		cfg.PostgresConfig.DBPort,
		cfg.PostgresConfig.DBName,
		cfg.PostgresConfig.DBSSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	files, err := filepath.Glob(filepath.Join("migrations", "postgres-migrations", "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}

	return nil
}

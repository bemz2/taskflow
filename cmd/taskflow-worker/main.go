package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	appinternal "taskflow/internal"
	kafkaclient "taskflow/internal/client/kafka"
	"taskflow/internal/client/postgres"
	"taskflow/internal/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := appinternal.NewConfig[appinternal.AppConfig](".env")
	if err != nil {
		log.Fatal(fmt.Errorf("load config: %w", err))
	}

	pool, err := postgres.NewPool(ctx, cfg.PostgresConfig)
	if err != nil {
		log.Fatal(fmt.Errorf("connect postgres: %w", err))
	}
	defer pool.Close()

	reader := kafkaclient.NewReader(cfg.KafkaConfig)
	defer reader.Close()

	log.Printf("taskflow-worker started: broker=%s topic=%s group=%s", kafkaclient.BrokerAddress(cfg.KafkaConfig), cfg.KafkaConfig.Topic, cfg.KafkaConfig.AnalyticsGroupID)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch kafka message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if err := applyTaskEvent(ctx, pool, msg.Value); err != nil {
			log.Printf("process analytics event: %v", err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("commit kafka message: %v", err)
		}
	}

	log.Print("taskflow-worker stopped")
}

func applyTaskEvent(ctx context.Context, db *pgxpool.Pool, payload []byte) error {
	var event service.TaskEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}

	switch event.Type {
	case service.TaskEventCreated:
		_, err := db.Exec(ctx, `
			INSERT INTO task_analytics (user_id, tasks_created, tasks_completed, updated_at)
			VALUES ($1, 1, 0, now())
			ON CONFLICT (user_id) DO UPDATE
			SET tasks_created = task_analytics.tasks_created + 1,
			    updated_at = now()
		`, event.UserID)
		return err
	case service.TaskEventCompleted:
		_, err := db.Exec(ctx, `
			INSERT INTO task_analytics (user_id, tasks_created, tasks_completed, updated_at)
			VALUES ($1, 0, 1, now())
			ON CONFLICT (user_id) DO UPDATE
			SET tasks_completed = task_analytics.tasks_completed + 1,
			    updated_at = now()
		`, event.UserID)
		return err
	case service.TaskEventDeleted:
		_, err := db.Exec(ctx, `
			INSERT INTO task_analytics (user_id, tasks_created, tasks_completed, updated_at)
			VALUES ($1, 0, 0, now())
			ON CONFLICT (user_id) DO UPDATE
			SET updated_at = now()
		`, event.UserID)
		return err
	default:
		return fmt.Errorf("unknown task event type: %s", event.Type)
	}
}

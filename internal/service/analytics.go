package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

type TaskEventType string

const (
	TaskEventCreated   TaskEventType = "task_created"
	TaskEventCompleted TaskEventType = "task_completed"
	TaskEventDeleted   TaskEventType = "task_deleted"
)

type TaskEvent struct {
	Type      TaskEventType `json:"type"`
	UserID    uuid.UUID     `json:"user_id"`
	TaskID    uuid.UUID     `json:"task_id"`
	CreatedAt time.Time     `json:"created_at"`
}

type AnalyticsPublisher interface {
	PublishTaskEvent(ctx context.Context, event TaskEvent) error
	Close() error
}

type NoopAnalyticsPublisher struct{}

type KafkaAnalyticsPublisher struct {
	writer *kafkago.Writer
}

func NewNoopAnalyticsPublisher() AnalyticsPublisher {
	return NoopAnalyticsPublisher{}
}

func (NoopAnalyticsPublisher) PublishTaskEvent(context.Context, TaskEvent) error {
	return nil
}

func (NoopAnalyticsPublisher) Close() error {
	return nil
}

func NewKafkaAnalyticsPublisher(writer *kafkago.Writer) AnalyticsPublisher {
	if writer == nil {
		return NewNoopAnalyticsPublisher()
	}

	return &KafkaAnalyticsPublisher{writer: writer}
}

func (p *KafkaAnalyticsPublisher) PublishTaskEvent(ctx context.Context, event TaskEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(event.UserID.String()),
		Value: payload,
		Time:  event.CreatedAt,
	})
}

func (p *KafkaAnalyticsPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}

	return p.writer.Close()
}

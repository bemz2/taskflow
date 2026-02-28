package kafka

import (
	"fmt"
	"taskflow/internal"

	kafkago "github.com/segmentio/kafka-go"
)

func BrokerAddress(cfg internal.KafkaConfig) string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func NewWriter(cfg internal.KafkaConfig) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(BrokerAddress(cfg)),
		Topic:        cfg.Topic,
		RequiredAcks: kafkago.RequireOne,
		Balancer:     &kafkago.LeastBytes{},
	}
}

func NewReader(cfg internal.KafkaConfig) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{BrokerAddress(cfg)},
		Topic:   cfg.Topic,
		GroupID: cfg.AnalyticsGroupID,
	})
}

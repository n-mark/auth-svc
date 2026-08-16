package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"auth-service/internal/config"
	"auth-service/internal/model"
)

// KafkaImpl is the Kafka counterpart of RabbitImpl.
//
// Mapping from the RabbitMQ topology: the `auth` exchange becomes the `auth`
// topic, and the routing key (`user.created`) is carried inside the payload in
// `eventType`. Consumers filter on that field instead of relying on broker-side
// wildcard bindings, which Kafka does not have.
type KafkaImpl struct {
	cfg    config.KafkaConfig
	writer *kafka.Writer
}

func NewKafkaImpl(cfg config.KafkaConfig) *KafkaImpl {
	w := &kafka.Writer{
		Addr: kafka.TCP(cfg.Brokers...),
		// Hash on the message key so all events of one user land in the same
		// partition and keep their relative order.
		Balancer: &kafka.Hash{},
		// acks=all: combined with min.insync.replicas=2 on the topic this
		// guarantees the write survives the loss of a single broker.
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		BatchTimeout:           50 * time.Millisecond,
	}
	return &KafkaImpl{cfg: cfg, writer: w}
}

func (k *KafkaImpl) ReportUserCreated(e model.UserCreatedEvent) error {
	// The event type used to be the AMQP routing key; on Kafka it must travel
	// with the message itself.
	if e.EventType == "" {
		e.EventType = k.cfg.UserCreatedEventType
	}

	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal user.created: %w", err)
	}

	slog.Info("publishing user.created",
		"topic", k.cfg.AuthTopic, "event_id", e.EventId, "event_type", e.EventType)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: k.cfg.AuthTopic,
		Key:   []byte(e.Email),
		Value: body,
	})
}

// Run blocks to match the Broker contract. This service is publish-only, so
// there is nothing to consume; we just keep the writer alive.
func (k *KafkaImpl) Run() {
	defer k.writer.Close()
	slog.Info("kafka broker ready (producer only)", "topic", k.cfg.AuthTopic)
	select {}
}

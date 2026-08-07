package messaging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"minimal-service/internal/config"
	"minimal-service/internal/model"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitImpl struct {
	conn          *amqp.Connection
	publisher     *amqp.Channel
	publisherLock sync.Mutex
	cfg           config.RabbitConfig
}

func NewRabbitImpl(cfg config.RabbitConfig) (*RabbitImpl, error) {
	conn, err := amqp.Dial(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	publisher, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}

	r := &RabbitImpl{conn: conn, cfg: cfg, publisher: publisher}

	return r, nil
}

func (r *RabbitImpl) declareExchange(ch *amqp.Channel, exchange string) error {
	return ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
}

func (r *RabbitImpl) produceAuthEvent(routingKey, messageId string, body []byte) error {
	r.publisherLock.Lock()
	defer r.publisherLock.Unlock()

	return r.publisher.Publish(
		r.cfg.AuthExchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    messageId,
			Body:         body,
		},
	)
}

func (r *RabbitImpl) ReportUserCreated(e model.UserCreatedEvent) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal user.created: %w", err)
	}
	slog.Info("publishing user.created", "event_id", e.EventId, "rk", r.cfg.ProduceRK)
	return r.produceAuthEvent(r.cfg.ProduceRK, e.EventId.String(), body)
}

func (r *RabbitImpl) Run() {
	defer r.conn.Close()
	defer r.publisher.Close()

	if err := r.declareExchange(r.publisher, r.cfg.AuthExchange); err != nil {
		slog.Error("declare topology", "op", "exchange", "err", err)
		return
	}

	select {}
}

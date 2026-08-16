package messaging

import (
	"fmt"
	"auth-service/internal/config"
	"auth-service/internal/model"
)

type Broker interface {
	Run()
	ReportUserCreated(e model.UserCreatedEvent) error
}

func InitBroker(cfg config.Config) (Broker, error) {
	switch cfg.BrokerType {
	case "RABBITMQ":
		br, err := NewRabbitImpl(config.GetRabbitConfig())
		if err != nil {
			return nil, fmt.Errorf("can't init rabbitmq impl: %s", err)
		}
		return br, nil
	case "KAFKA":
		return NewKafkaImpl(config.GetKafkaConfig()), nil
	default:
		return nil, fmt.Errorf("unsupported BROKER_TYPE %q (supported: RABBITMQ, KAFKA)", cfg.BrokerType)
	}
}

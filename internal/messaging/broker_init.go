package messaging

import (
	"fmt"
	"minimal-service/internal/config"
	"minimal-service/internal/model"
)

type Broker interface {
	Run()
	ReportUserCreated(e model.UserCreatedEvent) error
}

func InitBroker(cfg config.Config) (Broker, error) {
	var b Broker

	if "RABBITMQ" == cfg.BrokerType {
		br, err := NewRabbitImpl(config.GetRabbitConfig())
		if err != nil {
			return nil, fmt.Errorf("can't init rabbitmq impl: %s", err)
		}
		b = br
	}

	return b, nil
}

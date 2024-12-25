package saramax

import "github.com/IBM/sarama"

type Consumer interface {
	Start() error
}

type HandlerFunc[T any] func(msg *sarama.ConsumerMessage, event T) error

type BatchHandlerFunc[T any] func(msg []*sarama.ConsumerMessage, event []T) error

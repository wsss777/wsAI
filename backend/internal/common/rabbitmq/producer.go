package rabbitmq

import (
	"fmt"
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func (r *RabbitMQ) PublishWork(message []byte) error {
	err := r.publishWorkOnce(message)
	if err == nil {
		return nil
	}

	logger.L().Warn("RabbitMQ publish failed, reconnecting and retrying once",
		zap.Error(err),
		zap.String("queue", r.Key),
	)
	r.reconnect()
	return r.publishWorkOnce(message)
}

func (r *RabbitMQ) publishWorkOnce(message []byte) error {
	if r.channel == nil {
		return fmt.Errorf("channel is nil")
	}

	_, err := r.channel.QueueDeclare(
		r.Key,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.L().Error("RabbitMQ queue declare failed",
			zap.Error(err),
			zap.String("queue", r.Key),
		)
		return err
	}

	err = r.channel.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			Body:         message,
		},
	)
	if err != nil {
		logger.L().Error("RabbitMQ publish message failed",
			zap.Error(err),
			zap.String("queue", r.Key),
			zap.ByteString("message", message),
		)
		return err
	}

	return nil
}

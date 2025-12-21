package rabbitmq // 同包

import (
	"fmt"
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// PublishWork 发送消息（Work 模式）
func (r *RabbitMQ) PublishWork(message []byte) error {
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

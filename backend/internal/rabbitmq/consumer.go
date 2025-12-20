package rabbitmq

import (
	"fmt"
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// ConsumeWork 启动消费者（Work 模式）
func (r *RabbitMQ) ConsumeWork(handle func(msg *amqp.Delivery) error) error {
	if r.channel == nil {
		return fmt.Errorf("channel is nil")
	}

	q, err := r.channel.QueueDeclare(
		r.Key, false, false, false, false, nil,
	)
	if err != nil {
		logger.L().Error("RabbitMQ queue declare failed in consume",
			zap.Error(err),
			zap.String("queue", r.Key),
		)
		return err
	}

	msgs, err := r.channel.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		logger.L().Error("RabbitMQ register consumer failed",
			zap.Error(err),
			zap.String("queue", r.Key),
		)
		return err
	}

	logger.L().Info("RabbitMQ consumer started",
		zap.String("queue", r.Key),
	)

	for msg := range msgs {
		if err := handle(&msg); err != nil {
			logger.L().Error("RabbitMQ message handle failed",
				zap.Error(err),
				zap.ByteString("body", msg.Body),
				zap.String("queue", r.Key),
			)
		}
	}

	return nil
}

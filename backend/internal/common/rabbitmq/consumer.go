package rabbitmq

import (
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func (r *RabbitMQ) ConsumeWork(handle func(msg *amqp.Delivery) error) error {
	for {
		if r.channel == nil {
			r.reconnect()
			continue
		}

		if err := r.channel.Qos(1, 0, false); err != nil {
			logger.L().Error("RabbitMQ set QoS failed", zap.Error(err))
			r.reconnect()
			continue
		}

		msgs, err := r.channel.Consume(
			r.queueName, "", false, false, false, false, nil,
		)
		if err != nil {
			logger.L().Error("RabbitMQ register consumer failed",
				zap.Error(err),
				zap.String("queue", r.queueName),
			)
			r.reconnect()
			continue
		}

		logger.L().Info("RabbitMQ consumer started",
			zap.String("queue", r.queueName),
		)

		for msg := range msgs {
			if err := handle(&msg); err != nil {
				if nackErr := msg.Nack(false, true); nackErr != nil {
					logger.L().Error("RabbitMQ nack failed",
						zap.Error(nackErr),
						zap.String("queue", r.queueName),
						zap.Uint64("delivery_tag", msg.DeliveryTag),
					)
				}
				logger.L().Error("RabbitMQ message handle failed",
					zap.Error(err),
					zap.ByteString("body", msg.Body),
					zap.String("queue", r.queueName),
				)
				continue
			}

			if ackErr := msg.Ack(false); ackErr != nil {
				logger.L().Error("RabbitMQ ack failed",
					zap.Error(ackErr),
					zap.String("queue", r.queueName),
					zap.Uint64("delivery_tag", msg.DeliveryTag),
				)
			}
		}

		logger.L().Warn("RabbitMQ channel closed, reconnecting...")
		r.reconnect()
	}
}

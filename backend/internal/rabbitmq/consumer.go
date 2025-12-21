package rabbitmq

import (
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// ConsumeWork 启动消费者（Work 模式）
func (r *RabbitMQ) ConsumeWork(handle func(msg *amqp.Delivery) error) error {
	for {

		//if r.channel == nil {
		//	return fmt.Errorf("channel is nil")
		//}
		if err := r.channel.Qos(1, 0, false); err != nil {
			logger.L().Error("RabbitMQ set QoS failed", zap.Error(err))
			r.reconnect()
			continue
		}

		//q, err := r.channel.QueueDeclare(
		//	r.queueName, false, false, false, false, nil,
		//)
		//if err != nil {
		//	logger.L().Error("RabbitMQ queue declare failed in consume",
		//		zap.Error(err),
		//		zap.String("queue", r.queueName),
		//	)
		//	r.reconnect()
		//	continue
		//}

		msgs, err := r.channel.Consume(
			r.queueName, "", false, false, false, false, nil,
		)
		if err != nil {
			logger.L().Error("RabbitMQ register consumer failed",
				zap.Error(err),
				zap.String("queue", r.queueName),
			)
			return err
		}

		logger.L().Info("RabbitMQ consumer started",
			zap.String("queue", r.queueName),
		)

		for msg := range msgs {
			if err := handle(&msg); err != nil {
				logger.L().Error("RabbitMQ message handle failed",
					zap.Error(err),
					zap.ByteString("body", msg.Body),
					zap.String("queue", r.queueName),
				)
			}
		}
		logger.L().Warn("RabbitMQ channel closed, reconnecting...")
		r.reconnect()
	}
}

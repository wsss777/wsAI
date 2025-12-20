package rabbitmq // 包名 rabbitmq

import (
	"fmt"
	"wsai/backend/config"
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// conn 全局连接
var conn *amqp.Connection

// InitConn 初始化 RabbitMQ 连接
func InitConn() error {
	c := config.C.RabbitmqConfig
	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Vhost)

	logger.L().Info("RabbitMQ connecting",
		zap.String("host", c.Host),
		zap.Int("port", c.Port),
		zap.String("vhost", c.Vhost),
		zap.String("user", c.Username),
	)

	var err error
	conn, err = amqp.Dial(mqURL)
	if err != nil {
		logger.L().Fatal("RabbitMQ connection failed",
			zap.Error(err),
			zap.String("host", c.Host),
			zap.Int("port", c.Port),
			zap.String("vhost", c.Vhost),
		)
	}

	logger.L().Info("RabbitMQ connected successfully")
	return nil
}

// CloseConn 关闭全局连接
func CloseConn() error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

type RabbitMQ struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	Exchange string
	Key      string
}

// New 创建基础实例
func NewRabbitMQ(exchange, key string) *RabbitMQ {
	return &RabbitMQ{
		Exchange: exchange,
		Key:      key,
	}
}

// Destroy 关闭 channel 和 connection
func (r *RabbitMQ) Destroy() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			logger.L().Error("RabbitMQ channel close failed",
				zap.Error(err),
				zap.String("exchange", r.Exchange),
				zap.String("key", r.Key),
			)
		}
	}
}

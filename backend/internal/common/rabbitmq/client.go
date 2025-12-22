package rabbitmq // 包名 rabbitmq

import (
	"fmt"
	"sync"
	"wsai/backend/config"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/service/chatMessage"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// conn 全局连接
var conn *amqp.Connection

var RMQMessage *RabbitMQ
var connMu sync.Mutex
var once sync.Once

func InitRabbitMQ() {
	RMQMessage = NewWorkRabbitMQ("Message")
	go RMQMessage.ConsumeWork(chatMessage.ProcessMessageDelivery)
}
func DestroyRabbitMQ() {
	if RMQMessage != nil {
		RMQMessage.Destroy()
	}
}

// initConn 初始化 RabbitMQ 连接
func initConn() error {
	once.Do(func() {
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
	})
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
	conn      *amqp.Connection
	channel   *amqp.Channel
	Exchange  string
	Key       string
	queueName string
}

// New 创建基础实例
func NewRabbitMQ(exchange, key string) *RabbitMQ {
	return &RabbitMQ{
		Exchange:  exchange,
		Key:       key,
		queueName: key,
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

func NewWorkRabbitMQ(queue string) *RabbitMQ {
	rabbitmq := NewRabbitMQ("", queue)
	rabbitmq.queueName = queue
	if conn == nil {
		_ = initConn()
	}
	rabbitmq.conn = conn

	var err error
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	_, err = rabbitmq.channel.QueueDeclare(queue,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.L().Fatal("RabbitMQ channel failed",
			zap.Error(err),
			zap.String("queue", queue),
		)
	}
	return rabbitmq
}

func (r *RabbitMQ) reconnect() {
	if r.channel != nil {
		_ = r.channel.Close()
		r.channel = nil
	}

	// 如果全局连接断了，重新建立
	if conn == nil || conn.IsClosed() {
		conn = nil
		_ = initConn()
		r.conn = conn
	}

	ch, err := r.conn.Channel()
	if err != nil {
		logger.L().Error("Reconnect: create channel failed", zap.Error(err))
		return
	}
	r.channel = ch

	// 重新声明持久化队列
	_, err = ch.QueueDeclare(r.queueName, true, false, false, false, nil)
	if err != nil {
		logger.L().Error("Reconnect: queue declare failed", zap.Error(err))
	}
}

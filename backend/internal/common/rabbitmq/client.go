package rabbitmq

import (
	"fmt"
	"sync"
	"wsai/backend/config"
	"wsai/backend/internal/logger"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var conn *amqp.Connection

var RMQMessage *RabbitMQ
var RMQMessageConsumer *RabbitMQ
var connMu sync.Mutex
var once sync.Once

func InitRabbitMQ() {
	RMQMessage = NewWorkRabbitMQ("Message")
	RMQMessageConsumer = NewWorkRabbitMQ("Message")
	go RMQMessageConsumer.ConsumeWork(ProcessMessageDelivery)
}

func DestroyRabbitMQ() {
	if RMQMessage != nil {
		RMQMessage.Destroy()
	}
	if RMQMessageConsumer != nil {
		RMQMessageConsumer.Destroy()
	}
}

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

func NewRabbitMQ(exchange, key string) *RabbitMQ {
	return &RabbitMQ{
		Exchange:  exchange,
		Key:       key,
		queueName: key,
	}
}

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

	_, err = ch.QueueDeclare(r.queueName, true, false, false, false, nil)
	if err != nil {
		logger.L().Error("Reconnect: queue declare failed", zap.Error(err))
	}
}

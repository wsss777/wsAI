package rabbitmq

import (
	"fmt"
	"log"
	"wsai/backend/config"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection

func initConn() {
	c := config.C.RabbitmqConfig
	mqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Vhost)
	log.Println("mqURL is:", mqURL)
	var err error
	conn, err = amqp.Dial(mqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

}

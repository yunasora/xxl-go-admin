package mq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

func Connect(url, queueName string) error {
	var err error
	Conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	Channel, err = Conn.Channel()

	_, err = Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[MQ] 链接成功,	队列: %s", queueName)
	return nil
}

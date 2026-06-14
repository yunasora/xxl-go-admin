package mq

import (
	"encoding/json"
	"go-xxl-admin/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishTask(queueName string, task models.TriggerParam) error {
	//Struct-->JSON
	jsonData, err := json.Marshal(task)
	if err != nil {
		return err
	}

	//publish to queue
	err = Channel.Publish(
		"",        // exchange: 空 = 用默认 exchange
		queueName, // routing key = 队列名
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 持久化，重启不丢
			ContentType:  "application/json",
			Body:         jsonData,
		},
	)
	return err
}

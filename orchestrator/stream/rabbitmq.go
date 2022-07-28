package stream

import (
	"context"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange = "pitwall.orchestration"
)

type MsgHander func(msg []byte)

func Send(ctx context.Context, routingKey string, content []byte) error {
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	defer ch.Close()

	err = ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        content,
		})
	if err != nil {
		return err
	}

	return nil
}

func RegisterHandler(queueName string, routingKey string, handler MsgHander) error {
	conn, err := getConnection()
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	err = ch.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil)

	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	// should I handle this here?  Or return the channel to the caller.
	go func() {
		for d := range msgs {
			handler(d.Body)
		}
		conn.Close()
		ch.Close()

	}()

	return nil
}

func getConnection() (*amqp.Connection, error) {
	rabbitUser := "guest"
	if envVar := os.Getenv("RABBIT_USER"); envVar != "" {
		rabbitUser = envVar
	}
	rabbitPass := "guest"
	if envVar := os.Getenv("RABBIT_PASS"); envVar != "" {
		rabbitPass = envVar
	}
	rabbitServer := "localhost"
	if envVar := os.Getenv("RABBIT_SERVER"); envVar != "" {
		rabbitServer = envVar
	}
	rabbitPort := "5672"
	if envVar := os.Getenv("RABBIT_PORT"); envVar != "" {
		rabbitPort = envVar
	}

	return amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitServer, rabbitPort))
}

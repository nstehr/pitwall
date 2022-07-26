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

func Send(ctx context.Context, routingKey string, content []byte) error {
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

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPass, rabbitServer, rabbitPort))
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

package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQAdapter struct.
// This struct is used to represent a RabbitMQ adapter.
//
// Attributes:
//   - conn (*amqp.Connection): The RabbitMQ connection.
//   - channel (*amqp.Channel): The RabbitMQ channel.
//   - queue (string): The RabbitMQ queue name.
type RabbitMQAdapter struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQAdapter function.
// This function is used to create a new RabbitMQ adapter.
//
// Parameters:
//   - rabbitmqURI (string): The RabbitMQ URI.
//
// Returns:
//   - *RabbitMQAdapter: The RabbitMQ adapter.
//   - error: The error.
func NewRabbitMQAdapter(rabbitmqURI string) (*RabbitMQAdapter, error) {
	conn, err := amqp.Dial(rabbitmqURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQAdapter{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishMessage function.
// This function is used to publish a message to the RabbitMQ queue.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - queueName (string): The queue name.
//   - message (interface{}): The message.
//
// Returns:
//   - error: The error.
func (r *RabbitMQAdapter) PublishMessage(ctx context.Context, queueName string, message interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	_, err = r.channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonMessage,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf(" [x] Sent %s to queue %s\n", jsonMessage, queueName)
	return nil
}

// ConsumeMessages function.
// This function is used to consume messages from the RabbitMQ queue.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - queueName (string): The queue name.
//   - messages (chan<- []byte): The messages channel.
//
// Returns:
//   - error: The error.
func (r *RabbitMQAdapter) ConsumeMessages(ctx context.Context, queueName string, messages chan<- []byte) error {
	_, err := r.channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message from queue %s: %s", queueName, d.Body)
			messages <- d.Body
		}
	}()

	<-ctx.Done()
	log.Println("consumer cancelled")

	return nil
}

// Close function.
// This function is used to close the RabbitMQ connection and channel.
//
// Returns:
//   - error: The error.
func (r *RabbitMQAdapter) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

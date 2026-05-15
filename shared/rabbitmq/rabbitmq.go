package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	Conn     *amqp.Connection
	Channel  *amqp.Channel
	Exchange string
}

type Config struct {
	URL      string
	Exchange string
}

func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Client{
		Conn:    conn,
		Channel: ch,
	}, nil
}

// NewClient initializes a new RabbitMQ client using the RABBITMQ_URL environment variable or provided config
func NewClient(cfg ...Config) (*Client, error) {
	url := ""
	exchange := ""

	if len(cfg) > 0 {
		url = cfg[0].URL
		exchange = cfg[0].Exchange
	}

	if url == "" {
		url = os.Getenv("RABBITMQ_URL")
	}
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	client, err := Connect(url)
	if err != nil {
		return nil, err
	}

	if exchange != "" {
		client.SetExchange(exchange)
	}

	return client, nil
}

// SetExchange sets the default exchange for the client
func (c *Client) SetExchange(name string) {
	c.Exchange = name
}

func (c *Client) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// SetupExchangeQueue declares an exchange, a queue, and binds them
func (c *Client) SetupExchangeQueue(exchangeName, exchangeType, queueName, routingKey string) error {
	c.Exchange = exchangeName
	err := c.Channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := c.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.Channel.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// PublishJSON publishes a JSON-encoded message to the exchange
func (c *Client) PublishJSON(ctx context.Context, exchangeName, routingKey string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.Channel.PublishWithContext(ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishEvent publishes an event using the stored exchange
func (c *Client) PublishEvent(ctx context.Context, routingKey string, event interface{}) error {
	if c.Exchange == "" {
		return fmt.Errorf("exchange not set")
	}
	return c.PublishJSON(ctx, c.Exchange, routingKey, event)
}


// Consume starts consuming messages from a queue and passes them to a handler function
func (c *Client) Consume(queueName string, handler func(amqp.Delivery)) error {
	msgs, err := c.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
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
			handler(d)
		}
	}()

	log.Printf("Waiting for messages on queue %s", queueName)
	return nil
}

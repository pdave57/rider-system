// shared/rabbitmq/connection.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	mu       sync.RWMutex
	config   Config
	Exchange string
}

// Alias MessageBroker to Client for backward compatibility
type MessageBroker = Client

type Config struct {
	URL      string
	Exchange string
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

type ExchangeConfig struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

// Exchange names
const (
	ExchangeOrderEvents    = "order.events"
	ExchangePaymentEvents  = "payment.events"
	ExchangeDispatchEvents = "dispatch.events"
	ExchangeRiderEvents    = "rider.events"
	ExchangeAuthEvents     = "auth.events"
)

// Queue names
const (
	QueueOrderCreated       = "order.created"
	QueueOrderStatusChanged = "order.status.changed"
	QueuePaymentConfirmed   = "payment.confirmed"
	QueueDispatchAssigned   = "dispatch.assigned"
	QueueRiderAvailable     = "rider.available"
	QueueUserValidated      = "user.validated"
)

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
		conn:    conn,
		channel: ch,
	}, nil
}

// NewClient initializes a new RabbitMQ client using the RABBITMQ_URL environment variable or provided config
func NewClient(cfg ...Config) (*Client, error) {
	url := ""
	exchange := ""

	if len(cfg) > 0 {
		if cfg[0].URL != "" {
			url = cfg[0].URL
		} else if cfg[0].Host != "" {
			vhost := cfg[0].VHost
			if vhost == "" {
				vhost = "/"
			}
			port, _ := strconv.Atoi(cfg[0].Port)
			if port == 0 {
				port = 5672
			}
			url = amqp.URI{
				Scheme:   "amqp",
				Host:     cfg[0].Host,
				Port:     port,
				Username: cfg[0].User,
				Password: cfg[0].Password,
				Vhost:    vhost,
			}.String()
		}
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

	// Setup default exchanges
	if err := client.setupExchanges(); err != nil {
		log.Printf("[RabbitMQ] warning: failed to setup default exchanges: %v", err)
	}

	return client, nil
}

// NewMessageBroker is an alias for NewClient
func NewMessageBroker(cfg Config) (*MessageBroker, error) {
	return NewClient(cfg)
}

func (c *Client) setupExchanges() error {
	exchanges := []ExchangeConfig{
		{Name: ExchangeOrderEvents, Type: "topic", Durable: true},
		{Name: ExchangePaymentEvents, Type: "topic", Durable: true},
		{Name: ExchangeDispatchEvents, Type: "topic", Durable: true},
		{Name: ExchangeRiderEvents, Type: "topic", Durable: true},
		{Name: ExchangeAuthEvents, Type: "topic", Durable: true},
	}

	for _, ex := range exchanges {
		if err := c.DeclareExchange(ex.Name, ex.Type, ex.Durable); err != nil {
			return err
		}
	}
	return nil
}

// DeclareExchange explicitly declares an exchange
func (c *Client) DeclareExchange(name, kind string, durable bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.channel.ExchangeDeclare(
		name,    // name
		kind,    // type
		durable, // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
}

// SetupExchangeQueue declares an exchange, a queue, and binds them
func (c *Client) SetupExchangeQueue(exchangeName, exchangeType, queueName, routingKey string) error {
	if err := c.DeclareExchange(exchangeName, exchangeType, true); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	q, err := c.channel.QueueDeclare(
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

	err = c.channel.QueueBind(
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

// SetExchange sets the default exchange for the client
func (c *Client) SetExchange(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Exchange = name
}

// Publish sends a raw message to RabbitMQ
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // Mandatory
		false, // Immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}

// PublishJSON publishes a JSON-encoded message to the exchange
func (c *Client) PublishJSON(ctx context.Context, exchangeName, routingKey string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.Publish(ctx, exchangeName, routingKey, jsonBody)
}

// PublishEvent publishes an event using the stored exchange
func (c *Client) PublishEvent(ctx context.Context, routingKey string, event interface{}) error {
	c.mu.RLock()
	exchange := c.Exchange
	c.mu.RUnlock()

	if exchange == "" {
		return fmt.Errorf("default exchange not set")
	}
	return c.PublishJSON(ctx, exchange, routingKey, event)
}

// Consume starts consuming messages from a queue and passes them to a handler function
func (c *Client) Consume(queue, exchange, routingKey string, handler func([]byte) error) error {
	// Ensure queue and binding exist
	if err := c.SetupExchangeQueue(exchange, "topic", queue, routingKey); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Start consuming
	msgs, err := c.channel.Consume(
		queue,
		"",    // Consumer
		false, // AutoAck
		false, // Exclusive
		false, // NoLocal
		false, // NoWait
		nil,   // Args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("[RabbitMQ] Error handling message from %s: %v", queue, err)
				_ = msg.Nack(false, true) // Requeue
			} else {
				_ = msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close gracefully closes the channel and connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if c.channel != nil {
		if chErr := c.channel.Close(); chErr != nil {
			err = fmt.Errorf("failed to close channel: %w", chErr)
		}
	}
	if c.conn != nil {
		if connErr := c.conn.Close(); connErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; failed to close connection: %w", err, connErr)
			} else {
				err = fmt.Errorf("failed to close connection: %w", connErr)
			}
		}
	}
	return err
}
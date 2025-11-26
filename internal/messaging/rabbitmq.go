// Package messaging provides RabbitMQ messaging infrastructure for event-driven communication.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Exchange names for the event-driven architecture.
const (
	ExchangeCrewEvents     = "crew.events"
	ExchangeTaskEvents     = "task.events"
	ExchangeAlertEvents    = "alert.events"
	ExchangeCrewLocations  = "crew.locations"
)

// Routing keys for events.
const (
	RoutingKeyCrewLocation    = "crew.location.update"
	RoutingKeyCrewStatus      = "crew.status.update"
	RoutingKeyTaskStatus      = "task.status.update"
	RoutingKeyAlertCreated    = "alert.created"
	RoutingKeyAlertAcked      = "alert.acknowledged"
	RoutingKeyAlertResolved   = "alert.resolved"
)

// Connection represents a RabbitMQ connection with reconnection support.
type Connection struct {
	url        string
	conn       *amqp.Connection
	mu         sync.RWMutex
	connected  bool
	closeChan  chan struct{}
}

// NewConnection creates a new RabbitMQ connection.
func NewConnection(url string) *Connection {
	return &Connection{
		url:       url,
		closeChan: make(chan struct{}),
	}
}

// Connect establishes a connection to RabbitMQ.
func (c *Connection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.conn = conn
	c.connected = true

	go c.handleDisconnect()

	return nil
}

func (c *Connection) handleDisconnect() {
	notifyClose := c.conn.NotifyClose(make(chan *amqp.Error))
	select {
	case err := <-notifyClose:
		if err != nil {
			log.Printf("RabbitMQ connection closed: %v", err)
		}
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	case <-c.closeChan:
		return
	}
}

// Close closes the RabbitMQ connection.
func (c *Connection) Close() error {
	close(c.closeChan)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Channel returns a new AMQP channel.
func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.conn == nil {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return c.conn.Channel()
}

// IsConnected returns the connection status.
func (c *Connection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SetupExchanges declares all required exchanges for the event-driven architecture.
func SetupExchanges(ch *amqp.Channel) error {
	exchanges := []string{
		ExchangeCrewEvents,
		ExchangeTaskEvents,
		ExchangeAlertEvents,
		ExchangeCrewLocations,
	}

	for _, exchange := range exchanges {
		err := ch.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}

	return nil
}

// Publisher handles publishing events to RabbitMQ.
type Publisher struct {
	conn *Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

// NewPublisher creates a new event publisher.
func NewPublisher(conn *Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	if err := SetupExchanges(ch); err != nil {
		ch.Close()
		return nil, err
	}

	return &Publisher{
		conn: conn,
		ch:   ch,
	}, nil
}

// Publish publishes an event to the specified exchange with the given routing key.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, event interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Close closes the publisher channel.
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch != nil {
		return p.ch.Close()
	}
	return nil
}

// EventHandler is a function that handles incoming events.
type EventHandler func(body []byte) error

// Consumer handles consuming events from RabbitMQ.
type Consumer struct {
	conn      *Connection
	ch        *amqp.Channel
	queueName string
	handlers  map[string]EventHandler
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// NewConsumer creates a new event consumer.
func NewConsumer(conn *Connection, queueName string) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	if err := SetupExchanges(ch); err != nil {
		ch.Close()
		return nil, err
	}

	// Declare the queue
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
		handlers:  make(map[string]EventHandler),
		stopChan:  make(chan struct{}),
	}, nil
}

// BindQueue binds the queue to an exchange with a routing key and registers a handler.
func (c *Consumer) BindQueue(exchange, routingKey string, handler EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.ch.QueueBind(
		c.queueName, // queue name
		routingKey,  // routing key
		exchange,    // exchange
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	c.handlers[routingKey] = handler
	return nil
}

// Start begins consuming messages.
func (c *Consumer) Start() error {
	msgs, err := c.ch.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.handleMessage(msg)
			case <-c.stopChan:
				return
			}
		}
	}()

	return nil
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	c.mu.RLock()
	handler, ok := c.handlers[msg.RoutingKey]
	c.mu.RUnlock()

	if !ok {
		log.Printf("No handler for routing key: %s", msg.RoutingKey)
		msg.Nack(false, false)
		return
	}

	if err := handler(msg.Body); err != nil {
		log.Printf("Error handling message: %v", err)
		msg.Nack(false, true) // requeue on error
		return
	}

	msg.Ack(false)
}

// Stop stops the consumer.
func (c *Consumer) Stop() error {
	close(c.stopChan)
	if c.ch != nil {
		return c.ch.Close()
	}
	return nil
}

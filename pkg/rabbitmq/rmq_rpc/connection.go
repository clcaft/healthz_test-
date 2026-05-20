package rmqrpc

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

// Config -.
type Config struct {
	DsnList  []string
	WaitTime time.Duration
	Attempts int
}

// Connection -.
type Connection struct {
	URL              string
	ConsumerExchange string
	ConsumerType     string
	Queues           map[string]string
	Config
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Deliveries map[string]<-chan amqp.Delivery
	l          logger.Interface

	notifyConnectionClose chan *amqp.Error
	notifyChannelClose    chan *amqp.Error
}

// New -.
func New(consumerExchange, consumerType string, queues map[string]string, cfg Config, l logger.Interface) *Connection {
	conn := &Connection{
		ConsumerExchange: consumerExchange,
		ConsumerType:     consumerType,
		Queues:           queues,
		Config:           cfg,
		Deliveries:       make(map[string]<-chan amqp.Delivery),
		l:                l,
	}

	return conn
}

// AttemptConnect -.
func (c *Connection) AttemptConnect() error {
	var err error

	for _, url := range c.DsnList {
		c.URL = url
		if err = c.connect(); err == nil {
			break
		}

		c.l.Debug(context.TODO(), "rmq_rpc - AttemptConnect - c.connect: %s failed", c.URL)
	}

	if err != nil {
		return fmt.Errorf("all hosts attempted")
	}

	c.l.Debug(context.TODO(), "rmq_rpc - AttemptConnect - c.connect: %s connected", c.URL)

	return nil
}

func (c *Connection) connect() error {
	var err error

	c.Connection, err = amqp.Dial(c.URL)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	c.Channel, err = c.Connection.Channel()
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	err = c.Channel.ExchangeDeclare(
		c.ConsumerExchange,
		c.ConsumerType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	queueArgs := amqp.Table{"x-queue-type": "quorum"}
	for name, rKey := range c.Queues {
		queue, err := c.Channel.QueueDeclare(
			name,
			true,
			false,
			false,
			false,
			queueArgs,
		)
		if err != nil {
			return fmt.Errorf("c.Channel.QueueDeclare: %w", err)
		}

		err = c.Channel.QueueBind(
			queue.Name,
			rKey,
			c.ConsumerExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("c.Channel.QueueBind: %w", err)
		}

		c.Deliveries[queue.Name], err = c.Channel.Consume(
			queue.Name,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("c.Channel.Consume: %w", err)
		}
	}

	c.notifyConnectionClose = c.Connection.NotifyClose(make(chan *amqp.Error, 1))
	c.notifyConnectionClose = c.Channel.NotifyClose(make(chan *amqp.Error, 1))

	return nil
}

func (c *Connection) NotifyConnectionClose() chan *amqp.Error {
	return c.notifyConnectionClose
}

func (c *Connection) NotifyChannelClose() chan *amqp.Error {
	return c.notifyChannelClose
}

package slogelk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// log level (default: info)
	Level slog.Leveler

	Exchange string
	Url      string

	RoutingKey string
	Mandatory  bool
	Immediate  bool

	Category string
	Partner  string

	// optional: customize Amqp message builder
	Converter Converter
	// MessageConfigurator MessageConfigurator

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

var _ slog.Handler = (*amqpHandler)(nil)

type amqpHandler struct {
	option *Option
	client *client
	attrs  []slog.Attr
	groups []string
}

func (o *Option) NewAmqpHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelInfo
	}

	if o.Exchange == "" {
		panic("missing RabbitMQ Exchange")
	}

	// не вижу смысла проверять данные и бросать панику,
	// если соединения всё равно не произойдет и брошу панику далее
	if !strings.HasPrefix(o.Url, "amqp://") {
		panic("incorrect amqp URL")
	}
	c := &client{
		URL: o.Url,
	}
	if err := c.connect(); err != nil {
		fmt.Fprintln(os.Stderr, "[slogelk] error in connect function: ", err.Error())
	}

	go c.watchDog()

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	return &amqpHandler{
		option: o,
		client: c,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

type client struct {
	URL                   string
	mx                    sync.Mutex
	conn                  *amqp.Connection
	ch                    *amqp.Channel
	notifyConnectionClose chan *amqp.Error
	notifyChannelClose    chan *amqp.Error
}

func (c *client) connect() (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.conn, err = amqp.Dial(c.URL)
	if err != nil {
		return err // fmt.Errorf("amqp.Dial: %s", err)
	}

	c.ch, err = c.conn.Channel()
	if err != nil {
		return err // fmt.Errorf("c.Connection.Channel: %s", err)
	}

	c.notifyConnectionClose = c.conn.NotifyClose(make(chan *amqp.Error, 1))
	c.notifyChannelClose = c.ch.NotifyClose(make(chan *amqp.Error, 1))
	return
}

func (c *client) watchDog() {
	timeout := time.Second
	for {
		select {
		case <-time.After(timeout):
			// При обрыве связи соединение остается, после неудачного Dial соединение nil
			if c.conn != nil {
				if c.conn.IsClosed() {
					c.connect()
				}
			} else {
				c.connect()
			}
		case <-c.notifyConnectionClose:
			c.connect()
			<-time.After(timeout)
		case <-c.notifyChannelClose:
			c.connect()
			<-time.After(timeout)
		}
	}
}

func (h *amqpHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

type ctxMapKey struct{}

func SetMapToCtx(ctx context.Context, m map[string]any) context.Context {
	return context.WithValue(ctx, ctxMapKey{}, m)
}

func (h *amqpHandler) message(ctx context.Context, level slog.Leveler, msg string) map[string]any {
	m := make(map[string]any, 4)
	v := ctx.Value(ctxMapKey{})
	if val, ok := v.(map[string]any); ok && val != nil {
		m = val
	}
	l, ok := levelNames[level]
	if !ok {
		l = level.Level().String()
	}
	m["level"] = l
	m["category"] = h.option.Category
	m["partner"] = h.option.Partner
	m["data"] = msg
	return m
}

func (h *amqpHandler) send(ctx context.Context, msg map[string]any) error {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Fprintf(os.Stderr, "AmqpHandler panic has been caught in send func: %v; Stack trace: %s", rec, string(debug.Stack()))
		}
	}()
	b, _ := json.Marshal(msg)
	h.client.mx.Lock()
	defer h.client.mx.Unlock()
	err := h.client.ch.PublishWithContext(
		ctx,
		h.option.Exchange,
		h.option.RoutingKey,
		h.option.Mandatory,
		h.option.Immediate,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
	return err
}

//nolint:gocritic // implement slog.Handler interface
func (h *amqpHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.send(ctx, h.message(ctx, record.Level, h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record)))
}

func (h *amqpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &amqpHandler{
		option: h.option,
		client: h.client,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *amqpHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &amqpHandler{
		option: h.option,
		client: h.client,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/pkg/logger"
	rmqrpc "gl.eda1.ru/go/go-service-template/pkg/rabbitmq/rmq_rpc"
)

const (
	defaultWaitTime = 5 * time.Second
	defaultAttempts = 10
	defaultTimeout  = 2 * time.Second
)

// CallHandler -.
type CallHandler func(*amqp.Delivery) (interface{}, error)

// Server -.
type Server struct {
	conn   *rmqrpc.Connection
	stop   chan struct{}
	router map[string]CallHandler

	timeout time.Duration

	logger logger.Interface
}

// New -.
func New(ctx context.Context, dsn []string, serverExchange, serverType string, queues map[string]string, router map[string]CallHandler, l logger.Interface, opts ...Option) (*Server, error) {
	cfg := rmqrpc.Config{
		DsnList:  dsn,
		WaitTime: defaultWaitTime,
		Attempts: defaultAttempts,
	}

	s := &Server{
		conn:    rmqrpc.New(serverExchange, serverType, queues, cfg, l),
		stop:    make(chan struct{}),
		router:  router,
		timeout: defaultTimeout,
		logger:  l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	err := s.conn.AttemptConnect()
	if err != nil {
		s.logger.Info(ctx, "rmq_rpc server - NewServer - s.conn.AttemptConnect: %v", err)
	}

	for queue, route := range s.conn.Queues {
		go s.consumer(ctx, queue, route)
	}

	return s, nil
}

// WatchDog следит за состоянием соединения, при обрыве связи или после неудачной попытки
// соединения пытается восстановить связь.
func (s *Server) WatchDog(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info(ctx, "connection watcher stopped")
			return
		case <-time.After(defaultTimeout):
			// При обрыве связи соединение остается, после неудачного Dial соединение nil
			ctx = context.WithValue(ctx, "watchDogMsg", "<-time.After(defaultTimeout)")
			if s.conn.Connection != nil {
				if s.conn.Connection.IsClosed() {
					s.reconnect(ctx)
				}
			} else {
				s.reconnect(ctx)
			}
		case <-s.conn.NotifyConnectionClose():
			ctx = context.WithValue(ctx, "watchDogMsg", "s.conn.Connection.NotifyClose")
			// Переподключение при ошибке
			s.reconnect(ctx)
			<-time.After(defaultTimeout)
		case <-s.conn.NotifyChannelClose():
			ctx = context.WithValue(ctx, "watchDogMsg", "s.conn.Channel.NotifyClose")
			// Переподключение при ошибке
			s.reconnect(ctx)
			<-time.After(defaultTimeout)
		}
	}
}

func (s *Server) consumer(ctx context.Context, queue, route string) {
	for {
		d, opened := <-s.conn.Deliveries[queue]
		if !opened {
			s.logger.Debug(ctx, "RMQ-Server - consumer - queue: %s is not opened", queue)
			return
		}

		s.logger.Debug(ctx, "RMQ-Server - consumer - msg consumed")

		d.Type = route
		s.serveCall(ctx, &d)
	}
}

func (s *Server) serveCall(ctx context.Context, d *amqp.Delivery) {
	callHandler, ok := s.router[d.Type]
	if !ok {
		s.logger.Error(ctx, "rmq_rpc server - Server - serveCall - callHandler:", d.Type)
		_ = d.Ack(false)

		return
	}

	_, err := callHandler(d)
	if err != nil {
		// при некоторых ошибкач, например, отвал монги, нам нужно отправить команду обратно в очередь
		// а в случае если не найдена команда или несуществующее подразделение, грохаем
		// ошибки оборачиваются в AckError на уровне контроллера
		s.logger.Error(ctx, "rmq_rpc server - Server - serveCall - callHandler: %v", err)

		var ackError *AckError

		if errors.As(err, &ackError) {
			_ = d.Ack(false)
		} else {
			_ = d.Nack(false, true)
		}

		return
	}
	_ = d.Ack(false)
}

func (s *Server) publish(ctx context.Context, d *amqp.Delivery, body []byte, status string) {
	err := s.conn.Channel.PublishWithContext(ctx, d.ReplyTo, "", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Type:          status,
			Body:          body,
		})
	if err != nil {
		s.logger.Error(ctx, "rmq_rpc server - Server - publish - s.conn.Channel.Publish: %v", err)
	}
}

func (s *Server) reconnect(ctx context.Context) {
	// Останавливаем запушенных consumer
	close(s.stop)
	s.stop = make(chan struct{})

	s.logger.Debug(ctx, "reconnect, watchDog msg := %s", ctx.Value("watchDogMsg"))
	if s.conn.Channel != nil && !s.conn.Channel.IsClosed() {
		s.conn.Channel.Close()
	}
	if s.conn.Connection != nil && !s.conn.Connection.IsClosed() {
		s.conn.Connection.Close()
	}

	err := s.conn.AttemptConnect()
	if err != nil {
		return
	}

	for queue, route := range s.conn.Queues {
		go s.consumer(ctx, queue, route)
	}
}

// Shutdown -.
func (s *Server) Shutdown() error {
	if s.conn.Connection == nil {
		return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.conn.Connection: already disconnected")
	}

	close(s.stop)
	time.Sleep(s.timeout)

	if s.conn.Channel != nil {
		err := s.conn.Channel.Close()
		if err != nil {
			return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.conn.Channel.Close(): %w", err)
		}
	}
	err := s.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.conn.Connection.Close: %w", err)
	}

	return nil
}

// AckError - сообщение, на которое выполнение команды отдало такую ошибку надо ack
// для остальных ошибок Nack с requeue=true.
type AckError struct {
	err error
}

func NewAckError(err error) *AckError {
	return &AckError{err}
}

func (e AckError) Error() string {
	return e.err.Error()
}

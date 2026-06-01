package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
)

const rabbitSendQueue = "test-json-queue"

var ErrRabbitDSNEmpty = errors.New("rabbitmq dsn is empty")

type rabbitChannel struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func newRabbitChannel(dsn string, queue string) (*rabbitChannel, error) {
	if dsn == "" {
		return nil, ErrRabbitDSNEmpty
	}

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if _, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &rabbitChannel{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *rabbitChannel) Close() {
	if r.ch != nil {
		_ = r.ch.Close()
	}

	if r.conn != nil {
		_ = r.conn.Close()
	}
}

// @Summary     Send JSON message to RabbitMQ queue
// @Description Accepts JSON data, adds current time and sends it to RabbitMQ queue
// @ID          rabbit-send
// @Tags        rabbitmq
// @Accept      json
// @Produce     json
// @Param       request body dto.RabbitSendRequest true "RabbitMQ JSON message"
// @Success     200 {object} dto.Response
// @Failure     400 {object} dto.Response
// @Failure     500 {object} dto.Response
// @Router      /v1/rabbit/send [post]
func (r *Routes) sendRabbitMessage(c *gin.Context) {
	var request dto.RabbitSendRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Message: ErrBindJSON,
		})
		return
	}

	message := dto.RabbitMessage{
		Message:   request.Message,
		Source:    request.Source,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := publishJSONToQueue(c.Request.Context(), r.rabbitDSN(), rabbitSendQueue, message); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Ok",
		Data: dto.RabbitSendResponse{
			Message: "sent",
		},
	})
}

// @Summary     Read all JSON messages from RabbitMQ queue
// @Description Reads all available JSON messages from RabbitMQ queue and returns them as array
// @ID          rabbit-read-all
// @Tags        rabbitmq
// @Produce     json
// @Success     200 {object} dto.Response
// @Failure     500 {object} dto.Response
// @Router      /v1/rabbit/read-all [get]
func (r *Routes) readAllRabbitMessages(c *gin.Context) {
	messages, readErrors := readAllJSONFromQueue(c.Request.Context(), r.rabbitDSN(), rabbitSendQueue)

	message := "Ok"
	if len(readErrors) > 0 {
		message = fmt.Sprintf("read with %d error(s)", len(readErrors))
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: len(readErrors) == 0,
		Message: message,
		Data:    messages,
	})
}

func publishJSONToQueue(ctx context.Context, dsn string, queue string, data dto.RabbitMessage) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	rabbit, err := newRabbitChannel(dsn, queue)
	if err != nil {
		return err
	}
	defer rabbit.Close()

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return rabbit.ch.PublishWithContext(
		publishCtx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func readAllJSONFromQueue(ctx context.Context, dsn string, queue string) ([]dto.RabbitMessage, []error) {
	messages := make([]dto.RabbitMessage, 0)
	readErrors := make([]error, 0)

	rabbit, err := newRabbitChannel(dsn, queue)
	if err != nil {
		return messages, []error{err}
	}
	defer rabbit.Close()

	for {
		select {
		case <-ctx.Done():
			readErrors = append(readErrors, ctx.Err())
			return messages, readErrors
		default:
		}

		delivery, ok, err := rabbit.ch.Get(queue, false)
		if err != nil {
			readErrors = append(readErrors, err)
			return messages, readErrors
		}

		if !ok {
			break
		}

		var message dto.RabbitMessage
		if err = json.Unmarshal(delivery.Body, &message); err != nil {
			if ackErr := delivery.Ack(false); ackErr != nil {
				readErrors = append(readErrors, fmt.Errorf("json unmarshal: %w; ack invalid message: %w", err, ackErr))
				continue
			}

			readErrors = append(readErrors, fmt.Errorf("json unmarshal: %w", err))
			continue
		}

		if err = delivery.Ack(false); err != nil {
			readErrors = append(readErrors, fmt.Errorf("ack: %w", err))
			continue
		}

		messages = append(messages, message)
	}

	return messages, readErrors
}

func (r *Routes) rabbitDSN() string {
	if len(r.rmqDsnList) == 0 {
		return ""
	}

	return r.rmqDsnList[0]
}

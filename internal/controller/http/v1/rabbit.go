package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
)

const rabbitSendQueue = "test-json-queue"

var ErrRabbitDSNEmpty = errors.New("rabbitmq dsn is empty")

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
	messages, err := readAllJSONFromQueue(r.rabbitDSN(), rabbitSendQueue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Ok",
		Data:    messages,
	})
}

func publishJSONToQueue(ctx context.Context, dsn string, queue string, data dto.RabbitMessage) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if dsn == "" {
		return ErrRabbitDSNEmpty
	}

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(
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

func readAllJSONFromQueue(dsn string, queue string) ([]dto.RabbitMessage, error) {
	messages := make([]dto.RabbitMessage, 0)

	if dsn == "" {
		return messages, ErrRabbitDSNEmpty
	}

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return messages, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return messages, err
	}
	defer ch.Close()

	if _, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return messages, err
	}

	for {
		delivery, ok, err := ch.Get(queue, true)
		if err != nil {
			return messages, err
		}

		if !ok {
			break
		}

		var message dto.RabbitMessage
		if err = json.Unmarshal(delivery.Body, &message); err != nil {
			return messages, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *Routes) rabbitDSN() string {
	if len(r.rmqDsnList) == 0 {
		return ""
	}

	return r.rmqDsnList[0]
}

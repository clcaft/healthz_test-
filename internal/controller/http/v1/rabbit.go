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
// @Description Accepts JSON data and sends it to RabbitMQ queue
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

	if err := publishJSONToQueue(c.Request.Context(), r.rabbitDSN(), rabbitSendQueue, request); err != nil {
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

func publishJSONToQueue(ctx context.Context, dsn string, queue string, data dto.RabbitSendRequest) error {
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

func (r *Routes) rabbitDSN() string {
	if len(r.rmqDsnList) == 0 {
		return ""
	}

	return r.rmqDsnList[0]
}

package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
)

var ErrRabbitDSNEmpty = errors.New("rabbitmq dsn is empty")

// @Summary     Send JSON message to RabbitMQ queue
// @Description Accepts queue name and JSON data, then sends data to RabbitMQ
// @ID          rabbit-send
// @Tags        rabbitmq
// @Accept      json
// @Produce     json
// @Param       request body dto.RabbitSendRequest true "RabbitMQ JSON message"
// @Success     200 {object} dto.Response
// @Failure     400 {object} dto.Response
// @Failure     500 {object} dto.Response
// @Router      /v1/rabbit/send [post]
func sendRabbitMessage(c *gin.Context) {
	var request dto.RabbitSendRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Message: ErrBindJSON,
		})
		return
	}

	if request.Data == nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Message: "data is required",
		})
		return
	}

	if err := publishJSONToQueue(c.Request.Context(), request.Queue, request.Data); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Ok",
		Data: dto.RabbitSendResponse{
			Queue: request.Queue,
			Data:  request.Data,
		},
	})
}

func publishJSONToQueue(ctx context.Context, queue string, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dsn := rabbitDSN()
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

func rabbitDSN() string {
	dsnList := os.Getenv("RMQ_DSN_LIST")
	if dsnList == "" {
		return ""
	}

	parts := strings.Split(dsnList, ",")
	if len(parts) == 0 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}

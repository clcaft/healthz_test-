package v1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

const (
	ErrWrongRFC3339Format = "date parameter must be formated as 2006-01-02T15:04:05Z07:00"
	ErrWrongDateFormat    = "date parameter must be formated as YYYY-MM-DD"
	ErrParseIntArg        = "parameter must be a integer"
	ErrBindJSON           = "error when parse JSON"
	ErrNotFound           = "not found"
)

type Routes struct {
	l  logger.Interface
	uc *usecase.UseCases
}

func newRoutes(handler *gin.RouterGroup, uc *usecase.UseCases, l logger.Interface) {
	r := &Routes{l, uc}

	handler.POST("/example/plus/:ID", r.plusValue)
	handler.POST("/rabbit/send", sendRabbitMessage)
}

// @Summary     Увеличивает значение счётчика в БД
// @Description Находит в БД счётчик с указанным ID, создаёт в случае отсутствия, и увеличивает значение счётчика на указанное значение
// @ID          plus-value
// @Tags  	    example
// @Produce     json
// @Param       ID path integer true "Counter ID"
// @Success     200 {object} dto.Response
// @Failure     500 {object} dto.Response
// @Router      /v1/example/plus/{ID} [post]
func (r *Routes) plusValue(c *gin.Context) {
	ctx := r.getTrackingCtx(c)

	id, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		r.l.Warn(ctx, "http - v1 - plusValue - atoi: %v", err)
		c.JSON(http.StatusBadRequest, dto.Response{Message: ErrParseIntArg})
		return
	}

	var request dto.ExampleRequest
	if err = c.BindJSON(&request); err != nil {
		r.l.Warn(ctx, "http - v1 - plusValue - bindJson: %v", err)
		c.JSON(http.StatusBadRequest, dto.Response{Message: ErrBindJSON})
		return
	}

	result, err := r.uc.Example.PlusValue(ctx, int64(id), &request)
	if err != nil {
		r.l.Error(ctx, "http - v1 - plusValue %v", err)
		c.JSON(http.StatusBadRequest, dto.Response{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Ok", Data: result})
}

func (r *Routes) getTrackingCtx(c *gin.Context) context.Context {
	trackingInfo := map[string]interface{}{
		"uuid":        c.GetHeader("x-uuid"),
		"requestID":   c.GetHeader("x-request-id"),
		"url":         c.Request.URL.String(),
		"requestData": c.Request.Form.Encode(),
	}
	return logger.GetCtxWithTrackingInfo(c.Request.Context(), trackingInfo)
}

package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
)

// healthz godoc
// @Summary      Health check
// @Description  Returns service health status
// @Tags         health
// @Produce      json
// @Success      200  {object}  dto.HealthzResponse
// @Router       /healthz [get]
func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthzResponse{
		Status: "ok",
	})
}

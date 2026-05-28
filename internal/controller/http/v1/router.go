// Package v1 implements routing paths. Each services in own file.
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	"gl.eda1.ru/go/go-service-template/config"
	_ "gl.eda1.ru/go/go-service-template/docs"
	"gl.eda1.ru/go/go-service-template/internal/controller/http/v1/dto"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Go Service API
// @description Go Service API
// @version     1.0
// @host        localhost:8080
// @BasePath    /
func NewRouter(handler *gin.Engine, uc *usecase.UseCases, l logger.Interface, cfg *config.Config) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	handler.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// K8s probe
	handler.GET("/healthz", healthz)

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	h := handler.Group("/v1")
	{
		newRoutes(h, uc, l, cfg.RMQ.DsnList)
	}

	handler.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dto.Response{Message: "route not found"})
	})
}

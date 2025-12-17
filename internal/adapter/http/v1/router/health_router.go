package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type HealthRouter struct {
	handler *handler.HealthHandler
}

func NewHealthRouter(handler *handler.HealthHandler) *HealthRouter {
	return &HealthRouter{
		handler: handler,
	}
}

func (r *HealthRouter) Setup(rg *gin.RouterGroup) {
	rg.GET("/health", r.handler.HealthCheck)
}

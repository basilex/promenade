package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

// InitHealthModule initializes the health check module
// No dependencies required for health check
func InitHealthModule() *HealthRouter {
	handler := handler.NewHealthHandler()
	return NewHealthRouter(handler)
}

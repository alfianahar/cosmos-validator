package routers

import (
	"cosmos-tracker/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// HealthRoute registers health check endpoints
func HealthRoute(route *gin.Engine, apiVersion string) {
	healthGroup := route.Group(apiVersion)

	// Basic system health
	healthGroup.GET("/health", handlers.HealthCheck)
	healthGroup.GET("/health/data", handlers.DataHealth)
}

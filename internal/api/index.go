package routers

import (
	routersGroup "cosmos-tracker/internal/api/groups"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(route *gin.Engine) {
	// Handle 404 Not Found
	route.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"code":    http.StatusNotFound,
			"message": "Route Not Found",
		})
	})

	// API version prefix
	apiVersion := "/api/v1"

	// Register all route groups
	routersGroup.DelegationRoute(route, apiVersion)
	routersGroup.WatchlistRoute(route, apiVersion)
	routersGroup.HealthRoute(route, apiVersion)
}

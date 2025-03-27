package routers

import (
	"cosmos-tracker/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func WatchlistRoute(route *gin.Engine, apiVersion string) {
	groupRoutes := route.Group(apiVersion)

	groupRoutes.POST("/watchlist", handlers.AddToWatchlist)
	groupRoutes.GET("/watchlist", handlers.GetWatchlist)
	groupRoutes.DELETE("/watchlist/:id", handlers.RemoveFromWatchlist)
}

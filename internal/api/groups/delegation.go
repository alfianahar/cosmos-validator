package routers

import (
	"cosmos-tracker/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func DelegationRoute(route *gin.Engine, apiVersion string) {
	groupRoutes := route.Group(apiVersion)

	groupRoutes.GET("/validators/:validator/delegations/hourly", handlers.GetHourlyDelegations)
	groupRoutes.GET("/validators/:validator/delegations/daily", handlers.GetDailyDelegations)
	groupRoutes.GET("/validators/:validator/delegator/:delegator/history", handlers.GetDelegatorHistory)
}

package routers

import (
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	allowedHosts := os.Getenv("ALLOWED_HOSTS")
	r := gin.New()
	r.SetTrustedProxies([]string{allowedHosts})

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	RegisterRoutes(r) //routes register

	return r
}

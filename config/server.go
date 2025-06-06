package config

import (
	"fmt"
	"log"
	"os"
)

type ServerConfiguration struct {
	Port                 string
	Secret               string
	LimitCountPerRequest int64
}

func ServerConfig() string {
	if os.Getenv("SERVER_HOST") == "" {
		os.Setenv("SERVER_HOST", "0.0.0.0")
	}
	if os.Getenv("SERVER_PORT") == "" {
		os.Setenv("SERVER_PORT", "8080")
	}

	appServer := fmt.Sprintf("%s:%s", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))
	log.Print("Server Running at :", appServer)
	return appServer
}

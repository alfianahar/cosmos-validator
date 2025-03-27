package main

import (
	"log"

	"cosmos-tracker/config"
	api "cosmos-tracker/internal/api"
	"cosmos-tracker/internal/services"
	"cosmos-tracker/pkg/db"
)

func main() {
	// Connect to the database
	db.ConnectDB()

	// Start delegation tracking in background
	go services.StartCollector()

	// Start daily aggregation in background
	go services.ScheduleDailyAggregation()

	// Initialize Gin router
	r := api.SetupRouter()

	// Initialize configurations
	server := config.ServerConfig()
	log.Println("🚀 Server starting on", server)

	// Use Gin's Run method to start the server
	if err := r.Run(server); err != nil {
		log.Fatal("❌ Error starting server:", err)
	}
}

package handlers

import (
	"cosmos-tracker/internal/models"
	"cosmos-tracker/internal/services"
	"cosmos-tracker/pkg/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// provides a status overview of all system components
func HealthCheck(c *gin.Context) {
	// Check database connection
	sqlDB, err := db.DB.DB()
	dbStatus := "ok"
	if err != nil {
		dbStatus = "error: failed to get DB connection"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "error: database not responding"
	}

	// Check API connection
	apiStatus := "ok"
	if !services.IsHealthy() {
		apiStatus = "error: cannot connect to Cosmos API"
	}

	// Get basic statistics
	var watchlistCount int64
	var delegationCount int64
	db.DB.Model(&models.Watchlist{}).Count(&watchlistCount)
	db.DB.Model(&models.HourlyDelegation{}).Count(&delegationCount)

	// Return comprehensive health information
	c.JSON(http.StatusOK, gin.H{
		"status":    "operational",
		"timestamp": time.Now(),
		"components": gin.H{
			"database":   dbStatus,
			"cosmos_api": apiStatus,
		},
		"stats": gin.H{
			"watchlist_entries":    watchlistCount,
			"delegations_recorded": delegationCount,
		},
		"version": "1.0.0",
	})
}

// reports on data freshness and statistics
func DataHealth(c *gin.Context) {
	// Check how recent our data is
	var latestDelegation models.HourlyDelegation
	result := db.DB.Order("timestamp DESC").First(&latestDelegation)

	dataStatus := "ok"
	freshness := "unknown"

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		dataStatus = "error: cannot query data"
	} else if result.Error == gorm.ErrRecordNotFound {
		dataStatus = "warning: no data recorded yet"
	} else {
		// Check if data is stale (older than 2 hours)
		timeSinceUpdate := time.Since(latestDelegation.Timestamp)
		freshness = timeSinceUpdate.String()

		if timeSinceUpdate > 2*time.Hour {
			dataStatus = "warning: data may be stale"
		}
	}

	// Get data statistics
	var dailyDelegationCount int64
	var hourlyDelegationCount int64
	db.DB.Model(&models.HourlyDelegation{}).Count(&hourlyDelegationCount)
	db.DB.Model(&models.DailyDelegation{}).Count(&dailyDelegationCount)

	c.JSON(http.StatusOK, gin.H{
		"status":         dataStatus,
		"data_freshness": freshness,
		"statistics": gin.H{
			"hourly_records": hourlyDelegationCount,
			"daily_records":  dailyDelegationCount,
		},
	})
}

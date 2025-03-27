package handlers

import (
	"cosmos-tracker/internal/dto"
	"cosmos-tracker/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Add new validator + delegator to watchlist
func AddToWatchlist(c *gin.Context) {
	var entry dto.WatchlistEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.AddWatchlistEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to watchlist"})
}

// Get all watchlist entries
func GetWatchlist(c *gin.Context) {
	entries, err := services.GetWatchlist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve watchlist"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// Remove entry from watchlist
func RemoveFromWatchlist(c *gin.Context) {
	id := c.Param("id")

	if err := services.RemoveWatchlistEntry(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from watchlist"})
}

package handlers

import (
	"cosmos-tracker/internal/dto"
	"cosmos-tracker/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// extracts pagination parameters from the request
func getPaginationParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	return page, limit
}

// fetches hourly delegation changes for a validator with pagination
func GetHourlyDelegations(c *gin.Context) {
	validator := c.Param("validator")
	page, limit := getPaginationParams(c)

	data, total, err := services.FetchHourlyDelegationsWithPagination(validator, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.DelegationResponse{
		Data: data,
		Pagination: dto.Pagination{
			Page:       page,
			PerPage:    limit,
			TotalPages: int(totalPages),
			TotalData:  int(total),
		},
	}

	c.JSON(http.StatusOK, response)
}

// fetches daily delegation changes for a validator with pagination
func GetDailyDelegations(c *gin.Context) {
	validator := c.Param("validator")
	page, limit := getPaginationParams(c)

	data, total, err := services.FetchDailyDelegationsWithPagination(validator, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.DelegationResponse{
		Pagination: dto.Pagination{
			Page:       page,
			PerPage:    limit,
			TotalPages: int(totalPages),
			TotalData:  int(total),
		},
		Data: data,
	}

	c.JSON(http.StatusOK, response)
}

// fetches delegation history for a specific delegator with pagination
func GetDelegatorHistory(c *gin.Context) {
	validator := c.Param("validator")
	delegator := c.Param("delegator")
	page, limit := getPaginationParams(c)

	data, total, err := services.FetchDelegatorHistoryWithPagination(validator, delegator, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve data"})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	response := dto.DelegationResponse{
		Pagination: dto.Pagination{
			Page:       page,
			PerPage:    limit,
			TotalPages: int(totalPages),
			TotalData:  int(total),
		},
		Data: data,
	}

	c.JSON(http.StatusOK, response)
}

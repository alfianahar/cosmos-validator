package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"cosmos-tracker/internal/dto"
	"cosmos-tracker/internal/errors"
	"cosmos-tracker/internal/models"
	"cosmos-tracker/pkg/db"

	"gorm.io/gorm"
)

// Configurable HTTP client with timeouts and connection pooling
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          20, // Increase idle connections
		MaxConnsPerHost:       10, // Limit connections per host for rate limiting
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    true,            // Faster for smaller payloads
		ResponseHeaderTimeout: 5 * time.Second, // Timeout for header response
		ExpectContinueTimeout: 1 * time.Second, // Timeout for expect continue
	},
}

// Constants for retry mechanism
const (
	MaxRetries       = 5
	BaseRetryDelayMs = 1000  // Start with 1 second delay
	MaxRetryDelayMs  = 30000 // Maximum 30 second delay
)

// API response structure matching the actual Cosmos API format
type DelegationResponse struct {
	Delegations []struct {
		Delegation struct {
			DelegatorAddress string `json:"delegator_address"`
			ValidatorAddress string `json:"validator_address"`
			Shares           string `json:"shares"`
		} `json:"delegation"`
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balance"`
	} `json:"delegation_responses"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination,omitempty"`
}

// retrieves delegation information from the Cosmos API
func FetchDelegationData() {
	// Get watchlist entries to monitor
	watchlist, err := GetWatchlist()
	if err != nil {
		log.Printf("❌ Failed to get watchlist: %v", err)
		return
	}

	if len(watchlist) == 0 {
		log.Println("⚠️ Watchlist is empty. Add entries to start collecting delegation data.")
		return
	}

	// Track success and failure counts for metrics
	successCount := 0
	failureCount := 0

	// Process each watchlist entry
	for _, entry := range watchlist {
		log.Printf("🔍 Fetching delegations for %s -> %s", entry.ValidatorAddress, entry.ValidatorName)

		// Updated URL to use Polkachu API endpoint
		url := fmt.Sprintf("https://cosmos-api.polkachu.com/cosmos/staking/v1beta1/validators/%s/delegations",
			entry.ValidatorAddress)

		// Use retry mechanism
		resp, err := fetchWithAdvancedRetry(url, MaxRetries)
		if err != nil {
			log.Printf("❌ Error fetching delegation data after %d retries: %v", MaxRetries, err)
			failureCount++
			continue
		}
		defer resp.Body.Close()

		// Parse API response
		var result DelegationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("❌ Error decoding response: %v", err)
			failureCount++
			continue
		}

		// Process data within a transaction for consistency
		if err := processEntryData(entry, result, entry.ValidatorAddress); err != nil {
			log.Printf("❌ Error processing delegation data: %v", err)
			failureCount++
			continue
		}

		successCount++
		log.Printf("✅ Delegation data successfully updated for %s -> %s",
			entry.ValidatorAddress, entry.ValidatorName)
	}

	// Log collection summary
	log.Printf("📊 Collection summary: %d successful, %d failed", successCount, failureCount)
	if failureCount > 0 && successCount == 0 {
		log.Println("⚠️ WARNING: All collection attempts failed. Check API connectivity.")
	}
}

// saves delegation data from API response to database
func processEntryData(entry dto.WatchlistEntry, result DelegationResponse, validatorAddress string) error {
	return db.WithTransaction(func(tx *gorm.DB) error {
		// Process each delegation record
		for _, delegation := range result.Delegations {
			// Extract delegator address from the nested delegation object
			delegatorAddress := delegation.Delegation.DelegatorAddress

			// Skip if zero amount to avoid noise in the data
			delegationAmount, err := strconv.ParseInt(delegation.Balance.Amount, 10, 64)
			if err != nil {
				log.Printf("❌ Error parsing delegation amount '%s': %v", delegation.Balance.Amount, err)
				continue
			}

			// Parse shares for additional data
			var sharesFloat float64
			if delegation.Delegation.Shares != "" {
				sharesFloat, err = strconv.ParseFloat(delegation.Delegation.Shares, 64)
				if err != nil {
					log.Printf("⚠️ Error parsing shares value '%s': %v", delegation.Delegation.Shares, err)
					// Continue anyway since this is optional data
				}
			}

			// Fetch the last recorded amount to calculate the change
			var lastRecord models.HourlyDelegation
			if err := tx.Where("validator_address = ? AND delegator_address = ?",
				validatorAddress, delegatorAddress).
				Order("timestamp DESC").
				First(&lastRecord).Error; err != nil && err.Error() != "record not found" {
				return err
			}

			// Calculate change amount
			changeAmount := delegationAmount
			if lastRecord.ID != 0 {
				changeAmount = delegationAmount - lastRecord.DelegationAmount
			}

			// Find associated watchlist entry for foreign key
			var watchlistItem models.Watchlist
			if err := tx.Where("validator_address = ?",
				validatorAddress).First(&watchlistItem).Error; err != nil {
				log.Printf("⚠️ Warning: No watchlist entry found for %s -> %s",
					validatorAddress, delegatorAddress)
				continue
			}

			// Save the new delegation snapshot with watchlist reference
			entry := models.HourlyDelegation{
				WatchlistID:      watchlistItem.ID,
				ValidatorAddress: validatorAddress,
				DelegatorAddress: delegatorAddress,
				DelegationAmount: delegationAmount,
				ChangeAmount:     changeAmount,
				Shares:           sharesFloat, // Store the parsed shares value
				Timestamp:        time.Now(),
			}

			if err := tx.Create(&entry).Error; err != nil {
				return err
			}

			// Log significant delegation changes for monitoring
			if lastRecord.ID != 0 && math.Abs(float64(changeAmount)) > float64(delegationAmount)*0.05 {
				log.Printf("📈 Significant delegation change: %s -> %s changed by %d (%.2f%%)",
					validatorAddress, delegatorAddress, changeAmount,
					float64(changeAmount)*100/float64(lastRecord.DelegationAmount))
			}
		}
		return nil
	})
}

// implements exponential backoff with jitter for API resilience
func fetchWithAdvancedRetry(url string, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Try the request
		resp, err = httpClient.Get(url)

		// Success case
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Handle specific status codes
		if resp != nil {
			switch resp.StatusCode {
			case http.StatusTooManyRequests, http.StatusServiceUnavailable:
				log.Printf("⚠️ Rate limited (%d). Backing off...", resp.StatusCode)
				// Check for Retry-After header
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
						time.Sleep(time.Duration(seconds) * time.Second)
						resp.Body.Close()
						continue
					}
				}
			case http.StatusBadRequest, http.StatusNotFound:
				// Don't retry for client errors
				errMsg := fmt.Sprintf("API client error: %d", resp.StatusCode)
				resp.Body.Close()
				return nil, errors.NewBadRequestError(errMsg, nil)
			}
			resp.Body.Close()
		}

		// Calculate backoff with jitter for distributed clients
		baseDelay := math.Min(float64(BaseRetryDelayMs)*math.Pow(2, float64(attempt)), float64(MaxRetryDelayMs))
		jitter := (baseDelay * 0.2) * (0.5 + rand.Float64()) // Add 0-20% jitter
		backoffTime := time.Duration(baseDelay+jitter) * time.Millisecond

		log.Printf("🔄 Retrying API call in %v (attempt %d/%d)", backoffTime, attempt+1, maxRetries)
		time.Sleep(backoffTime)
	}

	if err != nil {
		return nil, errors.NewServiceUnavailableError("API service unavailable", err)
	}

	return nil, errors.NewInternalServerError("Maximum retries exceeded", nil)
}

// runs the delegation collector on a schedule
func StartCollector() {
	// Run immediately at startup
	log.Println("🚀 Initial data collection starting...")
	FetchDelegationData()

	// Then run hourly
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("⏳ Running scheduled collection...")
		FetchDelegationData()
	}
}

// performs a simple check to verify the collector service is functional
func IsHealthy() bool {
	// Check if we can connect to the Cosmos API
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Updated URL for health check to match new API provider
	req, err := http.NewRequestWithContext(ctx, "GET", "https://cosmos-api.polkachu.com/cosmos/base/tendermint/v1beta1/node_info", nil)
	if err != nil {
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

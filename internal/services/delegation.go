package services

import (
	"log"
	"time"

	"cosmos-tracker/internal/dto"
	"cosmos-tracker/internal/models"
	"cosmos-tracker/pkg/db"
)

// retrieves paginated hourly delegation changes
func FetchHourlyDelegationsWithPagination(validatorAddress string, page, limit int) ([]dto.HourlyDelegationDTO, int64, error) {
	var delegations []models.HourlyDelegation
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := db.DB.Model(&models.HourlyDelegation{}).
		Where("validator_address = ?", validatorAddress).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	err := db.DB.Where("validator_address = ?", validatorAddress).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&delegations).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to DTOs
	result := make([]dto.HourlyDelegationDTO, len(delegations))
	for i, d := range delegations {
		result[i] = dto.HourlyDelegationDTO{
			ID:               d.ID,
			ValidatorAddress: d.ValidatorAddress,
			DelegatorAddress: d.DelegatorAddress,
			DelegationAmount: d.DelegationAmount,
			ChangeAmount:     d.ChangeAmount,
			Timestamp:        d.Timestamp,
		}
	}

	return result, total, nil
}

// retrieves paginated daily delegation changes
func FetchDailyDelegationsWithPagination(validatorAddress string, page, limit int) ([]dto.DailyDelegationDTO, int64, error) {
	var delegations []models.DailyDelegation
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := db.DB.Model(&models.DailyDelegation{}).
		Where("validator_address = ?", validatorAddress).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	err := db.DB.Where("validator_address = ?", validatorAddress).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&delegations).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to DTOs
	result := make([]dto.DailyDelegationDTO, len(delegations))
	for i, d := range delegations {
		result[i] = dto.DailyDelegationDTO{
			ID:               d.ID,
			ValidatorAddress: d.ValidatorAddress,
			DelegatorAddress: d.DelegatorAddress,
			TotalDelegation:  d.TotalDelegation,
			Date:             d.Date,
		}
	}

	return result, total, nil
}

// retrieves paginated delegation history for a specific delegator
func FetchDelegatorHistoryWithPagination(validatorAddress, delegatorAddress string, page, limit int) ([]dto.HourlyDelegationDTO, int64, error) {
	var history []models.HourlyDelegation
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := db.DB.Model(&models.HourlyDelegation{}).
		Where("validator_address = ? AND delegator_address = ?", validatorAddress, delegatorAddress).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated data
	err := db.DB.Where("validator_address = ? AND delegator_address = ?", validatorAddress, delegatorAddress).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to DTOs
	result := make([]dto.HourlyDelegationDTO, len(history))
	for i, h := range history {
		result[i] = dto.HourlyDelegationDTO{
			ID:               h.ID,
			ValidatorAddress: h.ValidatorAddress,
			DelegatorAddress: h.DelegatorAddress,
			DelegationAmount: h.DelegationAmount,
			ChangeAmount:     h.ChangeAmount,
			Timestamp:        h.Timestamp,
		}
	}

	return result, total, nil
}

// compiles hourly data into daily summaries
func AggregateDailyDelegations() error {
	// Get current date at midnight for proper grouping
	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())

	// Find all watchlist entries
	var watchlistItems []models.Watchlist
	if err := db.DB.Find(&watchlistItems).Error; err != nil {
		return err
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, watchlist := range watchlistItems {
		// Get the latest delegation amount for each validator-delegator pair for yesterday
		var latestDelegation models.HourlyDelegation
		subQuery := tx.Model(&models.HourlyDelegation{}).
			Where("watchlist_id = ? AND timestamp >= ? AND timestamp < ?",
				watchlist.ID, yesterday, yesterday.AddDate(0, 0, 1)).
			Order("timestamp DESC").
			Limit(1)

		if err := subQuery.First(&latestDelegation).Error; err != nil {
			if err.Error() != "record not found" {
				tx.Rollback()
				return err
			}
			continue // No data for yesterday
		}

		// Check if daily record already exists
		var existingDaily models.DailyDelegation
		err := tx.Where("watchlist_id = ? AND date = ?", watchlist.ID, yesterday).
			First(&existingDaily).Error

		if err != nil && err.Error() != "record not found" {
			tx.Rollback()
			return err
		}

		// Create or update daily record
		if existingDaily.ID == 0 {
			dailyRecord := models.DailyDelegation{
				WatchlistID:      watchlist.ID,
				ValidatorAddress: watchlist.ValidatorAddress,
				DelegatorAddress: watchlist.DelegatorAddress,
				TotalDelegation:  latestDelegation.DelegationAmount,
				Date:             yesterday,
			}
			if err := tx.Create(&dailyRecord).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			existingDaily.TotalDelegation = latestDelegation.DelegationAmount
			if err := tx.Save(&existingDaily).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// schedules daily aggregation to run at midnight
func ScheduleDailyAggregation() {
	// Calculate time until next midnight
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
	duration := midnight.Sub(now)

	log.Printf("🕒 Scheduling daily aggregation to run in %v", duration)

	// Initial wait until midnight
	timer := time.NewTimer(duration)
	<-timer.C

	// Run the first aggregation
	if err := AggregateDailyDelegations(); err != nil {
		log.Printf("❌ Daily aggregation failed: %v", err)
	} else {
		log.Println("✅ Daily aggregation completed successfully")
	}

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := AggregateDailyDelegations(); err != nil {
			log.Printf("❌ Daily aggregation failed: %v", err)
		} else {
			log.Println("✅ Daily aggregation completed successfully")
		}
	}
}

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

	log.Printf("Found %d watchlist items to process", len(watchlistItems))

	tx := db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, watchlist := range watchlistItems {
		// Get all unique delegator addresses for this validator
		var delegatorAddresses []string
		if err := tx.Model(&models.HourlyDelegation{}).
			Where("validator_address = ? AND timestamp >= ? AND timestamp < ?",
				watchlist.ValidatorAddress, yesterday, yesterday.AddDate(0, 0, 1)).
			Distinct("delegator_address").
			Pluck("delegator_address", &delegatorAddresses).Error; err != nil {
			tx.Rollback()
			return err
		}

		log.Printf("Found %d delegators for validator %s", len(delegatorAddresses), watchlist.ValidatorAddress)

		// Process each delegator for this validator
		for _, delegatorAddr := range delegatorAddresses {
			var latestDelegation models.HourlyDelegation

			// Find the latest hourly record for this validator-delegator pair
			if err := tx.Model(&models.HourlyDelegation{}).
				Where("validator_address = ? AND delegator_address = ? AND timestamp >= ? AND timestamp < ?",
					watchlist.ValidatorAddress, delegatorAddr, yesterday, yesterday.AddDate(0, 0, 1)).
				Order("timestamp DESC").
				Limit(1).
				First(&latestDelegation).Error; err != nil {
				if err.Error() != "record not found" {
					log.Printf("Error querying hourly delegation: %v", err)
					tx.Rollback()
					return err
				}
				continue // No data for this delegator yesterday
			}

			log.Printf("Found latest delegation for %s -> %s: %d tokens",
				watchlist.ValidatorAddress, delegatorAddr, latestDelegation.DelegationAmount)

			// Check if daily record already exists
			var existingDaily models.DailyDelegation
			err := tx.Where("validator_address = ? AND delegator_address = ? AND date = ?",
				watchlist.ValidatorAddress, delegatorAddr, yesterday).
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
					DelegatorAddress: delegatorAddr, // Include delegator address
					TotalDelegation:  latestDelegation.DelegationAmount,
					TotalShares:      latestDelegation.Shares,
					Date:             yesterday,
				}
				if err := tx.Create(&dailyRecord).Error; err != nil {
					log.Printf("❌ Error creating daily delegation record: %v", err)
					tx.Rollback()
					return err
				}
				log.Printf("✅ Created new daily record for %s -> %s",
					watchlist.ValidatorAddress, delegatorAddr)
			} else {
				existingDaily.TotalDelegation = latestDelegation.DelegationAmount
				existingDaily.TotalShares = latestDelegation.Shares
				if err := tx.Save(&existingDaily).Error; err != nil {
					log.Printf("❌ Error updating daily delegation record: %v", err)
					tx.Rollback()
					return err
				}
				log.Printf("✅ Updated daily record for %s -> %s",
					watchlist.ValidatorAddress, delegatorAddr)
			}
		}
	}

	log.Println("⭐ Daily aggregation transaction completed, committing changes...")
	return tx.Commit().Error
}

// schedules daily aggregation to run at midnight
func ScheduleDailyAggregation() {
	// Run aggregation immediately at startup
	log.Println("🚀 Running initial daily aggregation...")
	if err := AggregateDailyDelegations(); err != nil {
		log.Printf("❌ Initial daily aggregation failed: %v", err)
	} else {
		log.Println("✅ Initial daily aggregation completed successfully")
	}

	// Calculate time until next midnight
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
	duration := midnight.Sub(now)

	log.Printf("🕒 Scheduling next daily aggregation to run in %v", duration)

	// Schedule the next run at midnight because for the first init
	go func() {
		time.Sleep(duration)

		log.Println("⏰ Running scheduled midnight aggregation...")
		if err := AggregateDailyDelegations(); err != nil {
			log.Printf("❌ Midnight aggregation failed: %v", err)
		} else {
			log.Println("✅ Midnight aggregation completed successfully")
		}

		// Then for every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("⏰ Running daily scheduled aggregation...")
			if err := AggregateDailyDelegations(); err != nil {
				log.Printf("❌ Daily aggregation failed: %v", err)
			} else {
				log.Println("✅ Daily aggregation completed successfully")
			}
		}
	}()
}

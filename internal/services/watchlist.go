package services

import (
	"cosmos-tracker/internal/dto"
	"cosmos-tracker/internal/models"
	"cosmos-tracker/pkg/db"
)

// adds a new entry to the watchlist
func AddWatchlistEntry(entry dto.WatchlistEntry) error {
	watchlistItem := models.Watchlist{
		ValidatorAddress: entry.ValidatorAddress,
		DelegatorAddress: entry.DelegatorAddress,
	}

	return db.DB.Create(&watchlistItem).Error
}

// returns all entries in the watchlist
func GetWatchlist() ([]dto.WatchlistEntry, error) {
	var watchlistItems []models.Watchlist
	if err := db.DB.Find(&watchlistItems).Error; err != nil {
		return nil, err
	}

	// Convert from DB model to DTO
	entries := make([]dto.WatchlistEntry, len(watchlistItems))
	for i, item := range watchlistItems {
		entries[i] = dto.WatchlistEntry{
			ID:               int(item.ID),
			ValidatorAddress: item.ValidatorAddress,
			DelegatorAddress: item.DelegatorAddress,
		}
	}

	return entries, nil
}

// removes an entry from the watchlist by ID
func RemoveWatchlistEntry(id string) error {
	return db.DB.Delete(&models.Watchlist{}, id).Error
}

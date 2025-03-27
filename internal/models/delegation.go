package models

import "time"

type HourlyDelegation struct {
	ID               uint      `gorm:"primaryKey"`
	WatchlistID      uint      `gorm:"index"`
	Watchlist        Watchlist `gorm:"foreignKey:WatchlistID"`
	ValidatorAddress string    `gorm:"index"` // safe column if not using watchlist
	DelegatorAddress string    `gorm:"index"` // safe column if not using watchlist
	DelegationAmount int64
	ChangeAmount     int64
	Shares           float64
	Timestamp        time.Time `gorm:"autoCreateTime;index"`
}

type DailyDelegation struct {
	ID               uint      `gorm:"primaryKey"`
	WatchlistID      uint      `gorm:"index"`
	Watchlist        Watchlist `gorm:"foreignKey:WatchlistID"`
	ValidatorAddress string    `gorm:"index"` // safe column if not using watchlist
	DelegatorAddress string    `gorm:"index"` // safe column if not using watchlist
	TotalDelegation  int64
	TotalShares      float64
	Date             time.Time `gorm:"autoCreateTime;index"`
}

package models

// Watchlist represents a validator-delegator pair to track
type Watchlist struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	ValidatorAddress string `gorm:"index" json:"validator_address"`
	ValidatorName    string `gorm:"type:varchar(100)" json:"validator_name"`
}

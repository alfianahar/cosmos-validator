package dto

// represents a watchlist entry for tracking validator-delegator pairs
type WatchlistEntry struct {
	ID               int    `json:"id"`
	ValidatorAddress string `json:"validator_address"`
	ValidatorName    string `json:"validator_name"`
}

package dto

import "time"

// represents hourly delegation metrics retrieved from the Cosmos network
type HourlyDelegationDTO struct {
	ID               uint      `json:"id"`
	ValidatorAddress string    `json:"validator_address"`
	DelegatorAddress string    `json:"delegator_address"`
	DelegationAmount int64     `json:"delegation_amount"`
	ChangeAmount     int64     `json:"change_amount"`
	Shares           float64   `json:"shares,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// represents aggregated daily delegation metrics
type DailyDelegationDTO struct {
	ID               uint      `json:"id"`
	ValidatorAddress string    `json:"validator_address"`
	DelegatorAddress string    `json:"delegator_address"`
	TotalDelegation  int64     `json:"total_delegation"`
	TotalShares      float64   `json:"total_shares,omitempty"`
	Date             time.Time `json:"date"`
}

// standardizes the API response format for all delegation endpoints
type DelegationResponse struct {
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination,omitempty"`
}

// provides standardized error format for API responses
type ErrorResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

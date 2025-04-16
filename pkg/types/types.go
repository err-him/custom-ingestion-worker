package types

import "time"

// Sample represents a data sample with validation
type Sample struct {
	CustomerID string    `json:"customerId"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ValidationError represents a validation error
type ValidationError struct {
	CustomerID string
	Reason     string
}

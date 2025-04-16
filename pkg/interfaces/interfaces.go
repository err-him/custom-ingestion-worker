package interfaces

import (
	"gohighlevel/pkg/types"
)

// Validator interface for sample validation
type Validator interface {
	ValidateSample(sample types.Sample) error
	WriteErrorLog(customerId, reason string) error
}

// Database interface for database operations
type Database interface {
	Init() error
	Close()
	InsertSample(sample types.Sample) error
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	IsAllowed(customerID string) bool
	GetRemainingRequests(customerID string) int
}

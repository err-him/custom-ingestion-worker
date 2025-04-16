package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter tracks requests per customer ID within a 1-minute window
type RateLimiter struct {
	requestsPerMinute int
	mu                sync.RWMutex
	requests          map[string][]time.Time
}

// NewRateLimiter creates a new rate limiter with the specified requests per minute limit
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		requests:          make(map[string][]time.Time),
	}
}

// IsAllowed checks if a request is allowed based on rate limits within a 1-minute window
func (r *RateLimiter) IsAllowed(customerID string, createdAt time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Initialize if customer doesn't exist
	if _, exists := r.requests[customerID]; !exists {
		r.requests[customerID] = []time.Time{createdAt}
		return true
	}

	// Get requests within the last 60 seconds
	var validRequests []time.Time
	windowStart := createdAt.Add(-1 * time.Minute)

	// Keep track of requests in the sliding window
	for _, t := range r.requests[customerID] {
		if t.After(windowStart) || t.Equal(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Update the requests list with only valid ones
	r.requests[customerID] = validRequests

	// Check if under limit
	if len(validRequests) < r.requestsPerMinute {
		r.requests[customerID] = append(r.requests[customerID], createdAt)
		return true
	}

	return false
}

// GetRemainingRequests returns the number of remaining requests allowed within the current minute
func (r *RateLimiter) GetRemainingRequests(customerID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)
	var validCount int

	if times, exists := r.requests[customerID]; exists {
		// Count requests within the last 60 seconds
		for _, t := range times {
			if t.After(windowStart) || t.Equal(windowStart) {
				validCount++
			}
		}
	}

	return r.requestsPerMinute - validCount
}

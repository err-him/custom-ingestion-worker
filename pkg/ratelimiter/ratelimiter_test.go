package ratelimiter

import (
	"testing"
	"time"
)

func TestRateLimiterBasic(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per minute
	customerID := "test123"
	now := time.Now()

	// Test 1: First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !limiter.IsAllowed(customerID, now) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Test 2: 6th request should be denied
	if limiter.IsAllowed(customerID, now) {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiterTimeWindow(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per minute
	customerID := "test123"
	baseTime := time.Date(2024, 3, 26, 12, 0, 0, 0, time.UTC)

	// Test 1: Make 5 requests within 30 seconds
	for i := 0; i < 5; i++ {
		requestTime := baseTime.Add(time.Duration(i) * 6 * time.Second) // 6 seconds apart
		if !limiter.IsAllowed(customerID, requestTime) {
			t.Errorf("Request %d at %v should be allowed", i+1, requestTime)
		}
	}

	// Test 2: 6th request at 31 seconds should be denied
	if limiter.IsAllowed(customerID, baseTime.Add(31*time.Second)) {
		t.Error("6th request should be denied within the same minute")
	}

	// Test 3: Request after 1 minute should be allowed (window reset)
	if !limiter.IsAllowed(customerID, baseTime.Add(61*time.Second)) {
		t.Error("Request after 1 minute should be allowed")
	}
}

func TestRateLimiterMultipleCustomers(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per minute
	customer1 := "cust1"
	customer2 := "cust2"
	now := time.Now()

	// Test 1: Customer 1 makes 5 requests
	for i := 0; i < 5; i++ {
		if !limiter.IsAllowed(customer1, now) {
			t.Errorf("Customer 1 request %d should be allowed", i+1)
		}
	}

	// Test 2: Customer 1's 6th request should be denied
	if limiter.IsAllowed(customer1, now) {
		t.Error("Customer 1's 6th request should be denied")
	}

	// Test 3: Customer 2 should still be able to make requests
	for i := 0; i < 5; i++ {
		if !limiter.IsAllowed(customer2, now) {
			t.Errorf("Customer 2 request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiterEdgeCases(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per minute
	customerID := "test123"
	baseTime := time.Date(2024, 3, 26, 12, 0, 0, 0, time.UTC)

	// Test 1: Requests exactly 60 seconds apart should always be allowed
	for i := 0; i < 3; i++ {
		requestTime := baseTime.Add(time.Duration(i) * time.Minute)
		if !limiter.IsAllowed(customerID, requestTime) {
			t.Errorf("Request at %v should be allowed", requestTime)
		}
	}

	// Test 2: Make 3 requests at the start of a minute
	startTime := baseTime.Add(5 * time.Minute)
	for i := 0; i < 3; i++ {
		if !limiter.IsAllowed(customerID, startTime) {
			t.Errorf("Request %d at start of minute should be allowed", i+1)
		}
	}

	// Test 3: Make 2 more requests 30 seconds later (should be allowed as we're within the limit)
	thirtySecondsLater := startTime.Add(30 * time.Second)
	for i := 0; i < 2; i++ {
		if !limiter.IsAllowed(customerID, thirtySecondsLater) {
			t.Errorf("Request %d at 30 seconds later should be allowed", i+1)
		}
	}

	// Test 4: Next request should be rejected as we've hit our 5 request limit in the sliding window
	if limiter.IsAllowed(customerID, thirtySecondsLater) {
		t.Error("Request should be rejected as we've hit the limit in the sliding window")
	}

	// Test 5: Request after 31 seconds from the first request should still be rejected
	// as we still have 5 requests in the last minute
	afterThirtyOneSeconds := startTime.Add(31 * time.Second)
	if limiter.IsAllowed(customerID, afterThirtyOneSeconds) {
		t.Error("Request should be rejected as we still have 5 requests in the last minute")
	}

	// Test 6: Request after 61 seconds from the first request should be allowed
	// as the first request is now outside the sliding window
	afterSixtyOneSeconds := startTime.Add(61 * time.Second)
	if !limiter.IsAllowed(customerID, afterSixtyOneSeconds) {
		t.Error("Request should be allowed as oldest request is now outside the sliding window")
	}
}

func TestRateLimiterRemainingRequests(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per minute
	customerID := "test123"
	now := time.Now()

	// Test 1: Initially should have 5 remaining requests
	if remaining := limiter.GetRemainingRequests(customerID); remaining != 5 {
		t.Errorf("Expected 5 remaining requests, got %d", remaining)
	}

	// Test 2: After 2 requests, should have 3 remaining
	limiter.IsAllowed(customerID, now)
	limiter.IsAllowed(customerID, now)
	if remaining := limiter.GetRemainingRequests(customerID); remaining != 3 {
		t.Errorf("Expected 3 remaining requests, got %d", remaining)
	}

	// Test 3: After using all requests, should have 0 remaining
	for i := 0; i < 3; i++ {
		limiter.IsAllowed(customerID, now)
	}
	if remaining := limiter.GetRemainingRequests(customerID); remaining != 0 {
		t.Errorf("Expected 0 remaining requests, got %d", remaining)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000) // High limit for benchmark
	customerID := "bench123"
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.IsAllowed(customerID, now)
	}
}

func BenchmarkRateLimiterParallel(b *testing.B) {
	limiter := NewRateLimiter(1000) // High limit for benchmark

	b.RunParallel(func(pb *testing.PB) {
		customerID := "bench123"
		now := time.Now()
		for pb.Next() {
			limiter.IsAllowed(customerID, now)
		}
	})
}

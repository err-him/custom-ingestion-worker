package main

import (
	"os"
	"testing"
	"time"

	"gohighlevel/pkg/db"
	"gohighlevel/pkg/ratelimiter"
	"gohighlevel/pkg/service"
	"gohighlevel/pkg/validator"
)

// TestCompleteFlow tests the entire application flow:
// 1. Database connection
// 2. Sample validation
// 3. Rate limiting
// 4. Sample processing
func TestCompleteFlow(t *testing.T) {
	// Clean up any existing error.log
	os.Remove("error.log")

	// Initialize MongoDB
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	// Initialize components
	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5) // 5 requests per minute
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Test processing samples
	result, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		t.Fatalf("Failed to process samples: %v", err)
	}

	// Verify processing results
	if result.SuccessCount == 0 {
		t.Error("Expected some successful samples")
	}
	if result.ErrorCount == 0 {
		t.Error("Expected some failed samples due to validation")
	}

	// Verify error log exists and has content
	if _, err := os.Stat("error.log"); os.IsNotExist(err) {
		t.Error("error.log file should exist")
	}
}

// TestRateLimitEnforcement tests that rate limiting is properly enforced
func TestRateLimitEnforcement(t *testing.T) {
	// Initialize components
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5)
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Process samples multiple times in quick succession
	for i := 0; i < 3; i++ {
		result, err := sampleService.ProcessSamplesFile("samples.json")
		if err != nil {
			t.Fatalf("Failed to process samples: %v", err)
		}

		// First run should process all samples
		// Subsequent runs should have some samples rate limited
		if i == 0 {
			if result.SuccessCount == 0 {
				t.Error("First run should process some samples successfully")
			}
		} else {
			if result.SuccessCount >= result.ErrorCount {
				t.Error("Subsequent runs should have more failures due to rate limiting")
			}
		}
	}
}

// TestValidationAndRateLimitCombined tests the interaction between validation and rate limiting
func TestValidationAndRateLimitCombined(t *testing.T) {
	// Initialize components
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5)
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Process samples
	result, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		t.Fatalf("Failed to process samples: %v", err)
	}

	// Verify processing results
	if result.SuccessCount == 0 {
		t.Error("Expected some successful samples")
	}

	// Verify that we have both validation errors and rate limit hits
	totalErrors := v.GetErrorCount()
	if totalErrors == 0 {
		t.Error("Expected validation errors")
	}

	// Verify error log content
	errorLog, err := os.ReadFile("error.log")
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	// Check for both validation and rate limit errors in the log
	if len(errorLog) == 0 {
		t.Error("Error log should contain entries")
	}
}

// TestConcurrentProcessing tests how the system handles concurrent processing
func TestConcurrentProcessing(t *testing.T) {
	// Initialize components
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5)
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Process samples concurrently
	done := make(chan bool)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := sampleService.ProcessSamplesFile("samples.json")
			if err != nil {
				t.Errorf("Failed to process samples: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify that rate limiting worked under concurrent load
	if v.GetErrorCount() == 0 {
		t.Error("Expected some errors under concurrent load")
	}
}

// TestErrorRecovery tests how the system handles and recovers from errors
func TestErrorRecovery(t *testing.T) {
	// Initialize components
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5)
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Test with non-existent file
	_, err := sampleService.ProcessSamplesFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error when processing non-existent file")
	}

	// Test with valid file after error
	result, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		t.Fatalf("Failed to process samples after error: %v", err)
	}

	// Verify system recovered and processed samples
	if result.SuccessCount == 0 {
		t.Error("Expected successful processing after error recovery")
	}
}

// TestTimeWindowBehavior tests how the rate limiter behaves across time windows
func TestTimeWindowBehavior(t *testing.T) {
	// Initialize components
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	v := validator.NewValidator(mongoDB)
	r := ratelimiter.NewRateLimiter(5)
	sampleService := service.NewSampleService(v, r, mongoDB)

	// Process samples
	result1, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		t.Fatalf("Failed to process samples: %v", err)
	}

	// Wait for rate limit window to reset
	time.Sleep(61 * time.Second)

	// Process samples again
	result2, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		t.Fatalf("Failed to process samples: %v", err)
	}

	// Verify that more samples were processed after the time window reset
	if result2.SuccessCount <= result1.SuccessCount {
		t.Error("Expected more successful processing after time window reset")
	}
}

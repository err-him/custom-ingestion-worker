package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestMainFlow tests the main application flow with a sample file
func TestMainFlow(t *testing.T) {
	// Create a temporary samples.json file
	samples := []struct {
		CustomerID string `json:"customerId"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		CreatedAt  string `json:"createdAt"`
	}{
		{
			CustomerID: "test1",
			Name:       "Test User 1",
			Email:      "test1@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "test2",
			Name:       "Test User 2",
			Email:      "test2@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "test3",
			Name:       "Invalid Email",
			Email:      "not-an-email",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
	}

	// Create temporary file
	file, err := os.CreateTemp("", "samples-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Write samples to file
	data := struct {
		Samples []interface{} `json:"samples"`
	}{
		Samples: make([]interface{}, len(samples)),
	}
	for i, sample := range samples {
		data.Samples[i] = sample
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()

	// Remove error.log if it exists
	os.Remove("error.log")

	// Run main with the test file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", file.Name()}

	// Run main in a goroutine since it calls os.Exit
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	// Wait for main to complete
	<-done

	// Check error.log exists and contains expected errors
	errorLog, err := os.ReadFile("error.log")
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	// Verify error log contains expected error for invalid email
	if len(errorLog) == 0 {
		t.Error("Expected error.log to contain validation errors")
	}
}

// TestMainWithInvalidFile tests main's behavior with an invalid file
func TestMainWithInvalidFile(t *testing.T) {
	// Create an invalid JSON file
	file, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Write invalid JSON
	file.WriteString("{ invalid json")
	file.Close()

	// Remove error.log if it exists
	os.Remove("error.log")

	// Run main with the invalid file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", file.Name()}

	// Run main in a goroutine since it calls os.Exit
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	// Wait for main to complete
	<-done

	// Check error.log exists and contains expected errors
	errorLog, err := os.ReadFile("error.log")
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	// Verify error log contains expected error for invalid JSON
	if len(errorLog) == 0 {
		t.Error("Expected error.log to contain JSON parsing errors")
	}
}

// TestMainWithRateLimit tests main's behavior with rate limiting
func TestMainWithRateLimit(t *testing.T) {
	// Create samples that exceed rate limit
	samples := make([]struct {
		CustomerID string `json:"customerId"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		CreatedAt  string `json:"createdAt"`
	}, 10)

	// Create 10 samples with same customer ID
	for i := range samples {
		samples[i] = struct {
			CustomerID string `json:"customerId"`
			Name       string `json:"name"`
			Email      string `json:"email"`
			CreatedAt  string `json:"createdAt"`
		}{
			CustomerID: "rate-test",
			Name:       "Rate Test User",
			Email:      "rate@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
	}

	// Create temporary file
	file, err := os.CreateTemp("", "samples-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Write samples to file
	data := struct {
		Samples []interface{} `json:"samples"`
	}{
		Samples: make([]interface{}, len(samples)),
	}
	for i, sample := range samples {
		data.Samples[i] = sample
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()

	// Remove error.log if it exists
	os.Remove("error.log")

	// Run main with the test file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", file.Name()}

	// Run main in a goroutine since it calls os.Exit
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	// Wait for main to complete
	<-done

	// Check error.log exists and contains expected errors
	errorLog, err := os.ReadFile("error.log")
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	// Verify error log contains rate limit errors
	if len(errorLog) == 0 {
		t.Error("Expected error.log to contain rate limit errors")
	}
}

// TestMainWithMongoDBConnection tests main's behavior with MongoDB connection issues
func TestMainWithMongoDBConnection(t *testing.T) {
	// Create a temporary samples.json file
	samples := []struct {
		CustomerID string `json:"customerId"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		CreatedAt  string `json:"createdAt"`
	}{
		{
			CustomerID: "test1",
			Name:       "Test User 1",
			Email:      "test1@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
	}

	// Create temporary file
	file, err := os.CreateTemp("", "samples-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	// Write samples to file
	data := struct {
		Samples []interface{} `json:"samples"`
	}{
		Samples: make([]interface{}, len(samples)),
	}
	for i, sample := range samples {
		data.Samples[i] = sample
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()

	// Remove error.log if it exists
	os.Remove("error.log")

	// Set invalid MongoDB connection string
	oldMongoURI := os.Getenv("MONGODB_URI")
	defer os.Setenv("MONGODB_URI", oldMongoURI)
	os.Setenv("MONGODB_URI", "mongodb://invalid:27017")

	// Run main with the test file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", file.Name()}

	// Run main in a goroutine since it calls os.Exit
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	// Wait for main to complete
	<-done

	// Check error.log exists and contains expected errors
	errorLog, err := os.ReadFile("error.log")
	if err != nil {
		t.Fatalf("Failed to read error.log: %v", err)
	}

	// Verify error log contains MongoDB connection errors
	if len(errorLog) == 0 {
		t.Error("Expected error.log to contain MongoDB connection errors")
	}
}

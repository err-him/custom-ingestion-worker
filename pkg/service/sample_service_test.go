package service

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"gohighlevel/pkg/ratelimiter"
	"gohighlevel/pkg/types"
	"gohighlevel/pkg/validator"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	samples map[string]types.Sample
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		samples: make(map[string]types.Sample),
	}
}

func (m *MockDatabase) Init() error {
	return nil
}

func (m *MockDatabase) Close() {}

func (m *MockDatabase) InsertSample(sample types.Sample) error {
	m.samples[sample.CustomerID] = sample
	return nil
}

func (m *MockDatabase) GetSample(customerID string) (types.Sample, error) {
	if sample, exists := m.samples[customerID]; exists {
		return sample, nil
	}
	return types.Sample{}, nil
}

func (m *MockDatabase) UpdateSample(sample types.Sample) error {
	m.samples[sample.CustomerID] = sample
	return nil
}

func (m *MockDatabase) DeleteSample(customerID string) error {
	delete(m.samples, customerID)
	return nil
}

// Helper function to create a test service
func setupTestService(t *testing.T) (*SampleService, *MockDatabase, func()) {
	// Create a temporary error.log file
	if err := os.Remove("error.log"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove error.log: %v", err)
	}

	mockDB := NewMockDatabase()
	v := validator.NewValidator(mockDB)
	r := ratelimiter.NewRateLimiter(5)
	s := NewSampleService(v, r, mockDB)

	cleanup := func() {
		os.Remove("error.log")
	}

	return s, mockDB, cleanup
}

// Helper function to create a test samples.json file
func createTestSamplesFile(t *testing.T, samples []CustomSample) string {
	file, err := os.CreateTemp("", "samples-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	data := struct {
		Samples []CustomSample `json:"samples"`
	}{
		Samples: samples,
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	file.Close()
	return file.Name()
}

func TestProcessSamplesFile(t *testing.T) {
	service, mockDB, cleanup := setupTestService(t)
	defer cleanup()

	// Create test samples
	samples := []CustomSample{
		{
			CustomerID: "1",
			Name:       "John Doe",
			Email:      "john@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "2",
			Name:       "Jane Smith",
			Email:      "jane@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "3",
			Name:       "Invalid Email",
			Email:      "invalid.email",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
	}

	// Create temporary file
	filePath := createTestSamplesFile(t, samples)
	defer os.Remove(filePath)

	// Process samples
	result, err := service.ProcessSamplesFile(filePath)
	if err != nil {
		t.Fatalf("ProcessSamplesFile() error = %v", err)
	}

	// Check results
	if result.SuccessCount != 2 {
		t.Errorf("Expected 2 successful samples, got %d", result.SuccessCount)
	}
	if result.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount)
	}

	// Verify successful samples were inserted
	for _, sample := range samples {
		if sample.Email == "invalid.email" {
			continue // Skip invalid sample
		}

		dbSample, err := mockDB.GetSample(sample.CustomerID)
		if err != nil {
			t.Errorf("Failed to get sample %s: %v", sample.CustomerID, err)
		}
		if dbSample.CustomerID != sample.CustomerID {
			t.Errorf("CustomerID mismatch: got %s, want %s", dbSample.CustomerID, sample.CustomerID)
		}
	}
}

func TestProcessSamplesRateLimit(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Create test samples with same customer ID to test rate limiting
	now := time.Now()
	samples := []CustomSample{}
	for i := 0; i < 10; i++ {
		samples = append(samples, CustomSample{
			CustomerID: "rate-test",
			Name:       "Rate Test",
			Email:      "rate@example.com",
			CreatedAt:  now.Format(time.RFC3339),
		})
	}

	// Create temporary file
	filePath := createTestSamplesFile(t, samples)
	defer os.Remove(filePath)

	// Process samples
	result, err := service.ProcessSamplesFile(filePath)
	if err != nil {
		t.Fatalf("ProcessSamplesFile() error = %v", err)
	}

	// Check results - should have 5 successful (rate limit) and 5 failed
	if result.SuccessCount != 5 {
		t.Errorf("Expected 5 successful samples (rate limit), got %d", result.SuccessCount)
	}
	if result.ErrorCount != 5 {
		t.Errorf("Expected 5 errors (rate limit exceeded), got %d", result.ErrorCount)
	}
}

func TestProcessSamplesInvalidData(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Create test samples with various invalid data
	samples := []CustomSample{
		{
			CustomerID: "", // Empty customer ID
			Name:       "Empty ID",
			Email:      "empty@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "invalid-email",
			Name:       "Invalid Email",
			Email:      "not-an-email",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "empty-name",
			Name:       "", // Empty name
			Email:      "name@example.com",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
		{
			CustomerID: "invalid-date",
			Name:       "Invalid Date",
			Email:      "date@example.com",
			CreatedAt:  "not-a-date",
		},
	}

	// Create temporary file
	filePath := createTestSamplesFile(t, samples)
	defer os.Remove(filePath)

	// Process samples
	result, err := service.ProcessSamplesFile(filePath)
	if err != nil {
		t.Fatalf("ProcessSamplesFile() error = %v", err)
	}

	// Check results - all should fail validation
	if result.SuccessCount != 0 {
		t.Errorf("Expected 0 successful samples, got %d", result.SuccessCount)
	}
	if result.ErrorCount != 4 {
		t.Errorf("Expected 4 errors, got %d", result.ErrorCount)
	}
}

func TestProcessSamplesTimeWindow(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Create test samples with timestamps 1 minute apart
	baseTime := time.Date(2024, 3, 26, 12, 0, 0, 0, time.UTC)
	samples := []CustomSample{}
	for i := 0; i < 10; i++ {
		samples = append(samples, CustomSample{
			CustomerID: "time-test",
			Name:       "Time Test",
			Email:      "time@example.com",
			CreatedAt:  baseTime.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
		})
	}

	// Create temporary file
	filePath := createTestSamplesFile(t, samples)
	defer os.Remove(filePath)

	// Process samples
	result, err := service.ProcessSamplesFile(filePath)
	if err != nil {
		t.Fatalf("ProcessSamplesFile() error = %v", err)
	}

	// Check results - should have 10 successful (different minutes)
	if result.SuccessCount != 10 {
		t.Errorf("Expected 10 successful samples, got %d", result.SuccessCount)
	}
	if result.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount)
	}
}

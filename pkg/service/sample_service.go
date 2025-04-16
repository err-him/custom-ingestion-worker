package service

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gohighlevel/pkg/db"
	"gohighlevel/pkg/ratelimiter"
	"gohighlevel/pkg/types"
	"gohighlevel/pkg/validator"
)

// SampleService orchestrates the processing of samples by coordinating
// between the validator, rate limiter, and database components.
type SampleService struct {
	validator   *validator.Validator
	rateLimiter *ratelimiter.RateLimiter
	db          db.Database
}

// NewSampleService creates a new sample service with the required dependencies.
func NewSampleService(v *validator.Validator, r *ratelimiter.RateLimiter, db db.Database) *SampleService {
	return &SampleService{
		validator:   v,
		rateLimiter: r,
		db:          db,
	}
}

// CustomSample is used for JSON decoding with custom time parsing.
// It matches the structure of samples in the JSON file.
type CustomSample struct {
	CustomerID string `json:"customerId"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt"`
}

// ProcessResult holds the statistics of sample processing.
type ProcessResult struct {
	SuccessCount int // Number of successfully processed samples
	ErrorCount   int // Number of samples that failed processing
}

// ProcessSamplesFile reads and processes samples from a JSON file.
// It handles file operations and JSON decoding, then delegates the
// actual processing to ProcessSamples.
func (s *SampleService) ProcessSamplesFile(filepath string) (ProcessResult, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("error opening file: %v", err)
	}
	defer jsonFile.Close()

	var data struct {
		Samples []CustomSample `json:"samples"`
	}
	if err := json.NewDecoder(jsonFile).Decode(&data); err != nil {
		return ProcessResult{}, fmt.Errorf("error decoding JSON: %v", err)
	}

	return s.ProcessSamples(data.Samples)
}

// ProcessSamples processes a batch of samples and returns the processing statistics.
// It tracks successful processing and uses the validator to count errors.
func (s *SampleService) ProcessSamples(samples []CustomSample) (ProcessResult, error) {
	successCount := 0
	for _, cs := range samples {
		if err := s.ProcessSample(cs); err == nil {
			successCount++
		}
	}
	return ProcessResult{
		SuccessCount: successCount,
		ErrorCount:   s.validator.GetErrorCount(),
	}, nil
}

// ProcessSample processes a single sample through the following steps:
// 1. Parses the creation timestamp
// 2. Validates the sample data
// 3. Checks rate limiting
// 4. Inserts the sample into the database
// Returns error if any step fails, nil on success.
func (s *SampleService) ProcessSample(cs CustomSample) error {
	// Parse time
	createdAt, err := time.Parse(time.RFC3339, cs.CreatedAt)
	if err != nil {
		s.validator.WriteErrorLog(cs.CustomerID, "invalid date format: "+cs.CreatedAt)
		return err
	}

	sample := types.Sample{
		CustomerID: cs.CustomerID,
		Email:      cs.Email,
		Name:       cs.Name,
		CreatedAt:  createdAt,
	}

	// Validate sample
	if err := s.validator.ValidateSample(sample); err != nil {
		return err // ValidateSample already logs the error
	}

	// Check rate limit
	if !s.rateLimiter.IsAllowed(sample.CustomerID, sample.CreatedAt) {
		s.validator.WriteErrorLog(sample.CustomerID, "rate limit exceeded")
		return fmt.Errorf("rate limit exceeded")
	}

	// Insert valid sample
	if err := s.db.InsertSample(sample); err != nil {
		s.validator.WriteErrorLog(sample.CustomerID, "failed to insert: "+err.Error())
		return err
	}

	return nil
}

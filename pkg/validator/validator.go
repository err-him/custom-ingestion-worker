package validator

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gohighlevel/pkg/interfaces"
	"gohighlevel/pkg/types"
)

// Validator handles the validation of sample data and error logging.
// It maintains a count of validation errors and provides methods to
// validate samples and log errors.
type Validator struct {
	db         interfaces.Database
	errorCount int // Tracks the number of validation errors encountered
}

// NewValidator creates a new validator instance with the given database connection.
func NewValidator(db interfaces.Database) *Validator {
	return &Validator{
		db:         db,
		errorCount: 0,
	}
}

// ValidateSample performs validation checks on a sample:
// 1. Checks if customer ID is present
// 2. Validates email format
// 3. Ensures name is not empty
// 4. Verifies timestamp is valid
// Returns error if any validation fails, nil otherwise.
func (v *Validator) ValidateSample(sample types.Sample) error {
	// Validate customer ID
	if sample.CustomerID == "" {
		v.writeErrorLog(sample.CustomerID, "customer_id is required")
		return fmt.Errorf("customer_id is required")
	}

	// Validate email
	if !isValidEmail(sample.Email) {
		v.writeErrorLog(sample.CustomerID, "invalid email format")
		return fmt.Errorf("invalid email format")
	}

	// Validate name
	if sample.Name == "" {
		v.writeErrorLog(sample.CustomerID, "name is required")
		return fmt.Errorf("name is required")
	}

	// Validate timestamps
	if sample.CreatedAt.IsZero() {
		v.writeErrorLog(sample.CustomerID, "created_at is required")
		return fmt.Errorf("created_at is required")
	}

	if sample.UpdatedAt.IsZero() {
		sample.UpdatedAt = time.Now()
	}

	return nil
}

// WriteErrorLog is a public method to write errors to the log file.
// It's used by other components that need to log validation errors.
func (v *Validator) WriteErrorLog(customerID, reason string) error {
	return v.writeErrorLog(customerID, reason)
}

// writeErrorLog writes an error entry to the error.log file and increments the error counter.
// Each error entry includes:
// - Status (always "error")
// - Customer ID
// - Error reason
// - Timestamp
func (v *Validator) writeErrorLog(customerID, reason string) error {
	file, err := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open error log: %v", err)
	}
	defer file.Close()

	errorLog := fmt.Sprintf(`{
	"status": "error",
	"customerId": "%s",
	"reason": "%s",
	"createdAt": "%s"
}
`, customerID, reason, time.Now().Format(time.RFC3339))

	_, err = file.WriteString(errorLog)
	if err != nil {
		return fmt.Errorf("failed to write to error log: %v", err)
	}

	v.errorCount++
	return nil
}

// GetErrorCount returns the total number of validation errors encountered.
func (v *Validator) GetErrorCount() int {
	return v.errorCount
}

// isValidEmail checks if the email string matches a valid email format.
// The regex pattern ensures:
// - Local part contains letters, numbers, and common special characters
// - Domain part has a valid structure with TLD
// - TLD is 2-4 characters long
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

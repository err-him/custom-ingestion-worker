package main

import (
	"fmt"
	"log"
	"os"

	"gohighlevel/pkg/db"
	"gohighlevel/pkg/ratelimiter"
	"gohighlevel/pkg/service"
	"gohighlevel/pkg/validator"
)

// rateLimit defines the maximum number of requests allowed per customer per minute
const rateLimit = 5

// main is the entry point of the application. It:
// 1. Sets up the error logging
// 2. Initializes the MongoDB connection
// 3. Creates validator, rate limiter, and sample service instances
// 4. Processes the samples from samples.json
// 5. Reports the processing results
func main() {
	// Remove error.log file if it exists to start fresh
	if err := os.Remove("error.log"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove old error.log: %v\n", err)
	}

	// Initialize MongoDB connection
	mongoDB := db.NewMongoDatabase()
	if err := mongoDB.Init(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()

	// Initialize components with their dependencies
	v := validator.NewValidator(mongoDB)                     // Validator for sample data
	r := ratelimiter.NewRateLimiter(rateLimit)               // Rate limiter to prevent too many requests, in this case 5 requests per customer per minute
	sampleService := service.NewSampleService(v, r, mongoDB) // Service to process samples

	// Process all samples from the JSON file
	result, err := sampleService.ProcessSamplesFile("samples.json")
	if err != nil {
		log.Fatalf("Failed to process samples: %v", err)
	}

	// Print processing statistics
	fmt.Printf("Total samples: %d\n", result.SuccessCount+result.ErrorCount)
	fmt.Printf("Successfully processed %d samples\n", result.SuccessCount)
	fmt.Printf("Failed to process %d samples\n", result.ErrorCount)
}

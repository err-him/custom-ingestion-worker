package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gohighlevel/pkg/types"
)

func setupTestDB(tb testing.TB) (*MongoDatabase, func()) {
	db := NewMongoDatabase()
	if err := db.Init(); err != nil {
		tb.Fatalf("Failed to initialize test database: %v", err)
	}

	// Return cleanup function
	return db, func() {
		ctx := context.Background()
		if err := db.collection.Drop(ctx); err != nil {
			tb.Errorf("Failed to clean up test database: %v", err)
		}
		db.Close()
	}
}

func TestMongoOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	sample := types.Sample{
		CustomerID: "test123",
		Email:      "test@example.com",
		Name:       "Test User",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Test InsertSample
	err := db.InsertSample(sample)
	if err != nil {
		t.Errorf("InsertSample() error = %v", err)
	}

	// Verify the sample was inserted
	ctx := context.Background()
	var result struct {
		CustomerID string    `bson:"customerId"`
		Email      string    `bson:"email"`
		Name       string    `bson:"name"`
		CreatedAt  time.Time `bson:"createdAt"`
	}
	err = db.collection.FindOne(ctx, map[string]string{"customerId": sample.CustomerID}).Decode(&result)
	if err != nil {
		t.Errorf("Failed to find inserted sample: %v", err)
	}

	if result.CustomerID != sample.CustomerID {
		t.Errorf("CustomerID mismatch: got %v, want %v", result.CustomerID, sample.CustomerID)
	}
	if result.Email != sample.Email {
		t.Errorf("Email mismatch: got %v, want %v", result.Email, sample.Email)
	}
	if result.Name != sample.Name {
		t.Errorf("Name mismatch: got %v, want %v", result.Name, sample.Name)
	}
}

func BenchmarkInsertSample(b *testing.B) {
	db, cleanup := setupTestDB(b)
	defer cleanup()

	sample := types.Sample{
		CustomerID: "bench123",
		Email:      "bench@example.com",
		Name:       "Bench User",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sample.CustomerID = fmt.Sprintf("bench%d", i)
		if err := db.InsertSample(sample); err != nil {
			b.Fatal(err)
		}
	}
}

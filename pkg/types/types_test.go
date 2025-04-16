package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSampleJSON(t *testing.T) {
	now := time.Now()
	sample := Sample{
		CustomerID: "cust123",
		Email:      "test@example.com",
		Name:       "Test User",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Test JSON marshaling
	data, err := json.Marshal(sample)
	if err != nil {
		t.Errorf("Failed to marshal Sample: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Sample
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("Failed to unmarshal Sample: %v", err)
	}

	// Verify fields
	if decoded.CustomerID != sample.CustomerID {
		t.Errorf("CustomerID mismatch: got %v, want %v", decoded.CustomerID, sample.CustomerID)
	}
	if decoded.Email != sample.Email {
		t.Errorf("Email mismatch: got %v, want %v", decoded.Email, sample.Email)
	}
	if decoded.Name != sample.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, sample.Name)
	}
}

func BenchmarkSampleJSONMarshal(b *testing.B) {
	sample := Sample{
		CustomerID: "cust123",
		Email:      "test@example.com",
		Name:       "Test User",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(sample)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSampleJSONUnmarshal(b *testing.B) {
	sample := Sample{
		CustomerID: "cust123",
		Email:      "test@example.com",
		Name:       "Test User",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(sample)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded Sample
		if err := json.Unmarshal(data, &decoded); err != nil {
			b.Fatal(err)
		}
	}
}

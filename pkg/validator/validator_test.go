package validator

import (
	"testing"
	"time"

	"gohighlevel/pkg/types"
)

type mockDB struct{}

func (m *mockDB) Init() error                            { return nil }
func (m *mockDB) Close()                                 {}
func (m *mockDB) InsertSample(sample types.Sample) error { return nil }

func TestValidateSample(t *testing.T) {
	validator := NewValidator(&mockDB{})

	tests := []struct {
		name    string
		sample  types.Sample
		wantErr bool
	}{
		{
			name: "Valid sample",
			sample: types.Sample{
				CustomerID: "cust123",
				Email:      "test@example.com",
				Name:       "Test User",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Missing CustomerID",
			sample: types.Sample{
				Email:     "test@example.com",
				Name:      "Test User",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Invalid email",
			sample: types.Sample{
				CustomerID: "cust123",
				Email:      "invalid-email",
				Name:       "Test User",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: true,
		},
		{
			name: "Missing name",
			sample: types.Sample{
				CustomerID: "cust123",
				Email:      "test@example.com",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSample(tt.sample)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSample() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkValidateSample(b *testing.B) {
	validator := NewValidator(&mockDB{})
	sample := types.Sample{
		CustomerID: "cust123",
		Email:      "test@example.com",
		Name:       "Test User",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSample(sample)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid email", "test@example.com", true},
		{"Invalid email - no @", "testexample.com", false},
		{"Invalid email - no domain", "test@", false},
		{"Invalid email - spaces", "test @example.com", false},
		{"Invalid email - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkIsValidEmail(b *testing.B) {
	email := "test@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isValidEmail(email)
	}
}

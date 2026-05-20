package domain

import (
	"testing"
)

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name          string
		formula       string
		birthYear     int
		referenceYear int
		expectedFee   int
		expectError   bool
	}{
		{
			name:          "Fixed amount",
			formula:       "FIXED:50000",
			referenceYear: 2026,
			expectedFee:   50000,
		},
		{
			name:          "Age based - under limit",
			formula:       "AGE_BASED:18:20000:50000",
			birthYear:     2015, // 11 år i 2026
			referenceYear: 2026,
			expectedFee:   20000,
		},
		{
			name:          "Age based - over limit",
			formula:       "AGE_BASED:18:20000:50000",
			birthYear:     2000, // 26 år i 2026
			referenceYear: 2026,
			expectedFee:   50000,
		},
		{
			name:          "Age based - exactly at limit",
			formula:       "AGE_BASED:18:20000:50000",
			birthYear:     2008, // 18 år i 2026
			referenceYear: 2026,
			expectedFee:   50000,
		},
		{
			name:          "Invalid formula",
			formula:       "INVALID:123",
			referenceYear: 2026,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{ID: "u1"}
			m := &Membership{
				ID: "m1",
				Metadata: map[string]any{},
			}
			if tt.birthYear != 0 {
				m.Metadata["birth_year"] = tt.birthYear
			}

			fee, err := CalculateFee(u, m, tt.formula, tt.referenceYear)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !tt.expectError && fee != tt.expectedFee {
				t.Errorf("expected fee: %d, got: %d", tt.expectedFee, fee)
			}
		})
	}
}

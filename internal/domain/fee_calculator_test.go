package domain

import (
	"testing"
)

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name          string
		member        *Member
		formula       string
		referenceYear int
		expected      int
		wantErr       bool
	}{
		{
			name: "Fixed amount",
			member: &Member{},
			formula: "FIXED:50000",
			referenceYear: 2026,
			expected: 50000,
			wantErr: false,
		},
		{
			name: "Age based - under limit",
			member: &Member{
				Metadata: map[string]any{"birth_year": 2010}, // 16 år i 2026
			},
			formula: "AGE_BASED:18:20000:50000",
			referenceYear: 2026,
			expected: 20000,
			wantErr: false,
		},
		{
			name: "Age based - over limit",
			member: &Member{
				Metadata: map[string]any{"birth_year": 2000}, // 26 år i 2026
			},
			formula: "AGE_BASED:18:20000:50000",
			referenceYear: 2026,
			expected: 50000,
			wantErr: false,
		},
		{
			name: "Invalid formula",
			member: &Member{},
			formula: "INVALID:123",
			referenceYear: 2026,
			expected: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateFee(tt.member, tt.formula, tt.referenceYear)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("CalculateFee() = %v, want %v", got, tt.expected)
			}
		})
	}
}

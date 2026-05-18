package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// CalculateFee beregner kontingent basert på en formel og medlemsdata.
// Støttede formler:
// - "FIXED:beløp" (f.eks. "FIXED:50000" for 500 kr)
// - "AGE_BASED:alder_grense:under_beløp:over_beløp" (f.eks. "AGE_BASED:18:20000:50000")
func CalculateFee(m *Member, formula string, referenceYear int) (int, error) {
	if formula == "" {
		return 0, fmt.Errorf("no formula defined")
	}

	parts := strings.Split(formula, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid formula format")
	}

	switch parts[0] {
	case "FIXED":
		amount, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid fixed amount: %w", err)
		}
		return amount, nil

	case "AGE_BASED":
		if len(parts) != 4 {
			return 0, fmt.Errorf("AGE_BASED requires 3 parameters (limit, under, over)")
		}
		limit, _ := strconv.Atoi(parts[1])
		underAmount, _ := strconv.Atoi(parts[2])
		overAmount, _ := strconv.Atoi(parts[3])

		// Finn fødselsår fra metadata hvis det finnes
		birthYearVal, ok := m.Metadata["birth_year"]
		if !ok {
			// Hvis vi ikke vet alder, faller vi tilbake til over_beløp for sikkerhets skyld
			return overAmount, nil
		}

		var birthYear int
		switch v := birthYearVal.(type) {
		case int:
			birthYear = v
		case float64:
			birthYear = int(v)
		case string:
			birthYear, _ = strconv.Atoi(v)
		}

		age := referenceYear - birthYear
		if age < limit {
			return underAmount, nil
		}
		return overAmount, nil

	default:
		return 0, fmt.Errorf("unknown formula type: %s", parts[0])
	}
}

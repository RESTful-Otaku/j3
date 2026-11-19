package units_test

import (
	"math"
	"testing"

	"github.com/restful-otaku/j3/internal/units"
)

func FuzzToKg(f *testing.F) {
	f.Add(75.0, "kg")
	f.Add(165.0, "lb")
	f.Add(11.0, "stone")
	f.Add(0.0, "kg")
	f.Add(75.0, "furlong")
	f.Fuzz(func(t *testing.T, value float64, unit string) {
		got, err := units.ToKg(value, unit)
		if err != nil {
			return
		}
		if math.IsNaN(got) || math.IsInf(got, 0) || got <= 0 {
			t.Fatalf("ToKg(%v, %q) produced invalid %v", value, unit, got)
		}
		if unit == "kg" || unit == "metric" {
			if got != value {
				t.Fatalf("identity conversion changed value: %v != %v", got, value)
			}
		}
	})
}

func FuzzToCm(f *testing.F) {
	f.Add(175.0, "cm")
	f.Add(1.8, "m")
	f.Add(70.0, "in")
	f.Add(5.9, "ft")
	f.Fuzz(func(t *testing.T, value float64, unit string) {
		got, err := units.ToCm(value, unit)
		if err != nil {
			return
		}
		if math.IsNaN(got) || math.IsInf(got, 0) || got <= 0 || got > 100000 {
			t.Fatalf("ToCm(%v, %q) produced implausible %v", value, unit, got)
		}
	})
}

func FuzzFeetInches(f *testing.F) {
	f.Add(5.0, 9.0)
	f.Add(6.0, 0.0)
	f.Add(-1.0, 5.0)
	f.Add(5.0, 13.0)
	f.Fuzz(func(t *testing.T, feet, inches float64) {
		got, err := units.FeetInchesToCm(feet, inches)
		if err != nil {
			return
		}
		if got < 0 || got > 100000 {
			t.Fatalf("FeetInchesToCm(%v,%v) = %v", feet, inches, got)
		}
	})
}

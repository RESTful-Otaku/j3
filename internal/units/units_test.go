package units_test

import (
	"math"
	"testing"

	"github.com/restful-otaku/j3/internal/units"
)

func TestToKg(t *testing.T) {
	cases := []struct {
		value float64
		unit  string
		want  float64
	}{
		{75, "kg", 75},
		{165, "lb", 74.84274105},
		{11, "stone", 69.85322498},
	}
	for _, c := range cases {
		got, err := units.ToKg(c.value, c.unit)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("ToKg(%v, %q) = %v, want %v", c.value, c.unit, got, c.want)
		}
	}
	if _, err := units.ToKg(1, "furlong"); err == nil {
		t.Error("expected error for unknown unit")
	}
}

func TestToCm(t *testing.T) {
	cases := []struct {
		value float64
		unit  string
		want  float64
	}{
		{175, "cm", 175},
		{1.8, "m", 180},
		{70, "in", 177.8},
	}
	for _, c := range cases {
		got, err := units.ToCm(c.value, c.unit)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("ToCm(%v, %q) = %v, want %v", c.value, c.unit, got, c.want)
		}
	}
	got, err := units.FeetInchesToCm(5, 9)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-175.26) > 1e-6 {
		t.Errorf("FeetInchesToCm = %v", got)
	}
}

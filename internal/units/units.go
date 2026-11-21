// Package units converts user-entered imperial values to canonical metric.
package units

import (
	"fmt"
	"math"
	"strings"
)

// ToKg converts a weight in the given unit to kilograms.
func ToKg(value float64, unit string) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return 0, fmt.Errorf("invalid weight %v", value)
	}
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "kg", "kilogram", "kilograms", "metric":
		return value, nil
	case "lb", "lbs", "pound", "pounds":
		return value * 0.45359237, nil
	case "st", "stone", "stones":
		return value * 6.35029318, nil
	default:
		return 0, fmt.Errorf("unknown weight unit %q", unit)
	}
}

// ToCm converts a height in the given unit to centimetres.
func ToCm(value float64, unit string) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return 0, fmt.Errorf("invalid height %v", value)
	}
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "cm", "centimetre", "centimetres", "metric":
		return value, nil
	case "m", "meter", "meters", "metre", "metres":
		return value * 100, nil
	case "in", "inch", "inches":
		return value * 2.54, nil
	case "ft", "foot", "feet":
		return value * 30.48, nil
	default:
		return 0, fmt.Errorf("unknown height unit %q", unit)
	}
}

// FeetInchesToCm combines imperial feet and inches into centimetres.
func FeetInchesToCm(feet, inches float64) (float64, error) {
	if math.IsNaN(feet) || math.IsNaN(inches) || math.IsInf(feet, 0) || math.IsInf(inches, 0) ||
		feet < 0 || inches < 0 || inches >= 12 {
		return 0, fmt.Errorf("invalid height %v ft %v in", feet, inches)
	}
	return feet*30.48 + inches*2.54, nil
}

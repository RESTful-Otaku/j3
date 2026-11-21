// Package rules contains pure, deterministic eligibility and pool algorithms.
package rules

import (
	"sort"
	"time"
)

// Stats are the canonical metric values captured at sign-up.
type Stats struct {
	DOB       string
	WeightKg  float64
	HeightCm  float64
	BeltOrder int
	Gender    string
}

// Category is the data needed to decide eligibility.
type Category struct {
	ID         string
	Name       string
	Kind       string
	Gender     string
	Min        float64
	Max        float64
	BeltRankID string
}

// Eligibility describes an offered category and why it was offered.
type Eligibility struct {
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	Strict     bool   `json:"strict"`
	MoveUp     bool   `json:"moveUp"`
	Reason     string `json:"reason"`
	Capacity   int    `json:"capacity"`
	Spaces     int    `json:"spaces"`
	Full       bool   `json:"full"`
}

func ageOn(dob string, on time.Time) (int, bool) {
	birth, err := time.Parse("2006-01-02", dob)
	if err != nil || birth.After(on) {
		return 0, false
	}
	age := on.Year() - birth.Year()
	if on.Month() < birth.Month() || (on.Month() == birth.Month() && on.Day() < birth.Day()) {
		age--
	}
	return age, true
}

func genderMatches(categoryGender, athleteGender string) bool {
	if categoryGender == "" || categoryGender == "open" || categoryGender == "coed" {
		return true
	}
	if athleteGender == "" || athleteGender == "open" || athleteGender == "coed" {
		return true
	}
	return categoryGender == athleteGender
}

func offer(id, name, reason string, strict, moveUp bool) Eligibility {
	return Eligibility{CategoryID: id, Name: name, Strict: strict, MoveUp: moveUp, Reason: reason}
}

// Compatible returns categories the athlete fits. Strict matches come first;
// when consentMoveUp is true, nearby heavier/older divisions are appended.
func Compatible(categories []Category, stats Stats, on time.Time, consentMoveUp bool) []Eligibility {
	age, hasAge := ageOn(stats.DOB, on)
	result := make([]Eligibility, 0, len(categories))
	type candidate struct {
		cat Category
		min float64
		why string
	}
	moveUps := []candidate{}

	for _, cat := range categories {
		if !genderMatches(cat.Gender, stats.Gender) {
			continue
		}
		switch {
		case cat.Kind == "belt" || (cat.BeltRankID != "" && cat.Max == 0):
			if stats.BeltOrder == 0 {
				continue
			}
			if stats.BeltOrder >= int(cat.Min) && (cat.Max == 0 || stats.BeltOrder <= int(cat.Max)) {
				result = append(result, offer(cat.ID, cat.Name, "belt grade in range", true, false))
			}
		case cat.Kind == "age":
			if !hasAge {
				continue
			}
			if age >= int(cat.Min) && age <= int(cat.Max) {
				result = append(result, offer(cat.ID, cat.Name, "age in range", true, false))
			} else if consentMoveUp && age < int(cat.Min) {
				moveUps = append(moveUps, candidate{cat, cat.Min, "older division with consent"})
			}
		case cat.Kind == "height":
			if stats.HeightCm <= 0 {
				continue
			}
			if stats.HeightCm >= cat.Min && stats.HeightCm <= cat.Max {
				result = append(result, offer(cat.ID, cat.Name, "height in range", true, false))
			} else if consentMoveUp && stats.HeightCm < cat.Min {
				moveUps = append(moveUps, candidate{cat, cat.Min, "taller division with consent"})
			}
		case cat.Kind == "weight" || (cat.Kind == "combined" && cat.Max > 0):
			if stats.WeightKg <= 0 {
				continue
			}
			if stats.WeightKg >= cat.Min && stats.WeightKg <= cat.Max {
				result = append(result, offer(cat.ID, cat.Name, "weight in range", true, false))
			} else if consentMoveUp && stats.WeightKg < cat.Min {
				moveUps = append(moveUps, candidate{cat, cat.Min, "heavier division with consent"})
			}
		default:
			// Open divisions (e.g. Kata) accept everyone.
			result = append(result, offer(cat.ID, cat.Name, "open division", true, false))
		}
	}

	if consentMoveUp && len(moveUps) > 0 {
		sort.Slice(moveUps, func(i, j int) bool { return moveUps[i].min < moveUps[j].min })
		for _, c := range moveUps {
			result = append(result, offer(c.cat.ID, c.cat.Name, c.why, false, true))
		}
	}
	return result
}

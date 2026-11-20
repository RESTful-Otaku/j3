package rules

import (
	"math"
	"sort"
)

// SeedEntrants orders fighters by seed (higher first) while spreading clubs
// so same-club fighters do not cluster at the top of a draw.
func SeedEntrants(entrants []Entrant, separateClubs bool) []Entrant {
	ordered := append([]Entrant(nil), entrants...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Seed != ordered[j].Seed {
			return ordered[i].Seed > ordered[j].Seed
		}
		return ordered[i].ID < ordered[j].ID
	})
	if !separateClubs {
		return ordered
	}
	result := make([]Entrant, 0, len(ordered))
	used := map[string]int{}
	for len(ordered) > 0 {
		picked := 0
		best := math.MaxInt
		for i, candidate := range ordered {
			load := used[candidate.ClubID]
			if load < best {
				best, picked = load, i
			}
		}
		chosen := ordered[picked]
		used[chosen.ClubID]++
		result = append(result, chosen)
		ordered = append(ordered[:picked], ordered[picked+1:]...)
	}
	return result
}

// WithByes pads a knockout field to a power of two using BYE markers so
// seeded fighters receive fair first-round byes.
func WithByes(ids []string) []string {
	if len(ids) == 0 {
		return ids
	}
	size := 1
	for size < len(ids) {
		size *= 2
	}
	padded := append([]string(nil), ids...)
	for len(padded) < size {
		padded = append(padded, "BYE")
	}
	return padded
}

// TieBreakStats are the comparable criteria used by an ordered tie-break chain.
type TieBreakStats struct {
	Wins, Losses, Draws, Submissions int
	PointsScored, PointsConceded     int
}

// TieBreak compares two athletes using an ordered chain of criterion names.
// Returns -1 if a wins, 1 if b wins, 0 if still level.
func TieBreak(a, b TieBreakStats, chain []string) int {
	if len(chain) == 0 {
		chain = []string{"wins", "submissions", "points_scored", "head_to_head", "points_conceded"}
	}
	for _, criterion := range chain {
		switch criterion {
		case "wins":
			if a.Wins != b.Wins {
				if a.Wins > b.Wins {
					return -1
				}
				return 1
			}
		case "submissions":
			if a.Submissions != b.Submissions {
				if a.Submissions > b.Submissions {
					return -1
				}
				return 1
			}
		case "points_scored":
			if a.PointsScored != b.PointsScored {
				if a.PointsScored > b.PointsScored {
					return -1
				}
				return 1
			}
		case "points_conceded":
			if a.PointsConceded != b.PointsConceded {
				if a.PointsConceded < b.PointsConceded {
					return -1
				}
				return 1
			}
		case "draws":
			if a.Draws != b.Draws {
				if a.Draws > b.Draws {
					return -1
				}
				return 1
			}
		}
	}
	return 0
}

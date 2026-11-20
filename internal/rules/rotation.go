package rules

// RotationPair is one sparring pairing on one mat in one round.
type RotationPair struct {
	Round int    `json:"round"`
	Mat   string `json:"mat"`
	Red   string `json:"red"`
	Blue  string `json:"blue"`
}

// BuildRotation creates a round-robin sparring rotation using the circle
// method. Every fighter meets every other fighter at most once, and pairs are
// spread across the available mats.
func BuildRotation(fighters, mats []string, rounds int) []RotationPair {
	if len(fighters) < 2 || len(mats) == 0 {
		return []RotationPair{}
	}
	ids := append([]string(nil), fighters...)
	if len(ids)%2 == 1 {
		ids = append(ids, "BYE")
	}
	totalRounds := len(ids) - 1
	if rounds > 0 && rounds < totalRounds {
		totalRounds = rounds
	}
	result := []RotationPair{}
	fixed := ids[0]
	rotating := append([]string(nil), ids[1:]...)
	for round := 1; round <= totalRounds; round++ {
		pairs := [][2]string{{fixed, rotating[0]}}
		for i := 1; i <= (len(rotating)-1)/2; i++ {
			pairs = append(pairs, [2]string{rotating[i], rotating[len(rotating)-i]})
		}
		for i, pair := range pairs {
			if pair[0] == "BYE" || pair[1] == "BYE" {
				continue
			}
			result = append(result, RotationPair{
				Round: round, Mat: mats[i%len(mats)], Red: pair[0], Blue: pair[1],
			})
		}
		rotating = append([]string{rotating[len(rotating)-1]}, rotating[:len(rotating)-1]...)
	}
	return result
}

// Package simulator runs a deterministic tournament from signup to medals.
// It powers QA, "what-if" planning and the animated replay view.
package simulator

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/restful-otaku/j3/internal/rules"
)

type Fighter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Club  string `json:"club"`
	Belt  string `json:"belt"`
	Stage string `json:"stage"`
}

type Bout struct {
	ID         string `json:"id"`
	Stage      string `json:"stage"`
	Round      string `json:"round"`
	RedID      string `json:"redId"`
	BlueID     string `json:"blueId"`
	Outcome    string `json:"outcome"`
	Method     string `json:"method"`
	PointsRed  int    `json:"pointsRed"`
	PointsBlue int    `json:"pointsBlue"`
	Sequence   int    `json:"sequence"`
}

type Medal struct {
	FighterID string `json:"fighterId"`
	Kind      string `json:"kind"`
}

type Standing struct {
	FighterID string `json:"fighterId"`
	Name      string `json:"name"`
	Club      string `json:"club"`
	Wins      int    `json:"wins"`
	Losses    int    `json:"losses"`
	Draws     int    `json:"draws"`
	Points    int    `json:"points"`
}

type Result struct {
	Bouts         []Bout         `json:"bouts"`
	Medals        []Medal        `json:"medals"`
	Standings     []Standing     `json:"standings"`
	TopMoves      map[string]int `json:"topMoves"`
	ClubPoints    map[string]int `json:"clubPoints"`
	ClubMedals    map[string]int `json:"clubMedals"`
	FighterMedals map[string]int `json:"fighterMedals"`
	Seed          int64          `json:"seed"`
}

type record struct{ wins, losses, draws, points int }

// Run simulates pools, round-robin, knockout and a bronze decider. The same
// seed always yields the same event, which is what makes verification possible.
func Run(seed int64, count int) Result {
	if count < 2 {
		count = 8
	}
	rng := rand.New(rand.NewSource(seed))
	fighters := make([]Fighter, count)
	clubs := []string{"Riverside", "Hilltop", "City Warriors", "Sakura"}
	belts := []string{"White", "Blue", "Purple", "Brown", "Black"}
	for i := range fighters {
		fighters[i] = Fighter{
			ID:   fmt.Sprintf("f%03d", i+1),
			Name: fmt.Sprintf("Fighter %02d", i+1),
			Club: clubs[i%len(clubs)],
			Belt: belts[i%len(belts)],
		}
	}
	entrants := make([]rules.Entrant, count)
	for i, f := range fighters {
		entrants[i] = rules.Entrant{ID: f.ID, ClubID: f.Club, Seed: i + 1}
	}
	pools := rules.GeneratePools(entrants, rules.PoolConstraints{
		TargetPoolSize: 5, MaxPools: 12, SameClubSeparation: true, Seed: seed,
	})

	result := Result{
		Bouts: []Bout{}, Medals: []Medal{}, TopMoves: map[string]int{},
		ClubPoints: map[string]int{}, ClubMedals: map[string]int{},
		FighterMedals: map[string]int{}, Seed: seed,
	}
	records := map[string]*record{}
	for _, f := range fighters {
		records[f.ID] = &record{}
	}
	sequence := 0

	addBout := func(stage, round, a, b string) {
		sequence++
		methods := []string{"points", "submission", "points", "points", "forfeit", "draw"}
		method := methods[rng.Intn(len(methods))]
		redPoints, bluePoints := rng.Intn(9)+1, rng.Intn(8)
		outcome := "win_red"
		switch method {
		case "draw":
			outcome, redPoints, bluePoints = "draw", 0, 0
		case "submission":
			outcome, redPoints, bluePoints = "submission_red", 0, 0
		case "forfeit":
			outcome = "forfeit_red"
		}
		if rng.Intn(2) == 0 && outcome != "draw" && !strings.Contains(outcome, "red") {
			outcome = replaceCorner(outcome)
			redPoints, bluePoints = bluePoints, redPoints
		}
		result.Bouts = append(result.Bouts, Bout{
			ID: fmt.Sprintf("b%03d", sequence), Stage: stage, Round: round,
			RedID: a, BlueID: b, Outcome: outcome, Method: method,
			PointsRed: redPoints, PointsBlue: bluePoints, Sequence: sequence,
		})
		result.TopMoves[method]++
		ra, rb := records[a], records[b]
		if outcome == "draw" {
			ra.draws++
			rb.draws++
			ra.points++
			rb.points++
		} else if strings.Contains(outcome, "red") {
			ra.wins++
			rb.losses++
			ra.points += 3
			result.ClubPoints[clubOf(fighters, a)] += 3
		} else {
			rb.wins++
			ra.losses++
			rb.points += 3
			result.ClubPoints[clubOf(fighters, b)] += 3
		}
	}

	// Pool round robin.
	advance := []string{}
	for _, pool := range pools {
		ids := pool.Members
		for i := range ids {
			for j := i + 1; j < len(ids); j++ {
				addBout("pool", pool.Name, ids[i], ids[j])
			}
		}
		ranked := append([]string(nil), ids...)
		sort.SliceStable(ranked, func(i, j int) bool { return records[ranked[i]].points > records[ranked[j]].points })
		if len(ranked) > 0 {
			advance = append(advance, ranked[0])
		}
		if len(ranked) > 1 {
			advance = append(advance, ranked[1])
		}
	}
	if len(advance) > 8 {
		advance = advance[:8]
	}

	semiLosers := []string{}
	var finalLoser string
	for len(advance) > 1 {
		roundName := "Knockout"
		switch len(advance) {
		case 8:
			roundName = "Quarter-final"
		case 4:
			roundName = "Semi-final"
		case 2:
			roundName = "Final"
		}
		next := []string{}
		for i := 0; i+1 < len(advance); i += 2 {
			addBout("knockout", roundName, advance[i], advance[i+1])
			last := result.Bouts[len(result.Bouts)-1]
			loser := advance[i+1]
			if strings.Contains(last.Outcome, "red") {
				next = append(next, advance[i])
			} else {
				next = append(next, advance[i+1])
				loser = advance[i]
			}
			if roundName == "Semi-final" {
				semiLosers = append(semiLosers, loser)
			}
			if roundName == "Final" {
				finalLoser = loser
			}
		}
		advance = next
	}
	if len(advance) > 0 {
		result.Medals = append(result.Medals, Medal{advance[0], "gold"}, Medal{advance[0], "champion"})
	}
	if finalLoser != "" {
		result.Medals = append(result.Medals, Medal{finalLoser, "silver"})
	}
	if len(semiLosers) == 2 {
		// No-rematch rule: if they already met, award two bronzes instead.
		plan := rules.ThirdPlacePlan(met(result.Bouts, semiLosers[0], semiLosers[1]), true)
		if plan == "rematch" {
			addBout("bronze", "Bronze decider", semiLosers[0], semiLosers[1])
			last := result.Bouts[len(result.Bouts)-1]
			if last.Outcome == "draw" {
				result.Medals = append(result.Medals, Medal{semiLosers[0], "bronze"}, Medal{semiLosers[1], "bronze"})
			} else if strings.Contains(last.Outcome, "red") {
				result.Medals = append(result.Medals, Medal{semiLosers[0], "bronze"})
			} else {
				result.Medals = append(result.Medals, Medal{semiLosers[1], "bronze"})
			}
		} else {
			result.Medals = append(result.Medals, Medal{semiLosers[0], "bronze"}, Medal{semiLosers[1], "bronze"})
		}
	}

	for _, fighter := range fighters {
		rec := records[fighter.ID]
		result.Standings = append(result.Standings, Standing{
			fighter.ID, fighter.Name, fighter.Club, rec.wins, rec.losses, rec.draws, rec.points,
		})
	}
	sort.SliceStable(result.Standings, func(i, j int) bool {
		return result.Standings[i].Points > result.Standings[j].Points
	})
	for _, medal := range result.Medals {
		if medal.Kind == "gold" || medal.Kind == "silver" || medal.Kind == "bronze" {
			club := clubOf(fighters, medal.FighterID)
			result.ClubMedals[club]++
			result.FighterMedals[medal.FighterID]++
		}
	}
	return result
}

func replaceCorner(value string) string {
	if strings.Contains(value, "red") {
		return "win_blue"
	}
	return "win_red"
}

func met(bouts []Bout, a, b string) bool {
	for _, bout := range bouts {
		if (bout.RedID == a && bout.BlueID == b) || (bout.RedID == b && bout.BlueID == a) {
			return true
		}
	}
	return false
}

func clubOf(fighters []Fighter, id string) string {
	for _, f := range fighters {
		if f.ID == id {
			return f.Club
		}
	}
	return "Independent"
}

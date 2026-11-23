// Package simulator runs a deterministic tournament from signup to medals.
// It powers QA, "what-if" planning and the animated replay view.
package simulator

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/restful-otaku/j3/internal/rules"
	"github.com/restful-otaku/j3/internal/techniques"
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
	Outcome    string   `json:"outcome"`
	Method     string   `json:"method"`
	Techniques []string `json:"techniques,omitempty"`
	PointsRed  int      `json:"pointsRed"`
	PointsBlue int      `json:"pointsBlue"`
	Sequence   int      `json:"sequence"`
}

type Medal struct {
	FighterID  string `json:"fighterId"`
	Kind       string `json:"kind"`
	Category   string `json:"category,omitempty"`
	Discipline string `json:"discipline,omitempty"`
	EventType  string `json:"eventType,omitempty"`
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

var scoringTechniques = func() []techniques.Technique {
	out := []techniques.Technique{}
	for _, tech := range techniques.All() {
		if tech.Kind == "score" && !tech.InstantWin && tech.Points > 0 {
			out = append(out, tech)
		}
	}
	return out
}()

var finishingTechniques = func() []techniques.Technique {
	out := []techniques.Technique{}
	for _, tech := range techniques.All() {
		if tech.InstantWin && tech.Kind == "submission" {
			out = append(out, tech)
		}
	}
	return out
}()

var scoreLabels = []struct{ key, label string }{
	{"strike", "strikes"}, {"kick", "kicks"}, {"throw", "throws"},
	{"takedown", "takedowns"}, {"sweep", "sweeps"}, {"reap", "reaps"},
	{"hook", "hooks"}, {"hold", "holds"}, {"ground", "ground work"},
	{"control", "control"}, {"break", "breaks"}, {"ippon", "ippon"},
	{"wazaari", "waza-ari"},
}

func scoreSide(rng *rand.Rand, target int) (int, []techniques.Technique) {
	total := 0
	picked := []techniques.Technique{}
	for attempts := 0; total < target && attempts < 80; attempts++ {
		tech := scoringTechniques[rng.Intn(len(scoringTechniques))]
		if tech.Points > target-total {
			continue
		}
		total += tech.Points
		picked = append(picked, tech)
	}
	if total == 0 {
		tech := scoringTechniques[0]
		total, picked = tech.Points, []techniques.Technique{tech}
	}
	return total, picked
}

func summariseScore(picked []techniques.Technique) string {
	counts := map[string]int{}
	for _, tech := range picked {
		counts[tech.PointKey]++
	}
	type labelCount struct{ label string; count int }
	top := []labelCount{}
	for _, entry := range scoreLabels {
		if counts[entry.key] > 0 {
			top = append(top, labelCount{entry.label, counts[entry.key]})
		}
	}
	sort.SliceStable(top, func(i, j int) bool { return top[i].count > top[j].count })
	if len(top) == 0 {
		return "decision"
	}
	if len(top) > 2 {
		top = top[:2]
	}
	labels := make([]string, 0, len(top))
	for _, entry := range top {
		labels = append(labels, entry.label)
	}
	return strings.Join(labels, " + ")
}

func techniqueNames(picked []techniques.Technique) []string {
	names := make([]string, 0, len(picked))
	for _, tech := range picked {
		names = append(names, tech.English)
	}
	return names
}

// scoreFight builds a full JJJ score from catalogue point values. Decisions
// keep the winner ahead; submissions are independent because a tap ends the
// contest no matter what the point tally says.
func scoreFight(rng *rand.Rand) (red, blue int, method string, used []string, submission bool) {
	roll := rng.Intn(100)
	if roll < 6 {
		points := 10 + rng.Intn(21)
		return points, points, "draw", []string{"draw"}, false
	}
	if roll < 10 {
		return 0, 0, "forfeit", []string{"forfeit"}, false
	}

	winnerTarget := 10 + rng.Intn(31) // 10–40, typical of a full fight
	loserTarget := 4 + rng.Intn(24)
	if loserTarget >= winnerTarget {
		loserTarget = winnerTarget - 2
	}
	redPoints, redPicked := scoreSide(rng, winnerTarget)
	bluePoints, bluePicked := scoreSide(rng, loserTarget)
	all := append(append([]techniques.Technique{}, redPicked...), bluePicked...)
	method = summariseScore(all)

	if roll < 30 {
		finish := finishingTechniques[rng.Intn(len(finishingTechniques))]
		// A losing fighter can be ahead on points; the submission still wins.
		redPoints = 10 + rng.Intn(31)
		bluePoints = 10 + rng.Intn(31)
		method = finish.English
		used = append(techniqueNames(all), finish.English)
		return redPoints, bluePoints, method, used, true
	}
	return redPoints, bluePoints, method, techniqueNames(all), false
}

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
		redPoints, bluePoints, method, used, submission := scoreFight(rng)
		outcome := "win_red"
		if submission {
			outcome = "submission_red"
		}
		if method == "draw" || method == "forfeit" {
			outcome = method + "_red"
			if method == "draw" {
				outcome = "draw"
			}
		}
		if outcome != "draw" && rng.Intn(2) == 0 {
			outcome = replaceCorner(outcome)
			redPoints, bluePoints = bluePoints, redPoints
		}
		result.Bouts = append(result.Bouts, Bout{
			ID: fmt.Sprintf("b%03d", sequence), Stage: stage, Round: round,
			RedID: a, BlueID: b, Outcome: outcome, Method: method, Techniques: used,
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
		result.Medals = append(result.Medals, Medal{FighterID: advance[0], Kind: "gold"}, Medal{FighterID: advance[0], Kind: "champion"})
	}
	if finalLoser != "" {
		result.Medals = append(result.Medals, Medal{FighterID: finalLoser, Kind: "silver"})
	}
	if len(semiLosers) == 2 {
		// No-rematch rule: if they already met, award two bronzes instead.
		plan := rules.ThirdPlacePlan(met(result.Bouts, semiLosers[0], semiLosers[1]), true)
		if plan == "rematch" {
			addBout("bronze", "Bronze decider", semiLosers[0], semiLosers[1])
			last := result.Bouts[len(result.Bouts)-1]
			if last.Outcome == "draw" {
				result.Medals = append(result.Medals, Medal{FighterID: semiLosers[0], Kind: "bronze"}, Medal{FighterID: semiLosers[1], Kind: "bronze"})
			} else if strings.Contains(last.Outcome, "red") {
				result.Medals = append(result.Medals, Medal{FighterID: semiLosers[0], Kind: "bronze"})
			} else {
				result.Medals = append(result.Medals, Medal{FighterID: semiLosers[1], Kind: "bronze"})
			}
		} else {
			result.Medals = append(result.Medals, Medal{FighterID: semiLosers[0], Kind: "bronze"}, Medal{FighterID: semiLosers[1], Kind: "bronze"})
		}
	}

	for i := range result.Medals {
		result.Medals[i].Category = "Open"
		result.Medals[i].Discipline = "JJ Fighting System"
		result.Medals[i].EventType = "fight"
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
	switch {
	case strings.HasPrefix(value, "submission_red"):
		return "submission_blue"
	case strings.HasPrefix(value, "submission_blue"):
		return "submission_red"
	case strings.HasPrefix(value, "forfeit_red"):
		return "forfeit_blue"
	case strings.HasPrefix(value, "forfeit_blue"):
		return "forfeit_red"
	case strings.Contains(value, "red"):
		return "win_blue"
	default:
		return "win_red"
	}
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

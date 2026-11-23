package seed

import (
	"context"
	"fmt"

	"github.com/restful-otaku/j3/internal/awards"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/rules"
)

// EnsureFull adds a complete, internally consistent tournament on top of the
// base seed: pools, a six-round bracket with promotion history, scored fights,
// kata performances, medals and a full audit trail.
func EnsureFull(store *db.DB) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []struct {
		query string
		args  []any
	}{}

	fighters := make([]string, 16)
	for i := range fighters {
		fighters[i] = fmt.Sprintf("f%02d", i+1)
	}

	type match struct {
		stage, id, a, b, outcome string
		winner                   int // 1 = a, 2 = b, 0 = undecided
	}
	matches := []match{}
	rounds := []struct {
		name, key string
		fighters  []string
	}{
		{"Qualifying", "qual", fighters},
	}
	// Derive winners round by round so promotion is consistent.
	current := fighters
	roundNames := []string{"Quarter-final", "Semi-final", "Final"}
	for roundIndex, name := range roundNames {
		next := []string{}
		for i := 0; i+1 < len(current); i += 2 {
			next = append(next, current[i])
		}
		rounds = append(rounds, struct {
			name, key string
			fighters  []string
		}{name, fmt.Sprintf("r%d", roundIndex+1), next})
		current = next
	}

	matchNo := 0
	var semiLosers []string
	for _, round := range rounds {
		for i := 0; i+1 < len(round.fighters); i += 2 {
			matchNo++
			id := fmt.Sprintf("full-fight-%02d", matchNo)
			a, b := round.fighters[i], round.fighters[i+1]
			outcome, winner := "", 0
			if matchNo <= 14 {
				winner = 1
				if i%2 == 1 && matchNo > 8 {
					winner = 2
				}
				outcome = "win_red"
				if winner == 2 {
					outcome = "win_blue"
				}
			}
			matches = append(matches, match{round.key, id, a, b, outcome, winner})
			if round.key == "r2" && len(semiLosers) < 2 {
				if winner == 1 {
					semiLosers = append(semiLosers, b)
				} else {
					semiLosers = append(semiLosers, a)
				}
			}
		}
	}
	// Semi-final is round index 3 (Qualifying, R16, QF, SF). Its losers feed bronze.
	semiLosers = nil
	for _, candidate := range matches {
		if candidate.stage == "r2" { // quarter-finals: their losers become the two bronze contenders
			if candidate.winner == 1 {
				semiLosers = append(semiLosers, candidate.b)
			} else {
				semiLosers = append(semiLosers, candidate.a)
			}
		}
	}
	if len(semiLosers) >= 2 {
		matchNo++
		matches = append(matches, match{"bronze", fmt.Sprintf("full-fight-%02d", matchNo),
			semiLosers[0], semiLosers[1], "draw", 0})
	}

	for _, m := range matches {
		state := "complete"
		if m.outcome == "" {
			state = "live"
		}
		if m.stage == "bronze" {
			state = "pending"
		}
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO fights
			(id, event_id, mat_id, category_id, round_label, red_fighter_id, blue_fighter_id, state, outcome, best_of)
			VALUES (?, 'ev-demo', 'mat-1', 'cat-adult-76', ?, ?, ?, ?, ?, 3)`,
			m.id, m.stage, m.a, m.b, state, m.outcome))
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO fight_events
			(fight_id, seq, kind, actor_user_id, payload_json)
			VALUES (?, 1, 'score', 'u-official', ?)`,
			m.id, `{"corner":"red","points":2,"technique":"Takedown"}`))
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO fight_events
			(fight_id, seq, kind, actor_user_id, payload_json)
			VALUES (?, 2, 'score', 'u-official', ?)`,
			m.id, `{"corner":"blue","points":1,"technique":"Sweep"}`))
	}

	// Bracket stages and slots, with per-round history via source_slot.
	stageNames := map[string]string{
		"qual": "Qualifying", "r1": "Quarter-final", "r2": "Semi-final",
		"r3": "Final", "bronze": "Bronze decider",
	}
	order := []string{"qual", "r1", "r2", "r3", "bronze"}
	slotByFighter := map[string]string{}
	for orderIndex, key := range order {
		stageID := "full-stage-" + key
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO bracket_stages
			(id, event_id, category_id, name, kind, sort_order) VALUES (?, 'ev-demo', 'cat-adult-76', ?, ?, ?)`,
			stageID, stageNames[key], "knockout", orderIndex))
		position := 0
		for _, m := range matches {
			if m.stage != key {
				continue
			}
			for corner, fighterID := range []string{m.a, m.b} {
				state, icon, source := "pending", "", ""
				if m.winner != 0 {
					isWinner := (corner == 0 && m.winner == 1) || (corner == 1 && m.winner == 2)
					if isWinner {
						state, icon = "advanced", "WP"
						slotByFighter[fighterID] = fmt.Sprintf("%s-slot-%d", stageID, position)
					} else {
						state, icon = "eliminated", "LP"
					}
				}
				slotID := fmt.Sprintf("%s-slot-%d", stageID, position)
				if previous, ok := slotByFighter[fighterID]; ok && key != "qual" {
					source = previous
				}
				stmts = append(stmts, stmt(`INSERT OR IGNORE INTO bracket_slots
					(id, stage_id, position, fighter_id, state, result_icon, source_slot)
					VALUES (?, ?, ?, ?, ?, ?, ?)`,
					slotID, stageID, position, fighterID, state, icon, source))
				position++
			}
		}
	}

	// Pools from the same entrants.
	entries := make([]rules.Entrant, 48)
	for i := range entries {
		entries[i] = rules.Entrant{ID: fmt.Sprintf("f%02d", i+1), ClubID: fmt.Sprintf("club-%d", i%4+1), Seed: i + 1}
	}
	pools := rules.GeneratePools(entries, rules.PoolConstraints{
		TargetPoolSize: 6, MaxPools: 8, SameClubSeparation: true, Seed: 42,
	})
	for i, pool := range pools {
		poolID := fmt.Sprintf("full-pool-%02d", i+1)
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO pools
			(id, event_id, category_id, name, sort_order) VALUES (?, 'ev-demo', 'cat-adult-76', ?, ?)`,
			poolID, pool.Name, i))
		for position, fighterID := range pool.Members {
			stmts = append(stmts, stmt(`INSERT OR IGNORE INTO pool_members
				(pool_id, fighter_id, seed) VALUES (?, ?, ?)`, poolID, fighterID, position+1))
		}
	}

	// Kata performances with complete panel scores.
	for i := 0; i < 6; i++ {
		performanceID := fmt.Sprintf("full-kata-%02d", i+1)
		stmts = append(stmts, stmt(`INSERT OR IGNORE INTO kata_performances
			(id, event_id, fighter_id, kata_name, panel_size, round_number, kata_type, status, total, bunkai_score, category_id, discipline_id)
			VALUES (?, 'ev-demo', ?, ?, 5, 1, 'shitei', 'complete', ?, 0, 'cat-kata', 'disc-kata')`,
			performanceID, fmt.Sprintf("f%02d", i+41),
			[]string{"Hekigan-ken", "Seiyunchin", "Kanku-dai", "Gekusai-dai", "Pinan sono ichi", "Koryu-no-kata"}[i],
			7.4+float64(i)*0.3))
		for judge := 1; judge <= 5; judge++ {
			score := 7.0 + float64((i+judge)%4)*0.5
			stmts = append(stmts, stmt(`INSERT OR IGNORE INTO kata_scores
				(performance_id, judge, score) VALUES (?, ?, ?)`, performanceID, judge, score))
		}
	}

	// Medals are derived below from every category/discipline/event-type podium.

	for _, action := range []string{
		"full.seed.teams", "full.seed.matches", "full.seed.bracket", "full.seed.kata", "full.seed.medals",
	} {
		stmts = append(stmts, stmt(`INSERT INTO audit_log
			(tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
			VALUES ('demo-tenant', 'u-runner', 'tournament_runner', ?, 'event', 'ev-demo', '{"seeded":true}')`,
			action))
	}

	for _, statement := range stmts {
		if _, err := tx.Exec(statement.query, statement.args...); err != nil {
			return fmt.Errorf("full seed %q: %w", statement.query, err)
		}
	}
	if _, err := awards.ReconcileTx(context.Background(), tx, "ev-demo"); err != nil {
		return fmt.Errorf("full seed medals: %w", err)
	}
	return tx.Commit()
}

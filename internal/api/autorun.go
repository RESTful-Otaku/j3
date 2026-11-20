package api

import (
	"fmt"
	"net/http"

	"github.com/restful-otaku/j3/internal/rules"
)

type readinessItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Entrants  int    `json:"entrants"`
	Ready     bool   `json:"ready"`
	PoolCount int    `json:"poolCount"`
}

// readiness reports which divisions can generate a valid bracket.
func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT c.id, c.name, COALESCE(e.n,0)
		FROM categories c
		LEFT JOIN (SELECT category_id, COUNT(*) n FROM entrants WHERE event_id='ev-demo' GROUP BY category_id) e
		  ON e.category_id = c.id
		WHERE c.event_id='ev-demo' ORDER BY c.sort_order`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	items := []readinessItem{}
	for rows.Next() {
		var item readinessItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Entrants); err != nil {
			writeError(w, err)
			return
		}
		item.Ready = item.Entrants >= 2
		item.PoolCount = (item.Entrants + 3) / 4
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": items})
}

// autoRun performs the heavy lifting: pools, brackets and official assignments
// for every category that has entrants, from the organiser's configuration.
func (s *Server) autoRun(w http.ResponseWriter, r *http.Request) {
	const eventID = "ev-demo"
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT c.id, COALESCE(e.n, 0)
		FROM categories c
		LEFT JOIN (SELECT category_id, COUNT(*) n FROM entrants WHERE event_id = ? GROUP BY category_id) e
		  ON e.category_id = c.id
		WHERE c.event_id = ? ORDER BY c.sort_order`, eventID, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	type categoryCount struct {
		id    string
		count int
	}
	categories := []categoryCount{}
	for rows.Next() {
		var item categoryCount
		if err := rows.Scan(&item.id, &item.count); err != nil {
			rows.Close()
			writeError(w, err)
			return
		}
		categories = append(categories, item)
	}
	rows.Close()

	poolsCreated, stagesCreated, assignments := 0, 0, 0
	for _, category := range categories {
		if category.count < 2 {
			continue
		}
		entrantRows, err := s.Store.QueryContext(r.Context(), `
			SELECT e.fighter_id, COALESCE(e.club_id,''), COALESCE(e.seed,0)
			FROM entrants e WHERE e.event_id = ? AND e.category_id = ?
			  AND (e.status='active' OR e.status IS NULL)`, eventID, category.id)
		if err != nil {
			writeError(w, err)
			return
		}
		input := []rules.Entrant{}
		for entrantRows.Next() {
			var entrant rules.Entrant
			if err := entrantRows.Scan(&entrant.ID, &entrant.ClubID, &entrant.Seed); err != nil {
				entrantRows.Close()
				writeError(w, err)
				return
			}
			input = append(input, entrant)
		}
		entrantRows.Close()
		if len(input) < 2 {
			continue
		}
		input = rules.SeedEntrants(input, true)
		pools := rules.GeneratePools(input, rules.PoolConstraints{
			TargetPoolSize: 4, MaxPools: 12, SameClubSeparation: true, Seed: 42,
		})

		tx, err := s.Store.BeginTx(r.Context(), nil)
		if err != nil {
			writeError(w, err)
			return
		}
		_, _ = tx.ExecContext(r.Context(), `DELETE FROM pool_members WHERE pool_id IN (SELECT id FROM pools WHERE event_id=? AND category_id=?)`, eventID, category.id)
		_, _ = tx.ExecContext(r.Context(), `DELETE FROM pools WHERE event_id=? AND category_id=?`, eventID, category.id)
		for i, pool := range pools {
			poolID := newID("pool")
			_, _ = tx.ExecContext(r.Context(), `INSERT INTO pools (id, event_id, category_id, name, sort_order) VALUES (?,?,?,?,?)`,
				poolID, eventID, category.id, pool.Name, i)
			for _, fighterID := range pool.Members {
				_, _ = tx.ExecContext(r.Context(), `INSERT OR IGNORE INTO pool_members (pool_id, fighter_id) VALUES (?,?)`, poolID, fighterID)
			}
			poolsCreated++
		}
		if err := tx.Commit(); err != nil {
			writeError(w, err)
			return
		}

		advance := []string{}
		for _, pool := range pools {
			if len(pool.Members) > 0 {
				advance = append(advance, pool.Members[0])
			}
			if len(pool.Members) > 1 {
				advance = append(advance, pool.Members[1])
			}
		}
		if len(advance) > 8 {
			advance = advance[:8]
		}
		stageNames := []string{"Quarter-final", "Semi-final", "Final"}
		tx, err = s.Store.BeginTx(r.Context(), nil)
		if err != nil {
			writeError(w, err)
			return
		}
		_, _ = tx.ExecContext(r.Context(), `DELETE FROM bracket_slots WHERE stage_id IN (SELECT id FROM bracket_stages WHERE event_id=? AND category_id=?)`, eventID, category.id)
		_, _ = tx.ExecContext(r.Context(), `DELETE FROM bracket_stages WHERE event_id=? AND category_id=?`, eventID, category.id)
		for order, name := range stageNames {
			if len(advance) < 2 {
				break
			}
			field := rules.WithByes(advance)
			stageID := newID("stage")
			_, _ = tx.ExecContext(r.Context(), `INSERT INTO bracket_stages
				(id, event_id, category_id, name, kind, sort_order) VALUES (?,?,?,?,'knockout',?)`,
				stageID, eventID, category.id, name, order)
			next := []string{}
			for i := 0; i+1 < len(field); i += 2 {
				for position, fighterID := range []string{field[i], field[i+1]} {
					if fighterID == "BYE" {
						continue
					}
					_, _ = tx.ExecContext(r.Context(), `INSERT INTO bracket_slots
						(id, stage_id, position, fighter_id) VALUES (?,?,?,?)`,
						newID("slot"), stageID, i*2+position, fighterID)
				}
				if field[i] == "BYE" {
					next = append(next, field[i+1])
				} else {
					next = append(next, field[i])
				}
			}
			advance = next
			stagesCreated++
		}
		if err := tx.Commit(); err != nil {
			writeError(w, err)
			return
		}
	}

	officialRows, err := s.Store.QueryContext(r.Context(),
		`SELECT id FROM users WHERE tenant_id='demo-tenant' AND role='official' LIMIT 8`)
	if err != nil {
		writeError(w, err)
		return
	}
	officials := []string{}
	for officialRows.Next() {
		var id string
		if officialRows.Scan(&id) == nil {
			officials = append(officials, id)
		}
	}
	officialRows.Close()
	if len(officials) > 0 {
		fightRows, err := s.Store.QueryContext(r.Context(), `SELECT id FROM fights WHERE event_id=?`, eventID)
		if err != nil {
			writeError(w, err)
			return
		}
		ids := []string{}
		for fightRows.Next() {
			var id string
			if fightRows.Scan(&id) == nil {
				ids = append(ids, id)
			}
		}
		fightRows.Close()
		for i, fightID := range ids {
			for roleIndex, role := range []string{"referee", "judge", "scorekeeper"} {
				userID := officials[(i+roleIndex)%len(officials)]
				if _, err := s.Store.ExecContext(r.Context(), `
					INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?,?,?)`,
					fightID, userID, role); err == nil {
					assignments++
				}
			}
		}
	}
	s.audit(r, "tournament.auto_run", "event", eventID,
		fmt.Sprintf(`{"pools":%d,"stages":%d,"officials":%d}`, poolsCreated, stagesCreated, assignments))
	writeJSON(w, http.StatusOK, map[string]any{
		"pools": poolsCreated, "stages": stagesCreated, "officials": assignments,
	})
}

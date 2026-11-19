package api

import (
	"net/http"

	"github.com/restful-otaku/j3/internal/rules"
)

type generatePoolsRequest struct {
	EventID            string `json:"eventId"`
	CategoryID         string `json:"categoryId"`
	TargetPoolSize     int    `json:"targetPoolSize"`
	MaxPools           int    `json:"maxPools"`
	SameClubSeparation bool   `json:"sameClubSeparation"`
	Seed               int64  `json:"seed"`
}

type poolMember struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Club       string `json:"club"`
	ClubColour string `json:"clubColour"`
	Belt       string `json:"belt"`
	Seed       int    `json:"seed"`
}

type poolView struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Members []poolMember `json:"members"`
}

func (s *Server) generatePools(w http.ResponseWriter, r *http.Request) {
	var req generatePoolsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.EventID == "" {
		req.EventID = "ev-demo"
	}
	query := `SELECT e.fighter_id, COALESCE(e.club_id, ''), COALESCE(e.seed, 0)
		FROM entrants e WHERE e.event_id = ?`
	args := []any{req.EventID}
	if req.CategoryID != "" {
		query += ` AND e.category_id = ?`
		args = append(args, req.CategoryID)
	}
	rows, err := s.Store.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	entrants := []rules.Entrant{}
	for rows.Next() {
		var e rules.Entrant
		if err := rows.Scan(&e.ID, &e.ClubID, &e.Seed); err != nil {
			rows.Close()
			writeError(w, err)
			return
		}
		entrants = append(entrants, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}

	generated := rules.GeneratePools(entrants, rules.PoolConstraints{
		TargetPoolSize: req.TargetPoolSize, MaxPools: req.MaxPools,
		SameClubSeparation: req.SameClubSeparation, Seed: req.Seed,
	})

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(),
		`DELETE FROM pool_members WHERE pool_id IN (SELECT id FROM pools WHERE event_id = ?)`,
		req.EventID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM pools WHERE event_id = ?`, req.EventID); err != nil {
		writeError(w, err)
		return
	}
	for i, pool := range generated {
		poolID := newID("pool")
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO pools (id, event_id, category_id, name, sort_order)
			VALUES (?, ?, NULLIF(?, ''), ?, ?)`,
			poolID, req.EventID, req.CategoryID, pool.Name, i); err != nil {
			writeError(w, err)
			return
		}
		for _, member := range pool.Members {
			if _, err := tx.ExecContext(r.Context(), `
				INSERT OR IGNORE INTO pool_members (pool_id, fighter_id) VALUES (?, ?)`,
				poolID, member); err != nil {
				writeError(w, err)
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	pools, err := s.listPools(r, req.EventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pools": pools, "entrantCount": len(entrants)})
}

func (s *Server) listPoolsHandler(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		eventID = "ev-demo"
	}
	pools, err := s.listPools(r, eventID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pools": pools})
}

func (s *Server) listPools(r *http.Request, eventID string) ([]poolView, error) {
	rows, err := s.Store.QueryContext(r.Context(),
		`SELECT id, name FROM pools WHERE event_id = ? ORDER BY sort_order`, eventID)
	if err != nil {
		return nil, err
	}
	type summary struct{ id, name string }
	summaries := []summary{}
	for rows.Next() {
		var item summary
		if err := rows.Scan(&item.id, &item.name); err != nil {
			rows.Close()
			return nil, err
		}
		summaries = append(summaries, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	pools := []poolView{}
	for _, summary := range summaries {
		pool := poolView{ID: summary.id, Name: summary.name, Members: []poolMember{}}
		memberRows, err := s.Store.QueryContext(r.Context(), `
			SELECT f.id, f.display_name, COALESCE(c.name, ''),
			       COALESCE(c.primary_colour, '#64748b'), COALESCE(br.label, ''),
			       COALESCE(pm.seed, 0)
			FROM pool_members pm
			JOIN fighters f ON f.id = pm.fighter_id
			LEFT JOIN clubs c ON c.id = f.club_id
			LEFT JOIN belt_ranks br ON br.id = f.belt_rank_id
			WHERE pm.pool_id = ? ORDER BY pm.seed DESC, f.display_name`, summary.id)
		if err != nil {
			return nil, err
		}
		for memberRows.Next() {
			var m poolMember
			if err := memberRows.Scan(&m.ID, &m.Name, &m.Club, &m.ClubColour, &m.Belt, &m.Seed); err != nil {
				memberRows.Close()
				return nil, err
			}
			pool.Members = append(pool.Members, m)
		}
		memberRows.Close()
		if err := memberRows.Err(); err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

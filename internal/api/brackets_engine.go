package api

import (
	"fmt"
	"net/http"

	"github.com/restful-otaku/j3/internal/rules"
)

type bracketRequest struct {
	EventID        string `json:"eventId"`
	CategoryID     string `json:"categoryId"`
	Format         string `json:"format"`
	TargetPoolSize int    `json:"targetPoolSize"`
	MaxPools       int    `json:"maxPools"`
	Seed           int64  `json:"seed"`
	ThirdPlace     bool   `json:"thirdPlace"`
}

type bracketSlotView struct {
	ID          string `json:"id"`
	Position    int    `json:"position"`
	FighterID   string `json:"fighterId"`
	FighterName string `json:"fighterName"`
	Club        string `json:"club"`
	ClubColour  string `json:"clubColour"`
	Belt        string `json:"belt"`
	State       string `json:"state"`
	ResultIcon  string `json:"resultIcon"`
	Source      string `json:"source"`
}

type bracketStageView struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`
	Kind  string            `json:"kind"`
	Slots []bracketSlotView `json:"slots"`
}

type bracketEntrant struct {
	ID, ClubID, Name string
	Seed             int
}

func (s *Server) buildBracket(w http.ResponseWriter, r *http.Request) {
	var req bracketRequest
	if err := decodeJSON(r, &req); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if req.EventID == "" {
		req.EventID = "ev-demo"
	}
	if req.TargetPoolSize < 2 {
		req.TargetPoolSize = 4
	}
	if req.MaxPools < 1 {
		req.MaxPools = 12
	}

	query := `SELECT e.fighter_id, COALESCE(e.club_id,''), COALESCE(f.display_name,e.fighter_id),
		COALESCE(e.seed,0) FROM entrants e JOIN fighters f ON f.id = e.fighter_id
		WHERE e.event_id = ? AND (e.status='active' OR e.status IS NULL)`
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
	entrants := []bracketEntrant{}
	for rows.Next() {
		var e bracketEntrant
		if err := rows.Scan(&e.ID, &e.ClubID, &e.Name, &e.Seed); err != nil {
			rows.Close()
			writeError(w, err)
			return
		}
		entrants = append(entrants, e)
	}
	rows.Close()

	input := make([]rules.Entrant, len(entrants))
	for i, e := range entrants {
		input[i] = rules.Entrant{ID: e.ID, ClubID: e.ClubID, Seed: e.Seed}
	}
	pools := rules.GeneratePools(input, rules.PoolConstraints{
		TargetPoolSize: req.TargetPoolSize, MaxPools: req.MaxPools,
		SameClubSeparation: true, Seed: req.Seed,
	})
	format := rules.SuggestFormat(req.Format, len(entrants), len(pools))

	type stageSpec struct {
		name, kind string
		slots      []string
	}
	specs := []stageSpec{}
	for _, pool := range pools {
		specs = append(specs, stageSpec{pool.Name, "pool", pool.Members})
	}
	if format != "round_robin" {
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
		for len(advance) > 1 && len(stageNames) > 0 {
			specs = append(specs, stageSpec{stageNames[0], "knockout", advance})
			stageNames = stageNames[1:]
			next := []string{}
			for i := 0; i+1 < len(advance); i += 2 {
				next = append(next, advance[i])
			}
			advance = next
		}
		if req.ThirdPlace && len(specs) >= 3 {
			specs = append(specs, stageSpec{"Bronze decider", "bronze", []string{}})
		}
	}

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(r.Context(),
		`DELETE FROM bracket_slots WHERE stage_id IN (SELECT id FROM bracket_stages WHERE event_id = ?)`, req.EventID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM bracket_stages WHERE event_id = ?`, req.EventID); err != nil {
		writeError(w, err)
		return
	}
	for i, spec := range specs {
		stageID := newID("stage")
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO bracket_stages (id, event_id, category_id, name, kind, sort_order)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, ?)`,
			stageID, req.EventID, req.CategoryID, spec.name, spec.kind, i); err != nil {
			writeError(w, err)
			return
		}
		for position, fighterID := range spec.slots {
			if _, err := tx.ExecContext(r.Context(), `
				INSERT INTO bracket_slots (id, stage_id, position, fighter_id) VALUES (?, ?, ?, ?)`,
				newID("slot"), stageID, position, fighterID); err != nil {
				writeError(w, err)
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "bracket.build", "event", req.EventID,
		fmt.Sprintf(`{"format":%q,"stages":%d}`, format, len(specs)))
	stages, err := s.listBrackets(r, req.EventID, req.CategoryID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"format": format, "stages": stages})
}

func (s *Server) listBracketsHandler(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		eventID = "ev-demo"
	}
	stages, err := s.listBrackets(r, eventID, r.URL.Query().Get("categoryId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stages": stages})
}

func (s *Server) listBrackets(r *http.Request, eventID, categoryID string) ([]bracketStageView, error) {
	query := `SELECT id, name, kind FROM bracket_stages WHERE event_id = ?`
	args := []any{eventID}
	if categoryID != "" {
		query += ` AND category_id = ?`
		args = append(args, categoryID)
	}
	query += ` ORDER BY sort_order`
	rows, err := s.Store.QueryContext(r.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	type summary struct{ id, name, kind string }
	summaries := []summary{}
	for rows.Next() {
		var item summary
		if err := rows.Scan(&item.id, &item.name, &item.kind); err != nil {
			rows.Close()
			return nil, err
		}
		summaries = append(summaries, item)
	}
	rows.Close()

	stages := []bracketStageView{}
	for _, summary := range summaries {
		stage := bracketStageView{ID: summary.id, Name: summary.name, Kind: summary.kind, Slots: []bracketSlotView{}}
		slotRows, err := s.Store.QueryContext(r.Context(), `
			SELECT bs.id, bs.position, COALESCE(bs.fighter_id,''), COALESCE(f.display_name,''),
			       COALESCE(c.name,''), COALESCE(c.primary_colour,'#64748b'),
			       COALESCE(b.label,''), bs.state, bs.result_icon, bs.source_slot
			FROM bracket_slots bs
			LEFT JOIN fighters f ON f.id = bs.fighter_id
			LEFT JOIN clubs c ON c.id = f.club_id
			LEFT JOIN belt_ranks b ON b.id = f.belt_rank_id
			WHERE bs.stage_id = ? ORDER BY bs.position`, summary.id)
		if err != nil {
			return nil, err
		}
		for slotRows.Next() {
			var slot bracketSlotView
			if err := slotRows.Scan(&slot.ID, &slot.Position, &slot.FighterID, &slot.FighterName,
				&slot.Club, &slot.ClubColour, &slot.Belt, &slot.State, &slot.ResultIcon, &slot.Source); err != nil {
				slotRows.Close()
				return nil, err
			}
			stage.Slots = append(stage.Slots, slot)
		}
		slotRows.Close()
		stages = append(stages, stage)
	}
	return stages, nil
}

type promoteRequest struct {
	ResultIcon string `json:"resultIcon"`
	Winner     bool   `json:"winner"`
}

// promoteSlot records a result in the current round (keeping the card there for
// history) and copies the winner into the next stage's first free slot.
func (s *Server) promoteSlot(w http.ResponseWriter, r *http.Request) {
	slotID := r.PathValue("id")
	var body promoteRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	var stageID, fighterID, source string
	var position int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT stage_id, position, COALESCE(fighter_id,''), source_slot FROM bracket_slots WHERE id = ?`, slotID).
		Scan(&stageID, &position, &fighterID, &source); err != nil {
		writeError(w, err)
		return
	}
	state, icon := "eliminated", "LP"
	if body.Winner {
		state, icon = "advanced", "WP"
	}
	if body.ResultIcon != "" {
		icon = body.ResultIcon
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		UPDATE bracket_slots SET state = ?, result_icon = ? WHERE id = ?`, state, icon, slotID); err != nil {
		writeError(w, err)
		return
	}
	if body.Winner && fighterID != "" {
		var nextSlot string
		err := s.Store.QueryRowContext(r.Context(), `
			SELECT bs.id FROM bracket_slots bs
			JOIN bracket_stages st ON st.id = bs.stage_id
			WHERE st.event_id = (SELECT event_id FROM bracket_stages WHERE id = ?)
			  AND st.sort_order > (SELECT sort_order FROM bracket_stages WHERE id = ?)
			  AND bs.fighter_id IS NULL
			ORDER BY st.sort_order, bs.position LIMIT 1`, stageID, stageID).Scan(&nextSlot)
		if err == nil {
			_, _ = s.Store.ExecContext(r.Context(), `
				UPDATE bracket_slots SET fighter_id = ?, source_slot = ? WHERE id = ?`, fighterID, slotID, nextSlot)
		}
	}
	s.audit(r, "bracket.promote", "slot", slotID, fmt.Sprintf(`{"result":%q}`, icon))
	stages, err := s.listBrackets(r, "ev-demo", "")
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"stages": stages})
}

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func insertPeopleTx(ctx context.Context, tx *sql.Tx, req createTournamentRequest) error {
	for _, club := range req.Clubs {
		colour := club.Colour
		if colour == "" {
			colour = "#2563eb"
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO clubs (id, tenant_id, name, primary_colour)
			VALUES (?, 'demo-tenant', ?, ?)`, newID("club"), club.Name, colour); err != nil {
			return err
		}
	}
	for _, fighter := range req.Fighters {
		gender := fighter.Gender
		if gender == "" {
			gender = "open"
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO fighters
			 (id, tenant_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender)
			 VALUES (?, 'demo-tenant', NULL, NULLIF(?, ''), ?, ?, ?, ?)`,
			newID("f"), fighter.BeltRankID, fighter.Name, fighter.DOB, fighter.WeightKg, gender); err != nil {
			return err
		}
	}
	return nil
}

// importPeople loads only the clubs and fighters from a saved template.
func (s *Server) importPeople(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "participants", "create") {
		return
	}
	id := r.PathValue("id")
	var config string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT config_json FROM tournament_templates WHERE id = ?`, id).Scan(&config); err != nil {
		writeError(w, err)
		return
	}
	var req createTournamentRequest
	if err := json.Unmarshal([]byte(config), &req); err != nil {
		writeError(w, err)
		return
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	if err := insertPeopleTx(r.Context(), tx, req); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "template.import_people", "template", id,
		`{"clubs":`+strconv.Itoa(len(req.Clubs))+`,"fighters":`+strconv.Itoa(len(req.Fighters))+`}`)
	writeJSON(w, http.StatusOK, map[string]any{
		"clubs": len(req.Clubs), "fighters": len(req.Fighters),
	})
}

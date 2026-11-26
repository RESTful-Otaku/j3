package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/restful-otaku/j3/internal/awards"
)

type medalView struct {
	ID           string `json:"id"`
	FighterID    string `json:"fighterId"`
	Name         string `json:"name"`
	Club         string `json:"club"`
	ClubColour   string `json:"clubColour"`
	AvatarPath   string `json:"avatarPath"`
	Kind         string `json:"kind"`
	CategoryID   string `json:"categoryId"`
	Category     string `json:"category"`
	DisciplineID string `json:"disciplineId"`
	Discipline   string `json:"discipline"`
	EventType    string `json:"eventType"`
	AwardKey     string `json:"awardKey"`
	Source       string `json:"source"`
}

func (s *Server) awardMedals(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "edit") {
		return
	}
	var body struct {
		Gold         string   `json:"gold"`
		Silver       string   `json:"silver"`
		Bronze       []string `json:"bronze"`
		Category     string   `json:"category"`
		Discipline   string   `json:"discipline"`
		EventType    string   `json:"eventType"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	awardsToStore := []struct{ id, kind string }{}
	if body.Gold != "" {
		awardsToStore = append(awardsToStore, struct{ id, kind string }{body.Gold, "gold"})
	}
	if body.Silver != "" {
		awardsToStore = append(awardsToStore, struct{ id, kind string }{body.Silver, "silver"})
	}
	for _, id := range body.Bronze {
		if id != "" {
			awardsToStore = append(awardsToStore, struct{ id, kind string }{id, "bronze"})
		}
	}
	if len(awardsToStore) == 0 {
		writeBadRequest(w, "at least one medal is required")
		return
	}
	eventID := s.defaultEventID(r)
	if body.EventType == "" {
		body.EventType = "fight"
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	for _, award := range awardsToStore {
		key := fmt.Sprintf("manual|%s|%s|%s|%s|%s", eventID, body.Category, body.Discipline, award.id, award.kind)
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO medals
				(id, event_id, category_id, fighter_id, kind, discipline_id, event_type, award_key, source)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?, ?, 'manual')
			ON CONFLICT(event_id, award_key) WHERE award_key IS NOT NULL AND award_key <> ''
			DO UPDATE SET fighter_id = excluded.fighter_id, source = 'manual'`,
			newID("medal"), eventID, body.Category, award.id, award.kind,
			body.Discipline, body.EventType, key); err != nil {
			writeError(w, err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	encoded, _ := json.Marshal(body)
	s.audit(r, "medals.award", "event", eventID, string(encoded))
	s.broadcastState("medals")
	s.listMedalsHandler(w, r)
}

// recomputeMedals fills every category/discipline/event-type podium from the
// completed results. It is idempotent and never overwrites manual overrides.
func (s *Server) recomputeMedals(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "edit") {
		return
	}
	count, err := s.recomputeAwards(r.Context(), s.defaultEventID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "medals.recompute", "event", s.defaultEventID(r), fmt.Sprintf(`{"slots":%d}`, count))
	s.broadcastState("medals")
	s.broadcastState("standings")
	s.listMedalsHandler(w, r)
}

func (s *Server) recomputeAwards(ctx context.Context, eventID string) (int, error) {
	count, err := awards.Reconcile(ctx, s.Store.DB, eventID)
	if err != nil {
		return 0, err
	}
	s.broadcastState("medals")
	s.broadcastState("standings")
	return count, nil
}

func (s *Server) listMedalsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT m.id, m.fighter_id, COALESCE(f.display_name, m.fighter_id),
		       COALESCE(club.name,''), COALESCE(club.primary_colour,'#64748b'), COALESCE(f.avatar_path,''), m.kind,
		       COALESCE(m.category_id,''), COALESCE(cat.name,''),
		       COALESCE(m.discipline_id,''), COALESCE(dr.name, disc.name, ''),
		       m.event_type, COALESCE(m.award_key,''), m.source
		FROM medals m
		LEFT JOIN fighters f ON f.id = m.fighter_id
		LEFT JOIN clubs club ON club.id = f.club_id
		LEFT JOIN categories cat ON cat.id = m.category_id AND cat.event_id = m.event_id
		LEFT JOIN discipline_rules dr ON dr.id = m.discipline_id AND dr.event_id = m.event_id
		LEFT JOIN disciplines disc ON disc.id = m.discipline_id AND disc.event_id = m.event_id
		WHERE m.event_id = ?`
	args := []any{s.eventIDFromRequest(r)}
	if categoryID := r.URL.Query().Get("categoryId"); categoryID != "" {
		query += ` AND m.category_id = ?`
		args = append(args, categoryID)
	}
	if disciplineID := r.URL.Query().Get("disciplineId"); disciplineID != "" {
		query += ` AND m.discipline_id = ?`
		args = append(args, disciplineID)
	}
	if eventType := r.URL.Query().Get("eventType"); eventType != "" {
		query += ` AND m.event_type = ?`
		args = append(args, eventType)
	}
	query += ` ORDER BY COALESCE(dr.name, disc.name, ''), cat.sort_order, cat.name,
		CASE m.kind WHEN 'gold' THEN 0 WHEN 'silver' THEN 1 WHEN 'bronze' THEN 2 ELSE 3 END, m.id`

	rows, err := s.Store.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	medals := []medalView{}
	for rows.Next() {
		var medal medalView
		if err := rows.Scan(&medal.ID, &medal.FighterID, &medal.Name, &medal.Club, &medal.ClubColour,
			&medal.AvatarPath,
			&medal.Kind, &medal.CategoryID, &medal.Category, &medal.DisciplineID,
			&medal.Discipline, &medal.EventType, &medal.AwardKey, &medal.Source); err != nil {
			writeError(w, err)
			return
		}
		medals = append(medals, medal)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"medals": medals})
}

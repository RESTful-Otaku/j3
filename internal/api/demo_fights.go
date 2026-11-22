package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/restful-otaku/j3/internal/auth"
)

type demoFighter struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Club       string  `json:"club"`
	ClubColour string  `json:"clubColour"`
	LogoPath   string  `json:"logoPath"`
	Belt       string  `json:"belt"`
	BeltColour string  `json:"beltColour"`
	HeightCm   float64 `json:"heightCm"`
	WeightKg   float64 `json:"weightKg"`
}

type demoFightEvent struct {
	Seq     int    `json:"seq"`
	Kind    string `json:"kind"`
	Payload any    `json:"payload"`
}

type demoFight struct {
	ID       string           `json:"id"`
	Category string           `json:"category"`
	Mat      string           `json:"mat"`
	Round    string           `json:"round"`
	State    string           `json:"state"`
	Outcome  string           `json:"outcome"`
	Red      demoFighter      `json:"red"`
	Blue     demoFighter      `json:"blue"`
	Events   []demoFightEvent `json:"events"`
}

// allDemoFights returns fights with fighter, club and scoring-event detail so
// the QA scoreboard can exercise the same data path the real displays use.
func (s *Server) allDemoFights(r *http.Request) ([]demoFight, error) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT f.id, COALESCE(c.name, ''), COALESCE(m.name, ''), f.round_label,
		       f.state, f.outcome,
		       COALESCE(fr.id, ''), COALESCE(fr.display_name, ''),
		       COALESCE(cr.name, ''), COALESCE(cr.primary_colour, '#64748b'),
		       COALESCE(cr.logo_path, ''), COALESCE(br.label, ''), COALESCE(br.colour, '#94a3b8'),
		       COALESCE(ur.height_cm, 0), COALESCE(fr.weight_kg, 0),
		       COALESCE(fb.id, ''), COALESCE(fb.display_name, ''),
		       COALESCE(cb.name, ''), COALESCE(cb.primary_colour, '#64748b'),
		       COALESCE(cb.logo_path, ''), COALESCE(bb.label, ''), COALESCE(bb.colour, '#94a3b8'),
		       COALESCE(ub.height_cm, 0), COALESCE(fb.weight_kg, 0)
		FROM fights f
		LEFT JOIN categories c  ON c.id = f.category_id
		LEFT JOIN mats m        ON m.id = f.mat_id
		LEFT JOIN fighters fr   ON fr.id = f.red_fighter_id
		LEFT JOIN fighters fb   ON fb.id = f.blue_fighter_id
		LEFT JOIN clubs cr      ON cr.id = fr.club_id
		LEFT JOIN clubs cb      ON cb.id = fb.club_id
		LEFT JOIN belt_ranks br ON br.id = fr.belt_rank_id
		LEFT JOIN belt_ranks bb ON bb.id = fb.belt_rank_id
		LEFT JOIN users ur      ON ur.id = fr.user_id
		LEFT JOIN users ub      ON ub.id = fb.user_id
		ORDER BY f.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fights := []demoFight{}
	for rows.Next() {
		var f demoFight
		if err := rows.Scan(
			&f.ID, &f.Category, &f.Mat, &f.Round, &f.State, &f.Outcome,
			&f.Red.ID, &f.Red.Name, &f.Red.Club, &f.Red.ClubColour, &f.Red.LogoPath,
			&f.Red.Belt, &f.Red.BeltColour, &f.Red.HeightCm, &f.Red.WeightKg,
			&f.Blue.ID, &f.Blue.Name, &f.Blue.Club, &f.Blue.ClubColour, &f.Blue.LogoPath,
			&f.Blue.Belt, &f.Blue.BeltColour, &f.Blue.HeightCm, &f.Blue.WeightKg,
		); err != nil {
			return nil, err
		}
		f.Events = []demoFightEvent{}
		fights = append(fights, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range fights {
		eventRows, err := s.Store.QueryContext(r.Context(), `
			SELECT seq, kind, payload_json FROM fight_events
			WHERE fight_id = ? ORDER BY seq`, fights[i].ID)
		if err != nil {
			return nil, err
		}
		for eventRows.Next() {
			var e demoFightEvent
			var raw string
			if err := eventRows.Scan(&e.Seq, &e.Kind, &raw); err != nil {
				eventRows.Close()
				return nil, err
			}
			var payload any
			_ = json.Unmarshal([]byte(raw), &payload)
			e.Payload = payload
			fights[i].Events = append(fights[i].Events, e)
		}
		eventRows.Close()
		if err := eventRows.Err(); err != nil {
			return nil, err
		}

		// Scorekeeper events are authoritative; include them in the display feed.
		scoreRows, err := s.Store.QueryContext(r.Context(), `
			SELECT seq, kind, corner, points, technique, card
			FROM score_events WHERE fight_id = ? ORDER BY seq`, fights[i].ID)
		if err != nil {
			return nil, err
		}
		for scoreRows.Next() {
			var event demoFightEvent
			var seq, points int
			var kind, corner, technique, card string
			if err := scoreRows.Scan(&seq, &kind, &corner, &points, &technique, &card); err != nil {
				scoreRows.Close()
				return nil, err
			}
			event.Seq = seq
			event.Kind = kind
			event.Payload = map[string]any{
				"corner": corner, "points": points, "technique": technique, "card": card,
			}
			fights[i].Events = append(fights[i].Events, event)
		}
		scoreRows.Close()
		if err := scoreRows.Err(); err != nil {
			return nil, err
		}
		sort.SliceStable(fights[i].Events, func(a, b int) bool {
			return fights[i].Events[a].Seq < fights[i].Events[b].Seq
		})
	}
	return fights, nil
}

func (s *Server) demoFights(w http.ResponseWriter, r *http.Request) {
	fights, err := s.allDemoFights(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"fights": fights})
}

// myContests returns only the contests the signed-in official is assigned to.
// Runners see every contest so they can override assignments during an event.
func (s *Server) myContests(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	fights, err := s.allDemoFights(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if identity.Role == auth.RoleAdmin || identity.Role == auth.RoleTournamentRunner {
		writeJSON(w, http.StatusOK, map[string]any{"contests": fights, "role": identity.Role})
		return
	}
	assigned := map[string]string{}
	rows, err := s.Store.QueryContext(r.Context(),
		`SELECT fight_id, role FROM fight_officials WHERE user_id = ?`, identity.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, role string
		if rows.Scan(&id, &role) == nil {
			assigned[id] = role
		}
	}
	mine := []demoFight{}
	for _, fight := range fights {
		if role, ok := assigned[fight.ID]; ok {
			fight.Round = role + " · " + fight.Round
			mine = append(mine, fight)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"contests": mine, "role": identity.Role})
}

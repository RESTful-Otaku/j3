package api

import (
	"net/http"
)

type displayOfficial struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type eventDisplayInfo struct {
	Name        string   `json:"name"`
	BannerPath  string   `json:"bannerPath"`
	LogoPath    string   `json:"logoPath"`
	ThemeTokens string   `json:"themeTokens"`
	DetailsJSON string   `json:"detailsJson"`
	Mats        []string `json:"mats"`
}

type standingRow struct {
	FighterID  string `json:"fighterId"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Belt       string `json:"belt"`
	Club       string `json:"club"`
	ClubColour string `json:"clubColour"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	Points     int    `json:"points"`
}

func (s *Server) eventDisplay(w http.ResponseWriter, r *http.Request) {
	var info eventDisplayInfo
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT e.name, COALESCE(t.banner_path, ''), COALESCE(t.logo_path, ''),
		       COALESCE(t.tokens_json, '{}'), COALESCE(rs.config_json, '{}')
		FROM events e LEFT JOIN themes t ON t.id = e.theme_id
		LEFT JOIN rule_sets rs ON rs.id = e.rule_set_id
		WHERE e.id = 'ev-demo'`).
		Scan(&info.Name, &info.BannerPath, &info.LogoPath, &info.ThemeTokens, &info.DetailsJSON)
	if err != nil {
		writeError(w, err)
		return
	}
	info.Mats = []string{}
	rows, err := s.Store.QueryContext(r.Context(),
		`SELECT name FROM mats WHERE event_id = 'ev-demo' ORDER BY sort_order`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			writeError(w, err)
			return
		}
		info.Mats = append(info.Mats, name)
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) displayOfficials(w http.ResponseWriter, r *http.Request) {
	fightID := r.URL.Query().Get("fightId")
	query := `
		SELECT u.id, u.display_name, fo.role
		FROM fight_officials fo JOIN users u ON u.id = fo.user_id`
	args := []any{}
	if fightID != "" {
		query += ` WHERE fo.fight_id = ?`
		args = append(args, fightID)
	}
	query += ` ORDER BY fo.role`
	rows, err := s.Store.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	officials := []displayOfficial{}
	for rows.Next() {
		var o displayOfficial
		if err := rows.Scan(&o.ID, &o.Name, &o.Role); err != nil {
			writeError(w, err)
			return
		}
		officials = append(officials, o)
	}
	writeJSON(w, http.StatusOK, map[string]any{"officials": officials})
}

func (s *Server) standings(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT f.id, f.display_name, COALESCE(c.name, 'Open'),
		       COALESCE(b.label, ''), COALESCE(cl.name, 'Independent'),
		       COALESCE(cl.primary_colour, '#64748b'),
		       SUM(CASE WHEN (ft.red_fighter_id = f.id AND ft.outcome = 'win_red')
		                  OR (ft.blue_fighter_id = f.id AND ft.outcome = 'win_blue') THEN 1 ELSE 0 END),
		       SUM(CASE WHEN (ft.red_fighter_id = f.id AND ft.outcome = 'win_blue')
		                  OR (ft.blue_fighter_id = f.id AND ft.outcome = 'win_red') THEN 1 ELSE 0 END),
		       SUM(CASE WHEN ft.outcome = 'draw' AND
		                  (ft.red_fighter_id = f.id OR ft.blue_fighter_id = f.id) THEN 1 ELSE 0 END)
		FROM fighters f
		LEFT JOIN fights ft ON ft.red_fighter_id = f.id OR ft.blue_fighter_id = f.id
		LEFT JOIN clubs cl ON cl.id = f.club_id
		LEFT JOIN belt_ranks b ON b.id = f.belt_rank_id
		LEFT JOIN entrants e ON e.fighter_id = f.id
		LEFT JOIN categories c ON c.id = e.category_id
		GROUP BY f.id, c.name, b.label, cl.name
		ORDER BY 7 DESC, 9 DESC, f.display_name`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	rowsData := []standingRow{}
	for rows.Next() {
		var row standingRow
		if err := rows.Scan(
			&row.FighterID, &row.Name, &row.Category, &row.Belt, &row.Club, &row.ClubColour,
			&row.Wins, &row.Losses, &row.Draws,
		); err != nil {
			writeError(w, err)
			return
		}
		row.Points = row.Wins*3 + row.Draws
		rowsData = append(rowsData, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"standings": rowsData})
}

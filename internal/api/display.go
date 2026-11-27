package api

import (
	"net/http"
	"sort"
	"strings"

	"github.com/restful-otaku/j3/internal/rules"
)

type displayOfficial struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	PhotoPath string `json:"photoPath"`
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
	ClubLogo   string `json:"clubLogo"`
	AvatarPath string `json:"avatarPath"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	Points     int    `json:"points"`
}

func (s *Server) eventDisplay(w http.ResponseWriter, r *http.Request) {
	eventID := s.eventIDFromRequest(r)
	var info eventDisplayInfo
	err := s.Store.QueryRowContext(r.Context(), `
		SELECT e.name, COALESCE(t.banner_path, ''), COALESCE(t.logo_path, ''),
		       COALESCE(t.tokens_json, '{}'), COALESCE(rs.config_json, '{}')
		FROM events e LEFT JOIN themes t ON t.id = e.theme_id
		LEFT JOIN rule_sets rs ON rs.id = e.rule_set_id
		WHERE e.id = ?`, eventID).
		Scan(&info.Name, &info.BannerPath, &info.LogoPath, &info.ThemeTokens, &info.DetailsJSON)
	if err != nil {
		writeError(w, err)
		return
	}
	info.Mats = []string{}
	rows, err := s.Store.QueryContext(r.Context(),
		`SELECT name FROM mats WHERE event_id = ? ORDER BY sort_order`, eventID)
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
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) displayOfficials(w http.ResponseWriter, r *http.Request) {
	fightID := r.URL.Query().Get("fightId")
	query := `
		SELECT u.id, u.display_name, fo.role, COALESCE(u.photo_path,'')
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
		if err := rows.Scan(&o.ID, &o.Name, &o.Role, &o.PhotoPath); err != nil {
			writeError(w, err)
			return
		}
		officials = append(officials, o)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"officials": officials})
}

func (s *Server) standings(w http.ResponseWriter, r *http.Request) {
	eventID := s.eventIDFromRequest(r)
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT f.id, f.display_name, COALESCE(c.name, 'Open'),
		       COALESCE(b.label, ''), COALESCE(cl.name, 'Independent'),
		       COALESCE(cl.primary_colour, '#64748b'), COALESCE(cl.logo_path, ''), COALESCE(f.avatar_path, ''),
		       COALESCE(SUM(CASE WHEN (ft.side = 'red' AND ft.outcome = 'win_red')
		                  OR (ft.side = 'blue' AND ft.outcome = 'win_blue') THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN (ft.side = 'red' AND ft.outcome = 'win_blue')
		                  OR (ft.side = 'blue' AND ft.outcome = 'win_red') THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN ft.outcome = 'draw' THEN 1 ELSE 0 END), 0)
		FROM fighters f
		LEFT JOIN (
		  SELECT red_fighter_id AS fighter_id, 'red' AS side, outcome FROM fights WHERE event_id = ?
		  UNION ALL
		  SELECT blue_fighter_id AS fighter_id, 'blue' AS side, outcome FROM fights WHERE event_id = ?
		) ft ON ft.fighter_id = f.id
		LEFT JOIN clubs cl ON cl.id = f.club_id
		LEFT JOIN belt_ranks b ON b.id = f.belt_rank_id
		LEFT JOIN entrants e ON e.fighter_id = f.id AND e.event_id = ?
		LEFT JOIN categories c ON c.id = e.category_id
		GROUP BY f.id, c.name, b.label, cl.name
		ORDER BY 8 DESC, 10 DESC, f.display_name`, eventID, eventID, eventID)
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
			&row.ClubLogo, &row.AvatarPath, &row.Wins, &row.Losses, &row.Draws,
		); err != nil {
			writeError(w, err)
			return
		}
		row.Points = row.Wins*3 + row.Draws
		rowsData = append(rowsData, row)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	chain := []string{"wins", "draws", "points_scored", "points_conceded"}
	if configured := r.URL.Query().Get("chain"); configured != "" {
		parts := strings.Split(configured, ",")
		allowed := map[string]bool{"wins": true, "losses": true, "draws": true,
			"submissions": true, "points_scored": true, "points_conceded": true}
		filtered := []string{}
		for _, part := range parts {
			if allowed[strings.TrimSpace(part)] {
				filtered = append(filtered, strings.TrimSpace(part))
			}
		}
		if len(filtered) > 0 {
			chain = filtered
		}
	}
	sort.SliceStable(rowsData, func(i, j int) bool {
		a := rules.TieBreakStats{
			Wins: rowsData[i].Wins, Draws: rowsData[i].Draws,
			PointsScored: rowsData[i].Points, PointsConceded: rowsData[i].Losses,
		}
		b := rules.TieBreakStats{
			Wins: rowsData[j].Wins, Draws: rowsData[j].Draws,
			PointsScored: rowsData[j].Points, PointsConceded: rowsData[j].Losses,
		}
		return rules.TieBreak(a, b, chain) < 0
	})
	writeJSON(w, http.StatusOK, map[string]any{"standings": rowsData})
}

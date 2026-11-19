package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type createDiscipline struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Style        string   `json:"style"`
	Scoring      string   `json:"scoring"`
	MatchSeconds int      `json:"matchSeconds"`
	Rounds       int      `json:"rounds"`
	Requirements []string `json:"requirements"`
}

type createCategory struct {
	Name       string  `json:"name"`
	Kind       string  `json:"kind"`
	Gender     string  `json:"gender"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	BeltRankID string  `json:"beltRankId"`
}

type createConstraints struct {
	Cap           int    `json:"cap"`
	MaxPools      int    `json:"maxPools"`
	AutoPlace     bool   `json:"autoPlace"`
	SeparateClubs bool   `json:"separateClubs"`
	GenderPolicy  string `json:"genderPolicy"`
}

type createInvite struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	ClubName string `json:"clubName"`
}

type createTournamentRequest struct {
	Name          string              `json:"name"`
	Title         string              `json:"title"`
	Venue         string              `json:"venue"`
	StartDate     string              `json:"startDate"`
	EndDate       string              `json:"endDate"`
	Theme         string              `json:"theme"`
	LogoPath      string              `json:"logoPath"`
	BannerPath    string              `json:"bannerPath"`
	TemplateName  string              `json:"templateName"`
	SaveTemplate  bool                `json:"saveTemplate"`
	Disciplines   []createDiscipline  `json:"disciplines"`
	Categories    []createCategory    `json:"categories"`
	Constraints   createConstraints   `json:"constraints"`
	Invites       []createInvite      `json:"invites"`
	Protection    []string            `json:"protection"`
	CornerColours map[string]string   `json:"cornerColours"`
	Schedule      []map[string]string `json:"schedule"`
}

func themeTokens(preset string) string {
	switch strings.ToLower(preset) {
	case "street":
		return `{"primary":"#ef4444","accent":"#fbbf24","ink":"#f8fafc","style":"street"}`
	case "ocean":
		return `{"primary":"#0ea5e9","accent":"#22d3ee","ink":"#f8fafc","style":"ocean"}`
	default:
		return `{"primary":"#2563eb","accent":"#f59e0b","ink":"#0f172a","style":"classic"}`
	}
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

// createTournament turns a validated wizard payload into a real event skeleton.
func (s *Server) createTournament(w http.ResponseWriter, r *http.Request) {
	var req createTournamentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeBadRequest(w, "event name is required")
		return
	}
	if len(req.Disciplines) == 0 {
		writeBadRequest(w, "add at least one discipline")
		return
	}
	if len(req.Categories) == 0 {
		writeBadRequest(w, "add at least one category")
		return
	}
	for _, d := range req.Disciplines {
		if strings.TrimSpace(d.Name) == "" {
			writeBadRequest(w, "every discipline needs a name")
			return
		}
	}

	detailsJSON, _ := json.Marshal(map[string]any{
		"disciplines": req.Disciplines, "protection": req.Protection,
		"cornerColours": req.CornerColours, "schedule": req.Schedule, "venue": req.Venue,
	})
	categoryJSON, _ := json.Marshal(req.Categories)
	constraintJSON, _ := json.Marshal(req.Constraints)
	inviteJSON, _ := json.Marshal(req.Invites)
	configJSON, _ := json.Marshal(req)

	eventID := newID("ev")
	themeID := newID("theme")
	ruleSetID := newID("rs")
	formatID := newID("fmt")
	templateID := newID("tpl")
	if !req.SaveTemplate {
		templateID = ""
	}

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()

	exec := func(query string, args ...any) {
		if err != nil {
			return
		}
		_, err = tx.ExecContext(r.Context(), query, args...)
	}

	exec(`INSERT INTO themes (id, tenant_id, name, tokens_json, logo_path, banner_path)
		VALUES (?, 'demo-tenant', ?, ?, ?, ?)`,
		themeID, req.Name+" theme", themeTokens(req.Theme), req.LogoPath, req.BannerPath)
	exec(`INSERT INTO rule_sets (id, tenant_id, name, config_json) VALUES (?, 'demo-tenant', ?, ?)`,
		ruleSetID, req.Name+" rules", string(detailsJSON))
	exec(`INSERT INTO format_policies (id, tenant_id, name, config_json) VALUES (?, 'demo-tenant', ?, ?)`,
		formatID, req.Name+" format", string(constraintJSON))
	exec(`INSERT INTO events
		(id, tenant_id, theme_id, rule_set_id, format_policy_id, name, starts_at, ends_at, status)
		VALUES (?, 'demo-tenant', ?, ?, ?, ?, ?, ?, 'draft')`,
		eventID, themeID, ruleSetID, formatID, req.TitleOrName(), req.StartDate, req.EndDate)
	if err != nil {
		writeError(w, err)
		return
	}

	for i, discipline := range req.Disciplines {
		rulesJSON, _ := json.Marshal(discipline)
		exec(`INSERT INTO disciplines (id, event_id, name, rules_json, sort_order) VALUES (?, ?, ?, ?, ?)`,
			newID("disc"), eventID, discipline.Name, string(rulesJSON), i)
		if err != nil {
			writeError(w, err)
			return
		}
	}
	for i, category := range req.Categories {
		exec(`INSERT INTO categories
			(id, event_id, name, kind, min_value, max_value, belt_rank_id, gender, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newID("cat"), eventID, category.Name, category.Kind, category.Min, category.Max,
			nullIfEmpty(category.BeltRankID), category.Gender, i)
		if err != nil {
			writeError(w, err)
			return
		}
	}
	for _, invite := range req.Invites {
		var clubID any
		if invite.ClubName != "" {
			var found string
			if tx.QueryRowContext(r.Context(),
				`SELECT id FROM clubs WHERE tenant_id = 'demo-tenant' AND name = ?`, invite.ClubName).
				Scan(&found) == nil {
				clubID = found
			}
		}
		exec(`INSERT INTO invitations
			(id, tenant_id, event_id, club_id, role, email, token_hash, expires_at)
			VALUES (?, 'demo-tenant', ?, ?, ?, ?, ?, '')`,
			newID("inv"), eventID, clubID, invite.Role, invite.Email, newID("tok"))
		if err != nil {
			writeError(w, err)
			return
		}
	}
	if templateID != "" {
		exec(`INSERT INTO tournament_templates (id, tenant_id, name, config_json)
			VALUES (?, 'demo-tenant', ?, ?)`,
			templateID, req.TemplateNameOrName(), string(configJSON))
		if err != nil {
			writeError(w, err)
			return
		}
	}
	exec(`INSERT INTO audit_log
		(tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		VALUES ('demo-tenant', 'u-runner', 'tournament_runner', 'tournament.create', 'event', ?, ?)`,
		eventID, fmt.Sprintf(`{"disciplines":%d,"categories":%d,"invites":%d}`,
			len(req.Disciplines), len(req.Categories), len(req.Invites)))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"eventId": eventID, "templateId": templateID,
		"disciplineCount": len(req.Disciplines), "categoryCount": len(req.Categories),
		"inviteCount": len(req.Invites), "categoryJson": string(categoryJSON),
		"inviteJson": string(inviteJSON),
	})
}

func (req createTournamentRequest) TitleOrName() string {
	if strings.TrimSpace(req.Title) != "" {
		return req.Title
	}
	return req.Name
}

func (req createTournamentRequest) TemplateNameOrName() string {
	if strings.TrimSpace(req.TemplateName) != "" {
		return req.TemplateName
	}
	return req.Name + " template"
}

func (s *Server) templates(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, name, config_json, created_at FROM tournament_templates
		WHERE tenant_id = 'demo-tenant' ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	type template struct {
		ID, Name, ConfigJSON, CreatedAt string
	}
	list := []template{}
	for rows.Next() {
		var item template
		if err := rows.Scan(&item.ID, &item.Name, &item.ConfigJSON, &item.CreatedAt); err != nil {
			writeError(w, err)
			return
		}
		list = append(list, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": list})
}

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/restful-otaku/j3/internal/techniques"
)

// techniquePointOption is a configured scoring value surfaced to officials.
// The console shows these instead of hard-coded numbers so every federation's
// rule set can award different points for the same technique.
type techniquePointOption struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Value  int    `json:"value"`
	Source string `json:"source"`
}

type techniqueView struct {
	techniques.Technique
	ResolvedPoints int `json:"resolvedPoints"`
}

var pointKeyAliases = map[string][]string{
	"strike":     {"strike", "punch", "kick", "tsuki", "geri", "hiji", "hiza", "teisho", "nerve"},
	"kick":       {"strike", "kick", "geri"},
	"throw":      {"throw", "nage", "kaeshi", "sacrifice"},
	"takedown":   {"takedown", "taoshi", "sweep", "reap", "hook", "harai"},
	"sweep":      {"takedown", "sweep", "harai"},
	"reap":       {"takedown", "reap", "harai"},
	"hook":       {"takedown", "strike", "hook", "kakato"},
	"hold":       {"hold", "pin", "osae", "kesa", "control", "guard"},
	"ground":     {"hold", "ground", "pin", "guard"},
	"control":    {"hold", "control", "guard"},
	"break":      {"hold", "break", "escape"},
	"standing":   {"strike", "standing", "throw"},
	"submission": {"submission", "tap", "choke", "lock", "armbar", "triangle"},
	"ippon":      {"ippon", "full point", "instant win"},
	"wazaari":    {"waza", "half point", "partial"},
	"penalty":    {"penalty", "shido", "card", "foul", "hansoku", "yellow", "red"},
	"kata":       {"kata", "performance", "excellent", "very good", "acceptable"},
}

func (s *Server) listTechniques(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	options := s.techniquePointOptions(r.Context(), s.eventIDFromRequest(r))
	matched := techniques.Search(query.Get("q"), query.Get("group"), 0)
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}
	views := make([]techniqueView, 0, len(matched))
	for _, item := range matched {
		resolved := resolveTechniquePoints(item, options, "")
		item.Points = resolved
		views = append(views, techniqueView{Technique: item, ResolvedPoints: resolved})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"eventId":      s.eventIDFromRequest(r),
		"techniques":   views,
		"groups":       techniques.Groups(),
		"pointOptions": options,
		"count":        len(views),
	})
}

// techniquePointOptions reads point values from the event's rule-set snapshot
// and its per-discipline rules. A catalogue entry never hard-codes a federation
// score when the organiser has configured one.
func (s *Server) techniquePointOptions(ctx context.Context, eventID string) []techniquePointOption {
	options := []techniquePointOption{}
	seen := map[string]bool{}
	add := func(key, label string, value float64, source string) {
		key = strings.ToLower(strings.TrimSpace(key))
		label = strings.TrimSpace(label)
		if key == "" || label == "" || seen[key+"|"+label] {
			return
		}
		seen[key+"|"+label] = true
		options = append(options, techniquePointOption{Key: key, Label: label, Value: int(value), Source: source})
	}

	var config string
	if err := s.Store.QueryRowContext(ctx, `
		SELECT COALESCE(rs.config_json,'{}')
		FROM events e LEFT JOIN rule_sets rs ON rs.id = e.rule_set_id
		WHERE e.id = ?`, eventID).Scan(&config); err == nil {
		var ruleset struct {
			Disciplines []struct {
				Name   string `json:"name"`
				Points []struct {
					Label string  `json:"label"`
					Value float64 `json:"value"`
				} `json:"points"`
			} `json:"disciplines"`
		}
		if json.Unmarshal([]byte(config), &ruleset) == nil {
			for _, discipline := range ruleset.Disciplines {
				for _, point := range discipline.Points {
					add(pointLabelKey(point.Label), point.Label, point.Value, "ruleset:"+discipline.Name)
				}
			}
		}
	}

	rows, err := s.Store.QueryContext(ctx,
		`SELECT name, rules_json FROM discipline_rules WHERE event_id = ? ORDER BY id`, eventID)
	if err == nil {
		defer rows.Close()
		type disciplineScore struct {
			name, raw string
		}
		disciplines := []disciplineScore{}
		for rows.Next() {
			var name, raw string
			if rows.Scan(&name, &raw) == nil {
				disciplines = append(disciplines, disciplineScore{name, raw})
			}
		}
		for _, discipline := range disciplines {
			var rules struct {
				Scoring map[string]any     `json:"scoring"`
				Hold    *int               `json:"holdPoints"`
				Panel   map[string]float64 `json:"panelScores"`
			}
			if json.Unmarshal([]byte(discipline.raw), &rules) != nil {
				continue
			}
			keys := make([]string, 0, len(rules.Scoring)+len(rules.Panel))
			for key := range rules.Scoring {
				keys = append(keys, key)
			}
			for key := range rules.Panel {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				if value, ok := numericValue(rules.Scoring[key]); ok {
					add(key, humanPointLabel(key), value, "discipline:"+discipline.name)
					continue
				}
				if value, ok := rules.Panel[key]; ok {
					add(key, humanPointLabel(key), value, "discipline:"+discipline.name)
				}
			}
			if rules.Hold != nil {
				add("hold", "Hold down", float64(*rules.Hold), "discipline:"+discipline.name)
			}
		}
	}
	return options
}

func numericValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func pointLabelKey(label string) string {
	key := strings.ToLower(strings.TrimSpace(label))
	key = strings.NewReplacer(" ", "", "-", "", "/", "", "_", "").Replace(key)
	return key
}

func humanPointLabel(key string) string {
	label := strings.NewReplacer("_", " ", "-", " ").Replace(key)
	parts := strings.Fields(label)
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		parts[i] = strings.ToUpper(string(runes[0])) + string(runes[1:])
	}
	return strings.Join(parts, " ")
}

func optionMatchesPointKey(option techniquePointOption, pointKey string) bool {
	aliases := pointKeyAliases[pointKey]
	if len(aliases) == 0 {
		aliases = []string{pointKey}
	}
	haystack := option.Key + " " + strings.ToLower(option.Label)
	for _, alias := range aliases {
		if strings.Contains(haystack, alias) {
			return true
		}
	}
	return false
}

func resolveTechniquePoints(item techniques.Technique, options []techniquePointOption, pointLabel string) int {
	if strings.TrimSpace(pointLabel) != "" {
		wanted := pointLabelKey(pointLabel)
		for _, option := range options {
			if pointLabelKey(option.Label) == wanted || option.Key == wanted {
				return option.Value
			}
		}
	}
	for _, option := range options {
		if optionMatchesPointKey(option, item.PointKey) {
			return option.Value
		}
	}
	return item.Points
}

// resolveScoringInput makes a catalogue selection authoritative: the server
// resolves its points, kind and card from the catalogue plus the event's rules
// so a client cannot award arbitrary points by editing a request.
func (s *Server) resolveScoringInput(ctx context.Context, fightID string, body *scoreRequest) {
	if body.TechniqueID == "" {
		return
	}
	item, ok := techniques.ByID(body.TechniqueID)
	if !ok {
		return
	}
	eventID, err := s.fightEventID(ctx, fightID)
	if err != nil || eventID == "" {
		eventID = s.defaultEventIDForContext(ctx)
	}
	options := s.techniquePointOptions(ctx, eventID)
	if body.Kind == "" {
		body.Kind = item.Kind
	}
	if body.Card == "" {
		body.Card = item.Card
	}
	// The label is derived from the catalogue, never from the request, so the
	// fight ledger and every display show the same official technique name.
	body.Technique = fmt.Sprintf("%s (%s)", item.English, item.Romaji)
	// Explicit negative card values are the configured penalty ladder; keep
	// them instead of replacing them with a generic scoring option.
	if item.Kind == "penalty" && body.Points < 0 {
		return
	}
	body.Points = resolveTechniquePoints(item, options, body.PointLabel)
}

func (s *Server) fightEventID(ctx context.Context, fightID string) (string, error) {
	var eventID string
	err := s.Store.QueryRowContext(ctx, `SELECT event_id FROM fights WHERE id = ?`, fightID).Scan(&eventID)
	return eventID, err
}

// defaultEventID in handlers.go takes an HTTP request; this helper is used by
// background resolution where only a context is available.
func (s *Server) defaultEventIDForContext(ctx context.Context) string {
	var id string
	if err := s.Store.QueryRowContext(ctx, `SELECT id FROM events ORDER BY starts_at DESC, id DESC LIMIT 1`).Scan(&id); err == nil {
		return id
	}
	return "ev-okami"
}

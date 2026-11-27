package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// sendAnnouncement records a multi-channel message. Provider credentials are
// required for real delivery; demo mode records the exact payload for QA and
// broadcasts to every connected display through the live event stream.
func (s *Server) sendAnnouncement(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "fights", "create") {
		return
	}
	var body struct {
		EventID  string `json:"eventId"`
		Title    string `json:"title"`
		Body     string `json:"body"`
		Channels string `json:"channels"`
		Audience string `json:"audience"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Title == "" {
		writeBadRequest(w, "title is required")
		return
	}
	if body.EventID == "" {
		body.EventID = s.defaultEventID(r)
	}
	if body.Channels == "" {
		body.Channels = "push"
	}
	if body.Audience == "" {
		body.Audience = "all"
	}
	audiences := map[string]bool{"all": true, "fighters": true, "officials": true, "clubs": true, "guardians": true, "participants": true, "spectators": true, "admins": true}
	channels := map[string]bool{}
	for _, channel := range strings.Split(body.Channels, ",") {
		channels[strings.TrimSpace(channel)] = true
	}
	if !audiences[body.Audience] {
		writeBadRequest(w, "invalid audience")
		return
	}
	for channel := range channels {
		if channel != "push" && channel != "email" && channel != "sms" {
			writeBadRequest(w, "invalid channel")
			return
		}
	}
	if len(body.Body) > 1000 {
		writeBadRequest(w, "message is too long")
		return
	}
	id := newID("ann")
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO announcements (id, event_id, title, body, channels, audience)
		VALUES (?, ?, ?, ?, ?, ?)`,
		id, body.EventID, body.Title, body.Body, body.Channels, body.Audience)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "announcement.send", "event", body.EventID,
		fmt.Sprintf(`{"title":%q,"channels":%q,"audience":%q}`, body.Title, body.Channels, body.Audience))
	s.Hub.broadcast(liveEvent{Type: "announcement", FightID: "", Payload: body})
	delivered := s.fanoutNotifications(body.EventID, "announcement", body.Title, body.Body, body.Channels, body.Audience)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": id, "delivered": delivered, "channels": body.Channels,
	})
}

func (s *Server) listAnnouncements(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT id, title, body, channels, audience, created_at
		FROM announcements ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	type announcement struct {
		ID, Title, Body, Channels, Audience, CreatedAt string
	}
	items := []announcement{}
	for rows.Next() {
		var item announcement
		if err := rows.Scan(&item.ID, &item.Title, &item.Body, &item.Channels,
			&item.Audience, &item.CreatedAt); err != nil {
			writeError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"announcements": items})
}

// suggestions derives participant-driven setup recommendations from entrant data.
func (s *Server) suggestions(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT f.id, COALESCE(f.date_of_birth,''), COALESCE(f.weight_kg,0),
		       COALESCE(b.sort_order,0), COALESCE(f.gender,'open'), COALESCE(c.name,'')
		FROM fighters f
		LEFT JOIN belt_ranks b ON b.id = f.belt_rank_id
		LEFT JOIN entrants e ON e.fighter_id = f.id AND e.event_id = 'ev-okami'
		LEFT JOIN categories c ON c.id = e.category_id
		WHERE f.tenant_id='demo-tenant' ORDER BY f.display_name LIMIT 60`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	type athlete struct {
		ID, Category string
		Weight       float64
		BeltOrder    int
	}
	athletes := []athlete{}
	for rows.Next() {
		var item athlete
		var dob, gender string
		if err := rows.Scan(&item.ID, &dob, &item.Weight, &item.BeltOrder, &gender, &item.Category); err != nil {
			writeError(w, err)
			return
		}
		athletes = append(athletes, item)
	}
	counts := map[string]int{}
	weights := map[string]int{}
	for _, item := range athletes {
		counts[item.Category]++
		if item.Weight < 55 {
			weights["light"]++
		} else if item.Weight < 75 {
			weights["middle"]++
		} else {
			weights["heavy"]++
		}
	}
	type suggestion struct {
		Kind, Title, Reason string
	}
	suggestions := []suggestion{}
	if weights["light"] > 0 {
		suggestions = append(suggestions, suggestion{"category", "Light weight pool", fmt.Sprintf("%d athletes under 55kg", weights["light"])})
	}
	if weights["middle"] > 0 {
		suggestions = append(suggestions, suggestion{"category", "Middle weight pool", fmt.Sprintf("%d athletes 55-75kg", weights["middle"])})
	}
	if weights["heavy"] > 0 {
		suggestions = append(suggestions, suggestion{"category", "Heavy weight pool", fmt.Sprintf("%d athletes over 75kg", weights["heavy"])})
	}
	categoryNames := make([]string, 0, len(counts))
	for category := range counts {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	for _, category := range categoryNames {
		count := counts[category]
		if category == "" {
			continue
		}
		format := "round robin"
		if count > 8 {
			format = "round robin into single elimination"
		}
		suggestions = append(suggestions, suggestion{"format", category, fmt.Sprintf("%d entrants — %s recommended", count, format)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions, "athletes": len(athletes)})
}

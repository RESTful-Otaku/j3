package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

// requireRole enforces RBAC from the authenticated session identity only.
// Client-supplied role/actor headers are never trusted.
func (s *Server) requireRole(w http.ResponseWriter, r *http.Request, action auth.Action) bool {
	identity, ok := auth.FromContext(r.Context())
	if !ok || !auth.Can(identity.Role, action) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "your role cannot perform this action"})
		return false
	}
	return true
}

// requireAssigned enforces role and per-contest assignment in one place.
// The runner is the event owner and may override assignment; officials must be
// assigned to the contest they are scoring.
func (s *Server) requireAssigned(w http.ResponseWriter, r *http.Request, action auth.Action, fightID string) bool {
	identity, ok := auth.FromContext(r.Context())
	if !ok || !auth.Can(identity.Role, action) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "your role cannot perform this action"})
		return false
	}
	if identity.Role != auth.RoleTournamentRunner && !s.isAssigned(r, fightID, identity.UserID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user is not assigned to this contest"})
		return false
	}
	return true
}

type liveEvent struct {
	Type    string `json:"type"`
	FightID string `json:"fightId"`
	Payload any    `json:"payload"`
}

// LiveHub fans out authoritative state changes to every open display.
type LiveHub struct {
	mu   sync.Mutex
	subs map[chan liveEvent]bool
}

func newLiveHub() *LiveHub { return &LiveHub{subs: map[chan liveEvent]bool{}} }

func (h *LiveHub) subscribe() chan liveEvent {
	ch := make(chan liveEvent, 32)
	h.mu.Lock()
	h.subs[ch] = true
	h.mu.Unlock()
	return ch
}

func (h *LiveHub) unsubscribe(ch chan liveEvent) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *LiveHub) broadcast(event liveEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- event:
		default: // slow displays skip frames rather than blocking the venue server
		}
	}
}

type scoreEventView struct {
	Seq       int    `json:"seq"`
	Kind      string `json:"kind"`
	Corner    string `json:"corner"`
	Points    int    `json:"points"`
	Technique string `json:"technique"`
	Card      string `json:"card"`
	Actor     string `json:"actor"`
	CreatedAt string `json:"createdAt"`
}

type FightStateView struct {
	FightID     string           `json:"fightId"`
	State       string           `json:"state"`
	Outcome     string           `json:"outcome"`
	Running     bool             `json:"running"`
	Seconds     int              `json:"seconds"`
	Round       int              `json:"round"`
	RedScore    int              `json:"redScore"`
	BlueScore   int              `json:"blueScore"`
	RedPenalty  int              `json:"redPenalty"`
	BluePenalty int              `json:"bluePenalty"`
	HoldSeconds int              `json:"holdSeconds"`
	HoldCorner  string           `json:"holdCorner"`
	Events      []scoreEventView `json:"events"`
}

func (s *Server) fightState(ctx context.Context, fightID string) (FightStateView, error) {
	view := FightStateView{FightID: fightID, Events: []scoreEventView{}}
	var running int
	var endsAt string
	err := s.Store.QueryRowContext(ctx, `
		SELECT f.state, f.outcome, COALESCE(fs.running,0), COALESCE(fs.seconds_remaining,180),
		       COALESCE(fs.round,1), COALESCE(fs.hold_seconds,0), COALESCE(fs.hold_corner,''),
		       COALESCE(fs.ends_at, '')
		FROM fights f LEFT JOIN fight_state fs ON fs.fight_id = f.id
		WHERE f.id = ?`, fightID).
		Scan(&view.State, &view.Outcome, &running, &view.Seconds, &view.Round,
			&view.HoldSeconds, &view.HoldCorner, &endsAt)
	if err != nil {
		return view, err
	}
	view.Running = running == 1
	if view.Running && endsAt != "" {
		if deadline, err := time.Parse("2006-01-02T15:04:05.999Z", endsAt); err == nil {
			remaining := int(time.Until(deadline).Seconds() + 0.999)
			if remaining < 0 {
				remaining = 0
			}
			view.Seconds = remaining
			view.Running = remaining > 0
		}
	}
	rows, err := s.Store.QueryContext(ctx, `
		SELECT seq, kind, corner, points, technique, card, actor_user_id, created_at
		FROM score_events WHERE fight_id = ? ORDER BY seq`, fightID)
	if err != nil {
		return view, err
	}
	defer rows.Close()
	for rows.Next() {
		var event scoreEventView
		if err := rows.Scan(&event.Seq, &event.Kind, &event.Corner, &event.Points,
			&event.Technique, &event.Card, &event.Actor, &event.CreatedAt); err != nil {
			return view, err
		}
		view.Events = append(view.Events, event)
		if event.Kind == "score" && event.Corner == "red" {
			view.RedScore += event.Points
		}
		if event.Kind == "score" && event.Corner == "blue" {
			view.BlueScore += event.Points
		}
		if event.Kind == "penalty" && event.Corner == "red" {
			view.RedPenalty++
		}
		if event.Kind == "penalty" && event.Corner == "blue" {
			view.BluePenalty++
		}
	}
	return view, rows.Err()
}

func (s *Server) broadcastFight(r *http.Request, fightID, kind string) {
	state, err := s.fightState(r.Context(), fightID)
	if err != nil {
		return
	}
	s.Hub.broadcast(liveEvent{Type: kind, FightID: fightID, Payload: state})
}

func (s *Server) getFightState(w http.ResponseWriter, r *http.Request) {
	state, err := s.fightState(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, fmt.Errorf("streaming unsupported"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if origin := allowedCORSOrigin(r); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	ch := s.Hub.subscribe()
	defer s.Hub.unsubscribe(ch)
	fmt.Fprint(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()
	for {
		select {
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

type timerRequest struct {
	Seconds int `json:"seconds"`
	Round   int `json:"round"`
}

func (s *Server) startFight(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	var body timerRequest
	_ = decodeJSON(r, &body)
	seconds := body.Seconds
	if seconds < 10 || seconds > 3600 {
		seconds = 180
	}
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO fight_state (fight_id, seconds_remaining, running, round, ends_at)
		VALUES (?, ?, 1, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now', ?))
		ON CONFLICT(fight_id) DO UPDATE SET running = 1,
		  seconds_remaining = excluded.seconds_remaining,
		  ends_at = strftime('%Y-%m-%dT%H:%M:%fZ','now', ?),
		  round = excluded.round`,
		id, seconds, max(1, body.Round), fmt.Sprintf("+%d seconds", seconds),
		fmt.Sprintf("+%d seconds", seconds))
	if err != nil {
		writeError(w, err)
		return
	}
	_, _ = s.Store.ExecContext(r.Context(), `UPDATE fights SET state='live', outcome='' WHERE id=?`, id)
	s.audit(r, "fight.start", "fight", id, fmt.Sprintf(`{"seconds":%d}`, seconds))
	s.broadcastFight(r, id, "fight.started")
	s.getFightState(w, r)
}

func (s *Server) pauseFight(w http.ResponseWriter, r *http.Request) {
	s.setRunning(w, r, 0, "fight.pause")
}

func (s *Server) resumeFight(w http.ResponseWriter, r *http.Request) {
	s.setRunning(w, r, 1, "fight.resume")
}

func (s *Server) setRunning(w http.ResponseWriter, r *http.Request, value int, action string) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO fight_state (fight_id, running) VALUES (?, ?)
		ON CONFLICT(fight_id) DO UPDATE SET
		  running = excluded.running,
		  ends_at = CASE WHEN excluded.running = 1
		    THEN strftime('%Y-%m-%dT%H:%M:%fZ','now', '+' || fight_state.seconds_remaining || ' seconds')
		    ELSE '' END,
		  seconds_remaining = CASE WHEN excluded.running = 0 AND fight_state.ends_at <> ''
		    THEN MAX(0, CAST((julianday(fight_state.ends_at) - julianday('now')) * 86400 AS INTEGER))
		    ELSE fight_state.seconds_remaining END`, id, value)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, action, "fight", id, fmt.Sprintf(`{"running":%d}`, value))
	s.broadcastFight(r, id, action)
	s.getFightState(w, r)
}

type scoreRequest struct {
	Corner    string `json:"corner"`
	Points    int    `json:"points"`
	Technique string `json:"technique"`
	Card      string `json:"card"`
	Kind      string `json:"kind"`
}

func (s *Server) addFightEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	var body scoreRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Corner != "red" && body.Corner != "blue" {
		writeBadRequest(w, "corner must be red or blue")
		return
	}
	kind := body.Kind
	if kind == "" {
		kind = "score"
	}
	if kind != "score" && kind != "penalty" && kind != "submission" {
		writeBadRequest(w, "unknown event kind")
		return
	}
	identity, _ := auth.FromContext(r.Context())
	actor := identity.UserID
	if actor == "" {
		actor = "system"
	}
	payload, _ := json.Marshal(body)
	// A single INSERT ... SELECT derives the next sequence atomically, so two
	// scorers submitting simultaneously cannot collide on the UNIQUE(fight_id,seq).
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO score_events
		 (fight_id, seq, kind, corner, points, technique, card, actor_user_id, payload_json)
		SELECT ?, COALESCE(MAX(seq),0)+1, ?, ?, ?, ?, ?, ?, ?
		FROM score_events WHERE fight_id = ?`,
		id, kind, body.Corner, body.Points, body.Technique, body.Card, actor, string(payload), id)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "fight."+kind, "fight", id, string(payload))
	s.broadcastFight(r, id, "fight.event")
	s.getFightState(w, r)
}

func (s *Server) startHold(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	var body struct {
		Corner  string `json:"corner"`
		Seconds int    `json:"seconds"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Corner != "red" && body.Corner != "blue" {
		writeBadRequest(w, "corner must be red or blue")
		return
	}
	if body.Seconds < 1 || body.Seconds > 120 {
		body.Seconds = 20
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO fight_state (fight_id, hold_seconds, hold_corner) VALUES (?, ?, ?)
		ON CONFLICT(fight_id) DO UPDATE SET hold_seconds = excluded.hold_seconds,
		  hold_corner = excluded.hold_corner`, id, body.Seconds, body.Corner); err != nil {
		writeError(w, err)
		return
	}
	payload, _ := json.Marshal(body)
	_, _ = s.Store.ExecContext(r.Context(), `
		INSERT INTO score_events (fight_id, seq, kind, corner, payload_json)
		VALUES (?, COALESCE((SELECT MAX(seq)+1 FROM score_events WHERE fight_id=?),1), 'hold', ?, ?)`,
		id, id, body.Corner, string(payload))
	s.audit(r, "fight.hold", "fight", id, string(payload))
	s.broadcastFight(r, id, "fight.hold")
	s.getFightState(w, r)
}

func (s *Server) undoFightEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	var nullableSeq sql.NullInt64
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT MAX(seq) FROM score_events WHERE fight_id = ?`, id).Scan(&nullableSeq); err != nil {
		writeError(w, err)
		return
	}
	if !nullableSeq.Valid {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "nothing to undo"})
		return
	}
	seq := int(nullableSeq.Int64)
	if _, err := s.Store.ExecContext(r.Context(),
		`DELETE FROM score_events WHERE fight_id = ? AND seq = ?`, id, seq); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "fight.undo", "fight", id, fmt.Sprintf(`{"seq":%d}`, seq))
	s.broadcastFight(r, id, "fight.undo")
	s.getFightState(w, r)
}

type submitRequest struct {
	Outcome string `json:"outcome"`
}

func (s *Server) submitFight(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !s.requireAssigned(w, r, auth.ActionScoreFight, id) {
		return
	}
	var body submitRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	valid := map[string]bool{"win_red": true, "win_blue": true, "draw": true}
	if !valid[body.Outcome] {
		writeBadRequest(w, "outcome must be win_red, win_blue or draw")
		return
	}
	_, err := s.Store.ExecContext(r.Context(), `
		UPDATE fights SET state='complete', outcome=?, completed_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id=?`, body.Outcome, id)
	if err != nil {
		writeError(w, err)
		return
	}
	_, _ = s.Store.ExecContext(r.Context(), `UPDATE fight_state SET running=0 WHERE fight_id=?`, id)
	s.audit(r, "fight.submit", "fight", id, fmt.Sprintf(`{"outcome":%q}`, body.Outcome))
	s.broadcastFight(r, id, "fight.complete")
	s.getFightState(w, r)
}

func (s *Server) audit(r *http.Request, action, targetType, targetID, after string) {
	userID, role := "system", "system"
	if identity, ok := auth.FromContext(r.Context()); ok {
		userID, role = identity.UserID, string(identity.Role)
	}
	_, _ = s.Store.ExecContext(r.Context(), `
		INSERT INTO audit_log
		 (tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		 VALUES ('demo-tenant', ?, ?, ?, ?, ?, ?)`,
		userID, role, action, targetType, targetID, after)
}

// StartClock decrements running fights once per second and broadcasts ticks.
func (s *Server) StartClock(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tickClock(ctx)
			}
		}
	}()
}

func (s *Server) tickClock(ctx context.Context) {
	rows, err := s.Store.QueryContext(ctx,
		`SELECT fight_id FROM fight_state WHERE running = 1 OR hold_seconds > 0`)
	if err != nil {
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		_, _ = s.Store.ExecContext(ctx, `
			UPDATE fight_state
			SET seconds_remaining = CASE WHEN running = 1 AND ends_at <> ''
			  THEN MAX(0, CAST((julianday(ends_at) - julianday('now')) * 86400 AS INTEGER))
			  ELSE seconds_remaining END,
			  running = CASE WHEN running = 1 AND ends_at <> ''
			    AND julianday(ends_at) <= julianday('now') THEN 0 ELSE running END,
			  hold_seconds = MAX(0, hold_seconds - 1),
			  hold_corner = CASE WHEN hold_seconds <= 1 THEN '' ELSE hold_corner END
			WHERE fight_id = ?`, id)
		state, err := s.fightState(ctx, id)
		if err == nil {
			s.Hub.broadcast(liveEvent{Type: "clock", FightID: id, Payload: state})
		}
	}
}

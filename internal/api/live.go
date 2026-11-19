package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
)

// requireRole enforces RBAC on live scoring mutations. The demo defaults to
// official; clients may pass X-J3-Role to exercise other roles.
func (s *Server) requireRole(w http.ResponseWriter, r *http.Request, action auth.Action) bool {
	role := auth.Role(r.Header.Get("X-J3-Role"))
	if role == "" {
		role = auth.RoleOfficial
	}
	if !auth.Can(role, action) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "your role cannot perform this action"})
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
	err := s.Store.QueryRowContext(ctx, `
		SELECT f.state, f.outcome, COALESCE(fs.running,0), COALESCE(fs.seconds_remaining,180),
		       COALESCE(fs.round,1), COALESCE(fs.hold_seconds,0), COALESCE(fs.hold_corner,'')
		FROM fights f LEFT JOIN fight_state fs ON fs.fight_id = f.id
		WHERE f.id = ?`, fightID).
		Scan(&view.State, &view.Outcome, &running, &view.Seconds, &view.Round,
			&view.HoldSeconds, &view.HoldCorner)
	if err != nil {
		return view, err
	}
	view.Running = running == 1
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
	var body timerRequest
	_ = decodeJSON(r, &body)
	seconds := body.Seconds
	if seconds < 10 || seconds > 3600 {
		seconds = 180
	}
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO fight_state (fight_id, seconds_remaining, running, round)
		VALUES (?, ?, 1, ?)
		ON CONFLICT(fight_id) DO UPDATE SET running = 1,
		  seconds_remaining = CASE WHEN excluded.seconds_remaining > 0
		    THEN excluded.seconds_remaining ELSE fight_state.seconds_remaining END,
		  round = excluded.round`,
		id, seconds, max(1, body.Round))
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
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO fight_state (fight_id, running) VALUES (?, ?)
		ON CONFLICT(fight_id) DO UPDATE SET running = excluded.running`, id, value)
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
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
	if actor := r.Header.Get("X-J3-User"); actor != "" && !s.isAssigned(r, id, actor) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user is not assigned to this contest"})
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
	var seq int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT COALESCE(MAX(seq),0)+1 FROM score_events WHERE fight_id = ?`, id).Scan(&seq); err != nil {
		writeError(w, err)
		return
	}
	payload, _ := json.Marshal(body)
	if _, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO score_events
		 (fight_id, seq, kind, corner, points, technique, card, actor_user_id, payload_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'u-official', ?)`,
		id, seq, kind, body.Corner, body.Points, body.Technique, body.Card, string(payload)); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "fight."+kind, "fight", id, string(payload))
	s.broadcastFight(r, id, "fight.event")
	s.getFightState(w, r)
}

func (s *Server) startHold(w http.ResponseWriter, r *http.Request) {
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
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
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
	var seq int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT MAX(seq) FROM score_events WHERE fight_id = ?`, id).Scan(&seq); err != nil {
		writeError(w, err)
		return
	}
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
	if !s.requireRole(w, r, auth.ActionScoreFight) {
		return
	}
	id := r.PathValue("id")
	if actor := r.Header.Get("X-J3-User"); actor != "" && !s.isAssigned(r, id, actor) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user is not assigned to this contest"})
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
	_, _ = s.Store.ExecContext(r.Context(), `
		INSERT INTO audit_log
		 (tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		 VALUES ('demo-tenant', 'u-official', 'official', ?, ?, ?, ?)`,
		action, targetType, targetID, after)
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
		`SELECT fight_id FROM fight_state WHERE (running = 1 AND seconds_remaining > 0) OR hold_seconds > 0`)
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
			SET seconds_remaining = CASE WHEN running = 1 AND seconds_remaining > 0
			  THEN MAX(0, seconds_remaining - 1) ELSE seconds_remaining END,
			  running = CASE WHEN running = 1 AND seconds_remaining <= 1 THEN 0 ELSE running END,
			  hold_seconds = MAX(0, hold_seconds - 1),
			  hold_corner = CASE WHEN hold_seconds <= 1 THEN '' ELSE hold_corner END
			WHERE fight_id = ?`, id)
		state, err := s.fightState(ctx, id)
		if err == nil {
			s.Hub.broadcast(liveEvent{Type: "clock", FightID: id, Payload: state})
		}
	}
}

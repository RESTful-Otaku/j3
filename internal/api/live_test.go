package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
)

func TestAuthoritativeFightState(t *testing.T) {
	srv := seededServer(t)
	start := postJSON(t, srv, "/api/v1/fights/fight-03/start", map[string]any{"seconds": 120, "round": 2})
	if start.Code != http.StatusOK {
		t.Fatalf("start = %d: %s", start.Code, start.Body.String())
	}
	var state api.FightStateView
	if err := json.Unmarshal(start.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Running || state.Seconds != 120 || state.Round != 2 {
		t.Fatalf("unexpected start state: %+v", state)
	}

	score := postJSON(t, srv, "/api/v1/fights/fight-03/event", map[string]any{
		"corner": "red", "points": 2, "technique": "Takedown", "kind": "score",
	})
	if score.Code != http.StatusOK {
		t.Fatalf("score = %d", score.Code)
	}
	if err := json.Unmarshal(score.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.RedScore != 2 {
		t.Fatalf("red score = %d", state.RedScore)
	}

	penalty := postJSON(t, srv, "/api/v1/fights/fight-03/event", map[string]any{
		"corner": "blue", "card": "yellow", "kind": "penalty",
	})
	if penalty.Code != http.StatusOK {
		t.Fatalf("penalty = %d", penalty.Code)
	}
	if err := json.Unmarshal(penalty.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.BluePenalty != 1 {
		t.Fatalf("blue penalties = %d", state.BluePenalty)
	}

	undo := postJSON(t, srv, "/api/v1/fights/fight-03/undo", map[string]any{})
	if undo.Code != http.StatusOK {
		t.Fatalf("undo = %d", undo.Code)
	}
	if err := json.Unmarshal(undo.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.BluePenalty != 0 || state.RedScore != 2 {
		t.Fatalf("undo applied incorrectly: %+v", state)
	}

	pause := postJSON(t, srv, "/api/v1/fights/fight-03/pause", map[string]any{})
	if err := json.Unmarshal(pause.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Running {
		t.Fatal("fight should be paused")
	}

	submit := postJSON(t, srv, "/api/v1/fights/fight-03/submit", map[string]any{"outcome": "win_red"})
	if submit.Code != http.StatusOK {
		t.Fatalf("submit = %d", submit.Code)
	}
	if err := json.Unmarshal(submit.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.State != "complete" || state.Outcome != "win_red" {
		t.Fatalf("unexpected final state: %+v", state)
	}
}

func TestHoldTimerAndRoleEnforcement(t *testing.T) {
	srv := seededServer(t)
	hold := postJSON(t, srv, "/api/v1/fights/fight-05/hold", map[string]any{"corner": "blue", "seconds": 15})
	if hold.Code != http.StatusOK {
		t.Fatalf("hold = %d: %s", hold.Code, hold.Body.String())
	}
	var state api.FightStateView
	if err := json.Unmarshal(hold.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.HoldCorner != "blue" || state.HoldSeconds != 15 {
		t.Fatalf("unexpected hold state: %+v", state)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fights/fight-05/submit",
		bytes.NewReader([]byte(`{"outcome":"win_red"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-J3-Role", "fighter")
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("fighter submit = %d, want 403", rec.Code)
	}
}

func TestOfficialAssignmentEnforced(t *testing.T) {
	srv := seededServer(t)
	assigned := postJSON(t, srv, "/api/v1/demo/fights/fight-06/officials", map[string]any{
		"userId": "u-official", "role": "referee",
	})
	if assigned.Code != http.StatusOK {
		t.Fatalf("assign = %d", assigned.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fights/fight-06/event",
		bytes.NewReader([]byte(`{"corner":"red","points":2,"kind":"score"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-J3-User", "u-runner")
	denied := httptest.NewRecorder()
	srv.Handler().ServeHTTP(denied, req)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("unassigned scorer = %d, want 403", denied.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/fights/fight-06/event",
		bytes.NewReader([]byte(`{"corner":"red","points":2,"kind":"score"}`)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-J3-User", "u-official")
	allowed := httptest.NewRecorder()
	srv.Handler().ServeHTTP(allowed, req2)
	if allowed.Code != http.StatusOK {
		t.Fatalf("assigned scorer = %d", allowed.Code)
	}
}

func TestFightStateRejectsInvalidInput(t *testing.T) {
	srv := seededServer(t)
	if rec := postJSON(t, srv, "/api/v1/fights/fight-02/event", map[string]any{"corner": "green", "points": 1}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid corner = %d", rec.Code)
	}
	if rec := postJSON(t, srv, "/api/v1/fights/fight-02/submit", map[string]any{"outcome": "maybe"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid outcome = %d", rec.Code)
	}
}

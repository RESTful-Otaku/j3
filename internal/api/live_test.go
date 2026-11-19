package api_test

import (
	"encoding/json"
	"net/http"
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

func TestFightStateRejectsInvalidInput(t *testing.T) {
	srv := seededServer(t)
	if rec := postJSON(t, srv, "/api/v1/fights/fight-02/event", map[string]any{"corner": "green", "points": 1}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid corner = %d", rec.Code)
	}
	if rec := postJSON(t, srv, "/api/v1/fights/fight-02/submit", map[string]any{"outcome": "maybe"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid outcome = %d", rec.Code)
	}
}

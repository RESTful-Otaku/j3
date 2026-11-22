package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func okamiAuthServer(t *testing.T) *api.Server {
	t.Helper()
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureOkami(store); err != nil {
		t.Fatal(err)
	}
	return api.New(store, "test")
}

func getJSONPath(t *testing.T, srv *api.Server, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestTechniqueCatalogueSearchAndConfiguredPoints(t *testing.T) {
	srv := okamiAuthServer(t)
	rec := getJSONPath(t, srv, "/api/v1/demo/techniques?eventId=ev-okami&q=seoi", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Count      int `json:"count"`
		Techniques []struct {
			ID             string   `json:"id"`
			English        string   `json:"english"`
			Romaji         string   `json:"romaji"`
			Japanese       string   `json:"japanese"`
			Group          string   `json:"group"`
			PointKey       string   `json:"pointKey"`
			Points         int      `json:"points"`
			ResolvedPoints int      `json:"resolvedPoints"`
			Tags           []string `json:"tags"`
			InstantWin     bool     `json:"instantWin"`
		} `json:"techniques"`
		Groups []struct {
			ID string `json:"id"`
		} `json:"groups"`
		PointOptions []struct {
			Key    string `json:"key"`
			Label  string `json:"label"`
			Value  int    `json:"value"`
			Source string `json:"source"`
		} `json:"pointOptions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Count < 5 || len(body.Techniques) != body.Count {
		t.Fatalf("seoi search returned %d results", body.Count)
	}
	if len(body.Groups) < 20 {
		t.Fatalf("only %d technique groups", len(body.Groups))
	}
	if len(body.PointOptions) == 0 {
		t.Fatal("expected point options derived from Okami rules")
	}
	found := false
	for _, item := range body.Techniques {
		if item.ID == "seoi_nage" {
			found = true
			if item.Romaji != "Seoi-nage" || item.Japanese == "" {
				t.Fatalf("seoi_nage metadata incomplete: %+v", item)
			}
			if item.ResolvedPoints != 3 {
				t.Fatalf("seoi_nage resolved points = %d, want 3 from discipline rules", item.ResolvedPoints)
			}
		}
	}
	if !found {
		t.Fatal("seoi_nage missing from search results")
	}

	groupRec := getJSONPath(t, srv, "/api/v1/demo/techniques?eventId=ev-okami&group=kicks", "")
	var grouped struct {
		Techniques []struct {
			Group string `json:"group"`
		} `json:"techniques"`
	}
	if err := json.Unmarshal(groupRec.Body.Bytes(), &grouped); err != nil {
		t.Fatal(err)
	}
	if len(grouped.Techniques) == 0 {
		t.Fatal("expected kick techniques")
	}
	for _, item := range grouped.Techniques {
		if item.Group != "kicks" {
			t.Fatalf("group filter leaked %q", item.Group)
		}
	}
}

func TestOfficialScoringResolvesCataloguePointsServerSide(t *testing.T) {
	srv := okamiAuthServer(t)
	token := sessionToken(t, srv, "u-keeper-1", "official")
	rec := postAuthenticatedJSON(t, srv, token, "/api/v1/fights/ok-fight-11/event", map[string]any{
		"corner":      "red",
		"techniqueId": "seoi_nage",
		"points":      99,
		"technique":   "client supplied label",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var state struct {
		Events []struct {
			Kind      string `json:"kind"`
			Points    int    `json:"points"`
			Technique string `json:"technique"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	last := state.Events[len(state.Events)-1]
	if last.Points != 3 {
		t.Fatalf("last event points = %d, want configured 3", last.Points)
	}
	if last.Technique != "Shoulder throw (Seoi-nage)" {
		t.Fatalf("last event technique = %q", last.Technique)
	}
}

func TestPenaltyCardPointsAreAttributedAndCounted(t *testing.T) {
	srv := okamiAuthServer(t)
	token := sessionToken(t, srv, "u-keeper-2", "official")
	rec := postAuthenticatedJSON(t, srv, token, "/api/v1/fights/ok-fight-12/event", map[string]any{
		"corner":      "blue",
		"kind":        "penalty",
		"card":        "red",
		"techniqueId": "red_card",
		"points":      -2,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var state struct {
		BluePenalty       int `json:"bluePenalty"`
		BluePenaltyPoints int `json:"bluePenaltyPoints"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.BluePenalty != 1 || state.BluePenaltyPoints != -2 {
		t.Fatalf("blue penalties = %d, points = %d; want 1 and -2", state.BluePenalty, state.BluePenaltyPoints)
	}
}

func TestTechnicalResultRequiresAssignmentAndRecordsReason(t *testing.T) {
	srv := okamiAuthServer(t)
	unassigned := sessionToken(t, srv, "u-judge-4", "official")
	rec := postAuthenticatedJSON(t, srv, unassigned, "/api/v1/fights/ok-fight-13/technical", map[string]any{
		"winner": "blue", "reason": "injury",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unassigned status = %d, want 403: %s", rec.Code, rec.Body.String())
	}

	assigned := sessionToken(t, srv, "u-keeper-1", "official")
	rec = postAuthenticatedJSON(t, srv, assigned, "/api/v1/fights/ok-fight-13/technical", map[string]any{
		"winner": "blue", "reason": "injury", "detail": "shoulder injury",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("assigned status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var state struct {
		State   string `json:"state"`
		Outcome string `json:"outcome"`
		Events  []struct {
			Kind      string `json:"kind"`
			Technique string `json:"technique"`
		} `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.State != "forfeit" || state.Outcome != "win_blue" {
		t.Fatalf("technical outcome = %s/%s", state.State, state.Outcome)
	}
	last := state.Events[len(state.Events)-1]
	if last.Kind != "technical" || !bytes.Contains([]byte(last.Technique), []byte("Injury")) {
		t.Fatalf("technical event = %+v", last)
	}

	rec = postAuthenticatedJSON(t, srv, assigned, "/api/v1/fights/ok-fight-17/technical", map[string]any{
		"winner": "red", "reason": "cheating",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("dq status = %d: %s", rec.Code, rec.Body.String())
	}
	var dq struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dq); err != nil {
		t.Fatal(err)
	}
	if dq.Outcome != "dq_red" {
		t.Fatalf("cheating outcome = %q, want dq_red", dq.Outcome)
	}
}

func TestMyContestsFiltersBySessionIdentity(t *testing.T) {
	srv := okamiAuthServer(t)
	runner := getJSONPath(t, srv, "/api/v1/demo/my-contests", sessionToken(t, srv, "u-tony", "tournament_runner"))
	official := getJSONPath(t, srv, "/api/v1/demo/my-contests", sessionToken(t, srv, "u-keeper-1", "official"))
	fighter := getJSONPath(t, srv, "/api/v1/demo/my-contests", sessionToken(t, srv, "ok-u-f01", "fighter"))
	if runner.Code != http.StatusOK || official.Code != http.StatusOK || fighter.Code != http.StatusOK {
		t.Fatalf("statuses runner=%d official=%d fighter=%d", runner.Code, official.Code, fighter.Code)
	}
	decode := func(rec *httptest.ResponseRecorder) int {
		var body struct {
			Contests []json.RawMessage `json:"contests"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return len(body.Contests)
	}
	all, assigned, none := decode(runner), decode(official), decode(fighter)
	if all < 24 {
		t.Fatalf("runner contests = %d, want at least 24", all)
	}
	if assigned == 0 || assigned >= all {
		t.Fatalf("official contests = %d, want a useful assigned subset", assigned)
	}
	if none != 0 {
		t.Fatalf("fighter contests = %d, want 0", none)
	}
}

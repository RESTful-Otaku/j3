package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
	"github.com/restful-otaku/j3/internal/auth"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func newTestServer(t *testing.T) *api.Server {
	t.Helper()
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return api.New(store, "test")
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %q", body["status"])
	}
}

func TestReadyz(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRoles(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil))
	var body struct {
		Roles []struct {
			ID string `json:"id"`
		} `json:"roles"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Roles) != 5 {
		t.Fatalf("roles = %d, want 5", len(body.Roles))
	}
}

func postAuthenticatedJSON(t *testing.T, srv *api.Server, token, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestDemoEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/demo", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func seededServer(t *testing.T) *api.Server {
	t.Helper()
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	srv := api.New(store, "test")
	srv.EnableTestBypass()
	return srv
}

// seededAuthServer returns a server with real bearer authentication (no test
// bypass) so authorization tests exercise the production path.
func seededAuthServer(t *testing.T) *api.Server {
	t.Helper()
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	return api.New(store, "test")
}

func sessionToken(t *testing.T, srv *api.Server, userID, role string) string {
	t.Helper()
	token := "tok-" + userID + "-" + role
	_, err := srv.Store.Exec(
		`INSERT OR REPLACE INTO sessions (token, user_id, role, expires_at)
		 VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now','+1 hour'))`,
		token, userID, role)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestDemoAccountsCoverEveryRole(t *testing.T) {
	rec := httptest.NewRecorder()
	seededServer(t).Handler().ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/api/v1/demo/accounts", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Accounts []struct {
			Role auth.Role `json:"role"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	seen := map[auth.Role]bool{}
	for _, a := range body.Accounts {
		seen[a.Role] = true
	}
	for _, role := range []auth.Role{
		auth.RoleTournamentRunner, auth.RoleOfficial, auth.RoleClubLeader,
		auth.RoleFighter, auth.RoleSpectator,
	} {
		if !seen[role] {
			t.Errorf("no demo account for role %s", role)
		}
	}
}

func TestDemoFights(t *testing.T) {
	rec := httptest.NewRecorder()
	seededServer(t).Handler().ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/api/v1/demo/fights", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Fights []struct {
			ID    string `json:"id"`
			State string `json:"state"`
			Red   struct {
				Name string `json:"name"`
			} `json:"red"`
			Blue struct {
				Name string `json:"name"`
			} `json:"blue"`
			Events []json.RawMessage `json:"events"`
		} `json:"fights"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Fights) == 0 {
		t.Fatal("expected demo fights")
	}
	foundLive := false
	for _, f := range body.Fights {
		if f.Red.Name == "" || f.Blue.Name == "" {
			t.Fatalf("fight %s missing fighter names", f.ID)
		}
		if f.State == "live" {
			foundLive = true
			if len(f.Events) == 0 {
				t.Fatalf("live fight %s has no scoring events", f.ID)
			}
		}
	}
	if !foundLive {
		t.Fatal("expected a live fight in demo data")
	}
}

func TestDemoViewResolvesActorCapabilities(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	srv := seededAuthServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/view?actor=u-official", nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken(t, srv, "u-runner", "tournament_runner"))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Actor struct {
			Role auth.Role `json:"role"`
		} `json:"actor"`
		Actions []auth.Action `json:"actions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Actor.Role != auth.RoleOfficial {
		t.Fatalf("role = %s", body.Actor.Role)
	}
	found := false
	for _, a := range body.Actions {
		if a == auth.ActionScoreFight {
			found = true
		}
	}
	if !found {
		t.Fatal("official should be able to score fights")
	}
	for _, a := range body.Actions {
		if a == auth.ActionManageEvent {
			t.Fatal("official must not be able to manage the event")
		}
	}
}

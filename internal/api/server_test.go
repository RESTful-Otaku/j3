package api_test

import (
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
	return api.New(store, "test")
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

func TestDemoViewResolvesActorCapabilities(t *testing.T) {
	rec := httptest.NewRecorder()
	seededServer(t).Handler().ServeHTTP(rec,
		httptest.NewRequest(http.MethodGet, "/api/v1/demo/view?actor=u-official", nil))
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

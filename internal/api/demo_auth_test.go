package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

type demoSessionResponse struct {
	Token     string `json:"token"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	UserID    string `json:"userId"`
	ExpiresAt string `json:"expiresAt"`
}

func demoLoginRequest(t *testing.T, srv *api.Server) (*httptest.ResponseRecorder, demoSessionResponse) {
	t.Helper()
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/demo", nil))
	var session demoSessionResponse
	if rec.Code == http.StatusOK && json.Unmarshal(rec.Body.Bytes(), &session) != nil {
		t.Fatalf("invalid demo session response: %s", rec.Body.String())
	}
	return rec, session
}

func authenticatedGet(t *testing.T, srv *api.Server, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func demoLoginAs(t *testing.T, srv *api.Server, userID string) (*httptest.ResponseRecorder, demoSessionResponse) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/demo", strings.NewReader(`{"userId":"`+userID+`"}`))
	req.Header.Set("Content-Type", "application/json")
	srv.Handler().ServeHTTP(rec, req)
	var session demoSessionResponse
	if rec.Code == http.StatusOK && json.Unmarshal(rec.Body.Bytes(), &session) != nil {
		t.Fatalf("invalid demo session response: %s", rec.Body.String())
	}
	return rec, session
}

func TestDemoLoginBootstrapsEmptyVenueDatabase(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	srv := newTestServer(t)
	rec, session := demoLoginRequest(t, srv)
	if rec.Code != http.StatusOK {
		t.Fatalf("demo login status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if session.Token == "" || session.ExpiresAt == "" || session.UserID == "" {
		t.Fatalf("incomplete session: %+v", session)
	}
	var userID string
	if err := srv.Store.QueryRow(`SELECT user_id FROM sessions WHERE token = ?`, session.Token).Scan(&userID); err != nil {
		t.Fatalf("session was not persisted: %v", err)
	}
	var role string
	if err := srv.Store.QueryRow(`SELECT role FROM users WHERE id = ?`, userID).Scan(&role); err != nil {
		t.Fatalf("session points at a missing user: %v", err)
	}
	if role != "tournament_runner" {
		t.Fatalf("bootstrapped role = %q, want tournament_runner", role)
	}
	if got := authenticatedGet(t, srv, "/api/v1/demo/accounts", session.Token); got.Code != http.StatusOK {
		t.Fatalf("accounts status = %d, want 200: %s", got.Code, got.Body.String())
	}
}

func TestDemoLoginUsesBaseSeedRunnerAndActiveEvent(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	srv := api.New(store, "test")
	rec, session := demoLoginRequest(t, srv)
	if rec.Code != http.StatusOK {
		t.Fatalf("demo login status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if session.UserID != "u-runner" {
		t.Fatalf("userId = %q, want seeded u-runner", session.UserID)
	}
	for _, path := range []string{
		"/api/v1/demo/accounts", "/api/v1/demo/event", "/api/v1/demo/event-config",
		"/api/v1/demo/schedule", "/api/v1/demo/medals",
	} {
		got := authenticatedGet(t, srv, path, session.Token)
		if got.Code != http.StatusOK {
			t.Errorf("%s status = %d, want 200: %s", path, got.Code, got.Body.String())
		}
	}
}

func TestDemoLoginDisabledOutsideDemoMode(t *testing.T) {
	t.Setenv("J3_DEMO", "0")
	rec, _ := demoLoginRequest(t, newTestServer(t))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDemoLoginMintsRealSessionForChosenUser(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	srv := seededAuthServer(t)
	rec, session := demoLoginAs(t, srv, "u-official")
	if rec.Code != http.StatusOK {
		t.Fatalf("demo login status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if session.UserID != "u-official" || session.Role != "official" {
		t.Fatalf("session identity = %s/%s, want u-official/official", session.UserID, session.Role)
	}
	var storedRole string
	if err := srv.Store.QueryRow(`SELECT role FROM sessions WHERE token = ?`, session.Token).Scan(&storedRole); err != nil {
		t.Fatalf("session was not persisted: %v", err)
	}
	if storedRole != "official" {
		t.Fatalf("stored role = %q, want canonical official", storedRole)
	}
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tournaments", strings.NewReader(`{}`))
	createReq.Header.Set("Authorization", "Bearer "+session.Token)
	createReq.Header.Set("Content-Type", "application/json")
	srv.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusForbidden {
		t.Fatalf("settings:create status = %d, want 403: %s", createRec.Code, createRec.Body.String())
	}
	if got := authenticatedGet(t, srv, "/api/v1/demo/standings", session.Token); got.Code != http.StatusOK {
		t.Fatalf("public standings status = %d, want 200: %s", got.Code, got.Body.String())
	}
	identityRec := authenticatedGet(t, srv, "/api/v1/demo/identities", session.Token)
	if identityRec.Code != http.StatusOK {
		t.Fatalf("identities status = %d, want 200: %s", identityRec.Code, identityRec.Body.String())
	}
	var identityBody struct {
		Current *struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"current"`
	}
	if err := json.Unmarshal(identityRec.Body.Bytes(), &identityBody); err != nil {
		t.Fatal(err)
	}
	if identityBody.Current == nil || identityBody.Current.ID != "u-official" || identityBody.Current.Role != "official" {
		t.Fatalf("identities.current = %+v, want the verified official session", identityBody.Current)
	}
}

func TestDemoIdentitiesArePublicOnlyInDemoMode(t *testing.T) {
	srv := seededAuthServer(t)
	t.Setenv("J3_DEMO", "1")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/demo/identities", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("identities status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Identities []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Role      string `json:"role"`
			RoleLabel string `json:"roleLabel"`
		} `json:"identities"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Identities) < 5 {
		t.Fatalf("identities = %d, want at least 5 seeded roles", len(body.Identities))
	}
	if strings.Contains(rec.Body.String(), `"email"`) {
		t.Fatal("identity directory must not expose private account fields")
	}
	t.Setenv("J3_DEMO", "0")
	disabled := httptest.NewRecorder()
	srv.Handler().ServeHTTP(disabled, httptest.NewRequest(http.MethodGet, "/api/v1/demo/identities", nil))
	if disabled.Code != http.StatusNotFound {
		t.Fatalf("disabled identities status = %d, want 404", disabled.Code)
	}
}

func TestSessionExpiryPrecisionIsAccepted(t *testing.T) {
	srv := seededAuthServer(t)
	for name, expiry := range map[string]string{
		"milliseconds":  "2099-01-02T03:04:05.123Z",
		"whole seconds": "2099-01-02T03:04:05Z",
		"microseconds":  "2099-01-02T03:04:05.123456Z",
	} {
		t.Run(name, func(t *testing.T) {
			token := "tok-" + name
			if _, err := srv.Store.Exec(
				`INSERT INTO sessions (token, user_id, role, expires_at) VALUES (?, 'u-runner', 'tournament_runner', ?)`,
				token, expiry); err != nil {
				t.Fatal(err)
			}
			rec := authenticatedGet(t, srv, "/api/v1/demo/accounts", token)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
			}
		})
	}
	if _, err := srv.Store.Exec(
		`INSERT INTO sessions (token, user_id, role, expires_at)
		 VALUES ('expired-token', 'u-runner', 'tournament_runner', '2000-01-01T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	for name, token := range map[string]string{"expired": "expired-token", "unknown": "no-such-token"} {
		rec := authenticatedGet(t, srv, "/api/v1/demo/accounts", token)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s token status = %d, want 401", name, rec.Code)
		}
	}
}

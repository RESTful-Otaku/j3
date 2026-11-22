package api_test

import (
	"net/http"
	"testing"
)

func TestSuspendedSessionIsRejected(t *testing.T) {
	srv := seededAuthServer(t)
	token := sessionToken(t, srv, "u-runner", "tournament_runner")
	if rec := getJSONPath(t, srv, "/api/v1/demo/view", token); rec.Code != http.StatusOK {
		t.Fatalf("active session = %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := srv.Store.Exec(`UPDATE users SET status = 'suspended' WHERE id = 'u-runner'`); err != nil {
		t.Fatal(err)
	}
	if rec := getJSONPath(t, srv, "/api/v1/demo/view", token); rec.Code != http.StatusUnauthorized {
		t.Fatalf("suspended session = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

func TestRunnerCannotRewriteAuthorityMatrix(t *testing.T) {
	srv := seededAuthServer(t)
	runner := sessionToken(t, srv, "u-runner", "tournament_runner")
	admin := sessionToken(t, srv, "u-owner", "admin")
	body := map[string]any{"role": "fighter", "resource": "fights", "flag": "delete", "value": true}
	if rec := postAuthenticatedJSON(t, srv, runner, "/api/v1/demo/permissions", body); rec.Code != http.StatusForbidden {
		t.Fatalf("runner update = %d, want 403: %s", rec.Code, rec.Body.String())
	}
	if rec := postAuthenticatedJSON(t, srv, admin, "/api/v1/demo/permissions", body); rec.Code != http.StatusOK {
		t.Fatalf("admin update = %d: %s", rec.Code, rec.Body.String())
	}
	invalid := map[string]any{"role": "admin", "resource": "unknown", "flag": "delete", "value": true}
	if rec := postAuthenticatedJSON(t, srv, admin, "/api/v1/demo/permissions", invalid); rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown resource = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestStaleSessionRoleIsIgnored(t *testing.T) {
	srv := seededAuthServer(t)
	var userID string
	if err := srv.Store.QueryRow(`SELECT id FROM users WHERE role = 'fighter' LIMIT 1`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	token := sessionToken(t, srv, userID, "admin")
	body := map[string]any{"role": "fighter", "resource": "fights", "flag": "delete", "value": true}
	if rec := postAuthenticatedJSON(t, srv, token, "/api/v1/demo/permissions", body); rec.Code != http.StatusForbidden {
		t.Fatalf("stale admin session = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

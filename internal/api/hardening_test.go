package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
)

func authedRequest(srv *api.Server, token, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestFighterCannotReadPermissionsOrAdminAccounts(t *testing.T) {
	srv := seededAuthServer(t)
	token := sessionToken(t, srv, "u-fighter-01", "fighter")
	for _, path := range []string{"/api/v1/demo/permissions", "/api/v1/demo/admin/users", "/api/v1/demo/audit"} {
		rec := authedRequest(srv, token, http.MethodGet, path, "")
		if rec.Code != http.StatusForbidden {
			t.Errorf("fighter GET %s = %d, want 403", path, rec.Code)
		}
	}
}

func TestDuplicateSignupEmailIsConflict(t *testing.T) {
	srv := seededServer(t)
	body := map[string]any{
		"name": "Runner Clone", "email": "runner@demo.j3", "password": "password123",
		"dateOfBirth": "1990-01-01", "unitSystem": "metric", "heightCm": 175, "weightKg": 70,
		"accountType": "fighter",
	}
	if rec := postJSON(t, srv, "/api/v1/signup", body); rec.Code != http.StatusConflict {
		t.Fatalf("duplicate signup = %d, want 409: %s", rec.Code, rec.Body.String())
	}
}

func TestPasswordResetFlow(t *testing.T) {
	srv := seededAuthServer(t)
	admin := sessionToken(t, srv, "u-runner", "tournament_runner")
	issued := authedRequest(srv, admin, http.MethodPost,
		"/api/v1/demo/admin/users/u-fighter-01/reset", "{}")
	if issued.Code != http.StatusOK {
		t.Fatalf("issue reset = %d", issued.Code)
	}
	var payload struct {
		ResetToken string `json:"resetToken"`
	}
	if err := json.Unmarshal(issued.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ResetToken == "" {
		t.Fatal("no reset token issued")
	}
	reset := authedRequest(srv, "", http.MethodPost, "/api/v1/auth/reset",
		`{"token":"`+payload.ResetToken+`","password":"newpassword123"}`)
	if reset.Code != http.StatusOK {
		t.Fatalf("complete reset = %d: %s", reset.Code, reset.Body.String())
	}
	again := authedRequest(srv, "", http.MethodPost, "/api/v1/auth/reset",
		`{"token":"`+payload.ResetToken+`","password":"anotherpass123"}`)
	if again.Code != http.StatusBadRequest {
		t.Fatalf("replayed reset = %d, want 400", again.Code)
	}
}

func TestRebuildingCategoryBracketKeepsOtherCategory(t *testing.T) {
	srv := seededServer(t)
	admin := sessionToken(t, srv, "u-runner", "tournament_runner")
	first := authedRequest(srv, admin, http.MethodPost, "/api/v1/demo/brackets/build",
		`{"eventId":"ev-demo","categoryId":"cat-adult-76","format":"Double elimination","seed":1}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("first build = %d", first.Code)
	}
	second := authedRequest(srv, admin, http.MethodPost, "/api/v1/demo/brackets/build",
		`{"eventId":"ev-demo","categoryId":"cat-youth-60","format":"Round robin","seed":2}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("second build = %d: %s", second.Code, second.Body.String())
	}
	var count int
	if err := srv.Store.QueryRow(
		`SELECT COUNT(*) FROM bracket_stages WHERE event_id='ev-demo' AND category_id='cat-adult-76'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("first category bracket was deleted")
	}
}

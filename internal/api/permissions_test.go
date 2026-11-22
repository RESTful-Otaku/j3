package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestLoginMFAAndPermissions(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	srv := seededServer(t)
	start := postJSON(t, srv, "/api/v1/auth/login", map[string]any{
		"email": "runner@demo.j3", "provider": "google",
	})
	if start.Code != http.StatusOK {
		t.Fatalf("login = %d: %s", start.Code, start.Body.String())
	}
	var challenge struct {
		DemoCode string `json:"demoCode"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &challenge); err != nil {
		t.Fatal(err)
	}
	if len(challenge.DemoCode) != 6 {
		t.Fatalf("bad code %q", challenge.DemoCode)
	}

	verified := postJSON(t, srv, "/api/v1/auth/verify", map[string]any{
		"email": "runner@demo.j3", "code": challenge.DemoCode,
	})
	if verified.Code != http.StatusOK {
		t.Fatalf("verify = %d: %s", verified.Code, verified.Body.String())
	}
	var session struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(verified.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Token == "" {
		t.Fatal("no token")
	}

	bad := postJSON(t, srv, "/api/v1/auth/verify", map[string]any{"email": "runner@demo.j3", "code": "000000"})
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("replayed code = %d", bad.Code)
	}

	adminToken := sessionToken(t, srv, "u-owner", "admin")
	perms := getJSONPath(t, srv, "/api/v1/demo/permissions?init=1", adminToken)
	if perms.Code != http.StatusOK {
		t.Fatalf("permissions = %d", perms.Code)
	}
	var body struct {
		Permissions []struct {
			Role string `json:"role"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(perms.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Permissions) == 0 {
		t.Fatal("no permissions")
	}

	updated := postAuthenticatedJSON(t, srv, adminToken, "/api/v1/demo/permissions", map[string]any{
		"role": "fighter", "resource": "fights", "flag": "delete", "value": true,
	})
	if updated.Code != http.StatusOK {
		t.Fatalf("update permission = %d: %s", updated.Code, updated.Body.String())
	}
}

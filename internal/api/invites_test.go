package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/api"
)

type inviteResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	State     string `json:"state"`
	Token     string `json:"token"`
	InviteURL string `json:"inviteUrl"`
}

func inviteAs(t *testing.T, srv *api.Server, token string, body map[string]any) (int, inviteResponse) {
	t.Helper()
	rec := postAuthenticatedJSON(t, srv, token, "/api/v1/invites", body)
	var result inviteResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &result)
	return rec.Code, result
}

func TestInviteAuthorityChainAndAcceptance(t *testing.T) {
	srv := seededAuthServer(t)
	admin := sessionToken(t, srv, "u-owner", "admin")
	runner := sessionToken(t, srv, "u-runner", "tournament_runner")
	official := sessionToken(t, srv, "u-official", "official")
	leader := sessionToken(t, srv, "u-leader-1", "club_leader")

	if code, _ := inviteAs(t, srv, admin, map[string]any{
		"email": "new-admin@demo.j3", "name": "New Admin", "role": "admin",
	}); code != http.StatusCreated {
		t.Fatalf("admin invite admin status = %d, want 201", code)
	}
	if code, _ := inviteAs(t, srv, runner, map[string]any{
		"email": "bad-admin@demo.j3", "role": "admin",
	}); code != http.StatusForbidden {
		t.Errorf("runner invite admin status = %d, want 403", code)
	}
	if code, _ := inviteAs(t, srv, runner, map[string]any{
		"email": "bad-runner@demo.j3", "role": "tournament_runner",
	}); code != http.StatusForbidden {
		t.Errorf("runner invite runner status = %d, want 403", code)
	}
	code, officialInvite := inviteAs(t, srv, runner, map[string]any{
		"email": "new-official@demo.j3", "name": "New Official", "role": "official",
	})
	if code != http.StatusCreated || officialInvite.Token == "" || officialInvite.InviteURL == "" {
		t.Fatalf("runner invite official = %d, %+v", code, officialInvite)
	}
	if code, _ := inviteAs(t, srv, official, map[string]any{
		"email": "new-owner@demo.j3", "role": "club_leader", "clubName": "New Club",
	}); code != http.StatusCreated {
		t.Errorf("official invite club owner status = %d, want 201", code)
	}
	if code, _ := inviteAs(t, srv, official, map[string]any{
		"email": "bad-official@demo.j3", "role": "official",
	}); code != http.StatusForbidden {
		t.Errorf("official invite official status = %d, want 403", code)
	}
	if code, _ := inviteAs(t, srv, official, map[string]any{
		"email": "new-spectator@demo.j3", "role": "spectator",
	}); code != http.StatusCreated {
		t.Errorf("official invite spectator status = %d, want 201", code)
	}
	if code, _ := inviteAs(t, srv, leader, map[string]any{
		"email": "own-fighter@demo.j3", "name": "Own Fighter", "role": "fighter",
	}); code != http.StatusCreated {
		t.Errorf("leader invite own fighter status = %d, want 201", code)
	}
	if code, _ := inviteAs(t, srv, leader, map[string]any{
		"email": "own-spectator@demo.j3", "role": "spectator",
	}); code != http.StatusCreated {
		t.Errorf("leader invite own spectator status = %d, want 201", code)
	}
	if code, _ := inviteAs(t, srv, leader, map[string]any{
		"email": "other-fighter@demo.j3", "role": "fighter", "clubId": "club-hilltop",
	}); code != http.StatusForbidden {
		t.Errorf("leader invite other club status = %d, want 403", code)
	}
	if code, _ := inviteAs(t, srv, leader, map[string]any{
		"email": "bad-leader@demo.j3", "role": "club_leader",
	}); code != http.StatusForbidden {
		t.Errorf("leader invite leader status = %d, want 403", code)
	}

	accept := postJSON(t, srv, "/api/v1/invites/accept", map[string]any{
		"token": officialInvite.Token, "name": "Accepted Official", "password": "newpassword123",
	})
	if accept.Code != http.StatusOK {
		t.Fatalf("accept status = %d: %s", accept.Code, accept.Body.String())
	}
	var accepted struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(accept.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted.Role != "official" || accepted.Status != "active" {
		t.Fatalf("unexpected acceptance: %+v", accepted)
	}
	accounts := getJSONPath(t, srv, "/api/v1/demo/accounts", admin)
	if accounts.Code != http.StatusOK {
		t.Fatalf("accounts status = %d", accounts.Code)
	}
	var body struct {
		Accounts []struct {
			Email  string `json:"email"`
			Role   string `json:"role"`
			Status string `json:"status"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(accounts.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, account := range body.Accounts {
		if account.Email == "new-admin@demo.j3" && account.Role != "admin" {
			t.Fatalf("invited admin role = %q, want admin", account.Role)
		}
		if account.Email == "new-official@demo.j3" &&
			(account.Role != "official" || account.Status != "active") {
			t.Fatalf("accepted official = %+v", account)
		}
	}
	audit := getJSONPath(t, srv, "/api/v1/demo/audit?action=invite.accept", admin)
	if audit.Code != http.StatusOK {
		t.Fatalf("audit status = %d", audit.Code)
	}
	var auditBody struct {
		Entries []struct {
			Action string `json:"action"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(audit.Body.Bytes(), &auditBody); err != nil {
		t.Fatal(err)
	}
	if len(auditBody.Entries) == 0 {
		t.Fatal("invite acceptance was not audited")
	}
}

func TestDemoCredentialsSheet(t *testing.T) {
	t.Setenv("J3_DEMO", "1")
	srv := seededServer(t)
	rec := getJSONPath(t, srv, "/api/v1/demo/credentials", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Password string `json:"password"`
		Total    int    `json:"total"`
		Quick    []struct {
			Role     string `json:"role"`
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"quickAccounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Password != "demo" || body.Total < 50 {
		t.Fatalf("unexpected demo sheet: password=%q total=%d", body.Password, body.Total)
	}
	seen := map[string]bool{}
	for _, account := range body.Quick {
		seen[account.Role] = true
		if account.Password != "demo" || account.Email == "" {
			t.Fatalf("incomplete quick account: %+v", account)
		}
	}
	for _, role := range []string{"admin", "tournament_runner", "official", "club_leader", "fighter", "spectator"} {
		if !seen[role] {
			t.Errorf("quick accounts missing %s", role)
		}
	}
	t.Setenv("J3_DEMO", "0")
	hidden := getJSONPath(t, srv, "/api/v1/demo/credentials", "")
	if hidden.Code != http.StatusNotFound {
		t.Errorf("production credentials status = %d, want 404", hidden.Code)
	}
}

func TestChangeOwnPasswordClearsBootstrapFlag(t *testing.T) {
	srv := seededAuthServer(t)
	token := sessionToken(t, srv, "u-owner", "admin")
	hash, err := bcrypt.GenerateFromPassword([]byte("temporary-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.Store.Exec(
		`UPDATE users SET password_hash = ?, must_change_password = 1 WHERE id = 'u-owner'`,
		string(hash)); err != nil {
		t.Fatal(err)
	}
	blocked := getJSONPath(t, srv, "/api/v1/demo/view", token)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("protected route with temporary password = %d, want 403: %s", blocked.Code, blocked.Body.String())
	}
	var problem struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(blocked.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem.Error != "password_change_required" {
		t.Fatalf("error = %q, want password_change_required", problem.Error)
	}
	wrong := postAuthenticatedJSON(t, srv, token, "/api/v1/auth/password", map[string]any{
		"currentPassword": "wrong", "newPassword": "personal-password",
	})
	if wrong.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password status = %d, want 401", wrong.Code)
	}
	ok := postAuthenticatedJSON(t, srv, token, "/api/v1/auth/password", map[string]any{
		"currentPassword": "temporary-password", "newPassword": "personal-password",
	})
	if ok.Code != http.StatusOK {
		t.Fatalf("change status = %d: %s", ok.Code, ok.Body.String())
	}
	var stored string
	var mustChange int
	if err := srv.Store.QueryRow(
		`SELECT password_hash, must_change_password FROM users WHERE id = 'u-owner'`).
		Scan(&stored, &mustChange); err != nil {
		t.Fatal(err)
	}
	if mustChange != 0 || bcrypt.CompareHashAndPassword([]byte(stored), []byte("personal-password")) != nil {
		t.Fatalf("password was not updated and flag cleared")
	}
}

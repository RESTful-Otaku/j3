package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTournamentAndAudit(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/tournaments", map[string]any{
		"name": "Autumn Cup", "title": "Autumn Cup 2026", "theme": "street",
		"startDate": "2026-11-20", "endDate": "2026-11-21", "saveTemplate": true,
		"templateName": "Autumn Cup template",
		"disciplines":  []map[string]any{{"name": "Fighting", "matchSeconds": 180, "rounds": 3, "scoring": "ippon"}},
		"categories":   []map[string]any{{"name": "Adult -76 kg", "kind": "weight", "min": 0, "max": 76, "gender": "male"}},
		"constraints":  map[string]any{"cap": 200, "maxPools": 16, "autoPlace": true, "separateClubs": true},
		"invites":      []map[string]any{{"email": "judge@example.com", "phone": "+441234", "role": "official"}},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		EventID         string `json:"eventId"`
		TemplateID      string `json:"templateId"`
		DisciplineCount int    `json:"disciplineCount"`
		InviteCount     int    `json:"inviteCount"`
		InviteLinks     []struct {
			Email     string `json:"email"`
			Role      string `json:"role"`
			Token     string `json:"token"`
			InviteURL string `json:"inviteUrl"`
		} `json:"inviteLinks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.EventID == "" || created.TemplateID == "" || created.DisciplineCount != 1 || created.InviteCount != 1 {
		t.Fatalf("unexpected result: %+v", created)
	}
	if len(created.InviteLinks) != 1 || created.InviteLinks[0].Token == "" ||
		created.InviteLinks[0].InviteURL != "/invite/"+created.InviteLinks[0].Token {
		t.Fatalf("wizard invite links = %+v", created.InviteLinks)
	}
	accepted := postJSON(t, srv, "/api/v1/invites/accept", map[string]any{
		"token": created.InviteLinks[0].Token, "name": "Judge Example", "password": "newpassword123",
	})
	if accepted.Code != http.StatusOK {
		t.Fatalf("wizard invite acceptance = %d: %s", accepted.Code, accepted.Body.String())
	}

	templates := httptest.NewRecorder()
	srv.Handler().ServeHTTP(templates, httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil))
	if templates.Code != http.StatusOK {
		t.Fatalf("templates status = %d", templates.Code)
	}

	audit := httptest.NewRecorder()
	srv.Handler().ServeHTTP(audit, httptest.NewRequest(http.MethodGet, "/api/v1/demo/audit?action=tournament.create", nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("audit status = %d", audit.Code)
	}
	var auditBody struct {
		Entries []struct {
			Action    string `json:"action"`
			ActorRole string `json:"actorRole"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(audit.Body.Bytes(), &auditBody); err != nil {
		t.Fatal(err)
	}
	if len(auditBody.Entries) == 0 || auditBody.Entries[0].Action != "tournament.create" {
		t.Fatalf("expected tournament.create audit entry, got %+v", auditBody.Entries)
	}
}

func TestImportTemplateRecreatesEvent(t *testing.T) {
	srv := seededServer(t)
	created := postJSON(t, srv, "/api/v1/tournaments", map[string]any{
		"name": "Template Event", "saveTemplate": true, "templateName": "Reusable",
		"disciplines": []map[string]any{{"name": "Ne-Waza"}},
		"categories":  []map[string]any{{"name": "Adult -70", "kind": "weight"}},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create = %d", created.Code)
	}
	var result struct {
		TemplateID string `json:"templateId"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	imported := postJSON(t, srv, "/api/v1/templates/"+result.TemplateID+"/import", map[string]any{})
	if imported.Code != http.StatusCreated {
		t.Fatalf("import = %d: %s", imported.Code, imported.Body.String())
	}
}

func TestImportPeopleFromTemplate(t *testing.T) {
	srv := seededServer(t)
	created := postJSON(t, srv, "/api/v1/tournaments", map[string]any{
		"name": "People Event", "saveTemplate": true, "templateName": "People Template",
		"disciplines": []map[string]any{{"name": "Randori"}},
		"categories":  []map[string]any{{"name": "Open", "kind": "combined"}},
		"clubs":       []map[string]any{{"name": "New Club", "colour": "#ff0000"}},
		"fighters": []map[string]any{{
			"name": "New Fighter", "club": "New Club", "weightKg": 70, "dob": "1999-01-01", "gender": "male",
		}},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", created.Code, created.Body.String())
	}
	var result struct {
		TemplateID string `json:"templateId"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	imported := postJSON(t, srv, "/api/v1/templates/"+result.TemplateID+"/import-people", map[string]any{})
	if imported.Code != http.StatusOK {
		t.Fatalf("import-people = %d: %s", imported.Code, imported.Body.String())
	}
}

func TestCreateTournamentValidation(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/tournaments", map[string]any{"name": "No disciplines"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

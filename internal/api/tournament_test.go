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
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.EventID == "" || created.TemplateID == "" || created.DisciplineCount != 1 || created.InviteCount != 1 {
		t.Fatalf("unexpected result: %+v", created)
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

func TestCreateTournamentValidation(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/tournaments", map[string]any{"name": "No disciplines"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCapacityRejectsFullCategory(t *testing.T) {
	srv := seededServer(t)
	if _, err := srv.Store.Exec(`UPDATE format_policies SET config_json = '{"cap":2}' WHERE id = 'fp-demo'`); err != nil {
		t.Fatal(err)
	}
	preview := postJSON(t, srv, "/api/v1/signup/preview", map[string]any{
		"dateOfBirth": "1998-04-12", "unitSystem": "metric", "heightCm": 175,
		"weightKg": 74, "gender": "male",
	})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview = %d", preview.Code)
	}
	var body struct {
		Eligible []struct {
			CategoryID string `json:"categoryId"`
			Full       bool   `json:"full"`
			Spaces     int    `json:"spaces"`
		} `json:"eligible"`
	}
	if err := json.Unmarshal(preview.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range body.Eligible {
		if item.CategoryID == "cat-adult-76" && item.Full && item.Spaces == 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("full category not flagged: %+v", body.Eligible)
	}

	signup := postJSON(t, srv, "/api/v1/signup", map[string]any{
		"name": "Late Entrant", "email": "late@demo.j3", "dateOfBirth": "1998-04-12",
		"unitSystem": "metric", "heightCm": 175, "weightKg": 74, "gender": "male",
		"accountType": "fighter", "categoryIds": []string{"cat-adult-76"},
	})
	if signup.Code != http.StatusBadRequest {
		t.Fatalf("signup into full category = %d, want 400: %s", signup.Code, signup.Body.String())
	}
}

func TestForfeitRecordsResultAndAudit(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/demo/fights/fight-01/forfeit", map[string]any{
		"reason": "injury", "winner": "red",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var result struct{ State, Outcome string }
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.State != "forfeit" || result.Outcome != "win_red" {
		t.Fatalf("unexpected result %+v", result)
	}
	audit := httptest.NewRecorder()
	srv.Handler().ServeHTTP(audit, httptest.NewRequest(http.MethodGet,
		"/api/v1/demo/audit?action=fight.forfeit", nil))
	var body struct {
		Entries []any `json:"entries"`
	}
	if err := json.Unmarshal(audit.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Entries) == 0 {
		t.Fatal("no forfeit audit entry")
	}
}

func TestWithdrawnEntrantExcludedFromPools(t *testing.T) {
	srv := seededServer(t)
	status := postJSON(t, srv, "/api/v1/demo/entrants/ent-01/status", map[string]any{"status": "withdrawn"})
	if status.Code != http.StatusOK {
		t.Fatalf("status = %d", status.Code)
	}
	generated := postJSON(t, srv, "/api/v1/demo/pools/generate", map[string]any{
		"eventId": "ev-demo", "categoryId": "cat-adult-76", "targetPoolSize": 4,
		"sameClubSeparation": true, "seed": 9,
	})
	if generated.Code != http.StatusOK {
		t.Fatalf("generate = %d", generated.Code)
	}
	var body struct {
		Pools []struct {
			Members []struct {
				ID string `json:"id"`
			} `json:"members"`
		} `json:"pools"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, pool := range body.Pools {
		for _, member := range pool.Members {
			if member.ID == "f01" {
				t.Fatal("withdrawn fighter still placed in a pool")
			}
		}
	}
}

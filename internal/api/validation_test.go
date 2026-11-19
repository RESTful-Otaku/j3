package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSignupRejectsInvalidData(t *testing.T) {
	srv := seededServer(t)
	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing name", map[string]any{"email": "a@b.com", "dateOfBirth": "1990-01-01", "unitSystem": "metric", "heightCm": 175, "weightKg": 70}},
		{"bad email", map[string]any{"name": "A", "email": "not-an-email", "dateOfBirth": "1990-01-01", "unitSystem": "metric", "heightCm": 175, "weightKg": 70}},
		{"bad date", map[string]any{"name": "A", "email": "a@b.com", "dateOfBirth": "yesterday", "unitSystem": "metric", "heightCm": 175, "weightKg": 70}},
		{"short height", map[string]any{"name": "A", "email": "a@b.com", "dateOfBirth": "1990-01-01", "unitSystem": "metric", "heightCm": 20, "weightKg": 70}},
		{"heavy weight", map[string]any{"name": "A", "email": "a@b.com", "dateOfBirth": "1990-01-01", "unitSystem": "metric", "heightCm": 175, "weightKg": 900}},
		{"unknown unit", map[string]any{"name": "A", "email": "a@b.com", "dateOfBirth": "1990-01-01", "unitSystem": "parsec", "heightCm": 175, "weightKg": 70}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := postJSON(t, srv, "/api/v1/signup", testCase.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestQueryAllowsSearchSortAndFilter(t *testing.T) {
	srv := seededServer(t)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet,
		"/api/v1/query?table=fighters&search=mori&sort=display_name&order=asc&limit=10", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Rows) == 0 {
		t.Fatal("expected search results")
	}

	filtered := httptest.NewRecorder()
	srv.Handler().ServeHTTP(filtered, httptest.NewRequest(http.MethodGet,
		"/api/v1/query?table=users&role=official&limit=50", nil))
	if filtered.Code != http.StatusOK {
		t.Fatalf("filter status = %d", filtered.Code)
	}
	var filteredBody struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(filtered.Body.Bytes(), &filteredBody); err != nil {
		t.Fatal(err)
	}
	if len(filteredBody.Rows) != 8 {
		t.Fatalf("officials = %d, want 8", len(filteredBody.Rows))
	}
}

func TestQueryRejectsUnknownTables(t *testing.T) {
	srv := seededServer(t)
	for _, table := range []string{"fights; DROP TABLE fighters", "audit_logs", "", "sqlite_master"} {
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet,
			"/api/v1/query?table="+url.QueryEscape(table), nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("table %q: status = %d, want 400", table, rec.Code)
		}
	}
}

func TestGeneratePoolsWithNoEntrants(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/demo/pools/generate", map[string]any{
		"eventId": "ev-demo", "categoryId": "cat-does-not-exist", "targetPoolSize": 4,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Pools        []any `json:"pools"`
		EntrantCount int   `json:"entrantCount"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Pools) != 0 || body.EntrantCount != 0 {
		t.Fatalf("expected no pools, got %+v", body)
	}
}

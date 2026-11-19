package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
)

func postJSON(t *testing.T, srv *api.Server, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestSignupConvertsImperialToMetric(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/signup", map[string]any{
		"name": "Test Fighter", "email": "test-fighter@demo.j3", "phone": "+44 7700 900000",
		"dateOfBirth": "1997-05-01", "unitSystem": "imperial",
		"heightFt": 5, "heightIn": 9, "weightLb": 165,
		"beltRankId": "br-blue", "gender": "male", "accountType": "fighter",
		"clubId": "club-riverside", "categoryIds": []string{"cat-adult-76"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		HeightCm float64 `json:"heightCm"`
		WeightKg float64 `json:"weightKg"`
		Status   string  `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.HeightCm < 175.2 || result.HeightCm > 175.3 {
		t.Errorf("height = %v, want ~175.26", result.HeightCm)
	}
	if result.WeightKg < 74.8 || result.WeightKg > 74.9 {
		t.Errorf("weight = %v, want ~74.84", result.WeightKg)
	}
	if result.Status != "active" {
		t.Errorf("status = %s", result.Status)
	}
}

func TestSignupPrivilegedRoleNeedsApproval(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/signup", map[string]any{
		"name": "Request Official", "email": "req-official@demo.j3",
		"dateOfBirth": "1990-01-01", "unitSystem": "metric",
		"heightCm": 178, "weightKg": 80, "accountType": "official",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var result struct {
		Status        string `json:"status"`
		RoleRequestID string `json:"roleRequestId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "invited" || result.RoleRequestID == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
	approve := httptest.NewRecorder()
	srv.Handler().ServeHTTP(approve, httptest.NewRequest(http.MethodPost,
		"/api/v1/demo/role-requests/"+result.RoleRequestID+"/approve", nil))
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status = %d: %s", approve.Code, approve.Body.String())
	}
}

func TestSignupPreviewOffersCompatibleCategories(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/signup/preview", map[string]any{
		"dateOfBirth": "1998-04-12", "unitSystem": "metric",
		"heightCm": 175, "weightKg": 74, "gender": "male",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Eligible []struct {
			CategoryID string `json:"categoryId"`
			Strict     bool   `json:"strict"`
		} `json:"eligible"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range body.Eligible {
		if e.CategoryID == "cat-adult-76" && e.Strict {
			found = true
		}
	}
	if !found {
		t.Fatal("compatible category not offered")
	}
}

func TestGenerateAndListPools(t *testing.T) {
	srv := seededServer(t)
	generated := postJSON(t, srv, "/api/v1/demo/pools/generate", map[string]any{
		"eventId": "ev-demo", "categoryId": "cat-adult-76",
		"targetPoolSize": 3, "maxPools": 6, "sameClubSeparation": true, "seed": 7,
	})
	if generated.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", generated.Code, generated.Body.String())
	}
	var body struct {
		Pools []struct {
			Name    string `json:"name"`
			Members []struct {
				Name string `json:"name"`
			} `json:"members"`
		} `json:"pools"`
		EntrantCount int `json:"entrantCount"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Pools) == 0 || body.EntrantCount == 0 {
		t.Fatal("expected generated pools")
	}
	total := 0
	for _, p := range body.Pools {
		total += len(p.Members)
	}
	if total != body.EntrantCount {
		t.Fatalf("placed %d of %d entrants", total, body.EntrantCount)
	}

	listed := httptest.NewRecorder()
	srv.Handler().ServeHTTP(listed, httptest.NewRequest(http.MethodGet, "/api/v1/demo/pools?eventId=ev-demo", nil))
	if listed.Code != http.StatusOK {
		t.Fatalf("list status = %d", listed.Code)
	}
}

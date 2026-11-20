package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestReadinessAndAutoRun(t *testing.T) {
	srv := seededServer(t)
	rec := httptestRecorder(srv, http.MethodGet, "/api/v1/demo/readiness")
	if rec.Code != http.StatusOK {
		t.Fatalf("readiness = %d", rec.Code)
	}
	var body struct {
		Categories []struct {
			Name     string `json:"name"`
			Entrants int    `json:"entrants"`
			Ready    bool   `json:"ready"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Categories) == 0 {
		t.Fatal("no readiness data")
	}
	ready := 0
	for _, item := range body.Categories {
		if item.Ready {
			ready++
		}
	}
	if ready == 0 {
		t.Fatal("no category ready for bracket generation")
	}

	auto := postJSON(t, srv, "/api/v1/demo/auto-run", map[string]any{})
	if auto.Code != http.StatusOK {
		t.Fatalf("auto-run = %d: %s", auto.Code, auto.Body.String())
	}
	var result struct {
		Pools     int `json:"pools"`
		Stages    int `json:"stages"`
		Officials int `json:"officials"`
	}
	if err := json.Unmarshal(auto.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Pools == 0 || result.Stages == 0 || result.Officials == 0 {
		t.Fatalf("auto-run incomplete: %+v", result)
	}
}

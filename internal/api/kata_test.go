package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestKataPanelScoringDropsHighAndLow(t *testing.T) {
	srv := seededServer(t)
	created := postJSON(t, srv, "/api/v1/demo/kata/performances", map[string]any{
		"fighterId": "f05", "kataName": "Seiyunchin", "panelSize": 5,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", created.Code, created.Body.String())
	}
	var body struct {
		Performances []struct {
			ID string `json:"id"`
		} `json:"performances"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	id := body.Performances[0].ID

	scores := []float64{5, 8, 9, 10, 3}
	for i, score := range scores {
		rec := postJSON(t, srv, "/api/v1/demo/kata/"+id+"/scores", map[string]any{
			"judge": i + 1, "score": score,
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("score %d = %d: %s", i, rec.Code, rec.Body.String())
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
	}
	listed := httptestRecorder(srv, http.MethodGet, "/api/v1/demo/kata")
	var result struct {
		Performances []struct {
			ID     string  `json:"id"`
			Status string  `json:"status"`
			Total  float64 `json:"total"`
		} `json:"performances"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var total float64
	for _, item := range result.Performances {
		if item.ID == id {
			total = item.Total
			if item.Status != "complete" {
				t.Fatal("performance not complete")
			}
		}
	}
	// Drop 3 and 10, average 5+8+9 = 7.33.
	if total < 7.3 || total > 7.4 {
		t.Fatalf("total = %v, want ~7.33", total)
	}
}

func TestKataScoreValidation(t *testing.T) {
	srv := seededServer(t)
	created := postJSON(t, srv, "/api/v1/demo/kata/performances", map[string]any{"fighterId": "f06", "panelSize": 3})
	var body struct {
		Performances []struct {
			ID string `json:"id"`
		} `json:"performances"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	id := body.Performances[0].ID
	if rec := postJSON(t, srv, "/api/v1/demo/kata/"+id+"/scores", map[string]any{"judge": 9, "score": 7}); rec.Code != http.StatusBadRequest {
		t.Fatalf("out-of-range judge = %d", rec.Code)
	}
	if rec := postJSON(t, srv, "/api/v1/demo/kata/"+id+"/scores", map[string]any{"judge": 1, "score": 12}); rec.Code != http.StatusBadRequest {
		t.Fatalf("out-of-range score = %d", rec.Code)
	}
}

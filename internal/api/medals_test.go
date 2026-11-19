package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAwardMedalsIncludingDualBronze(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/demo/medals/award", map[string]any{
		"gold": "f01", "silver": "f02", "bronze": []string{"f03", "f04"},
		"category": "cat-adult-76",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("award = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Medals []struct {
			Kind string `json:"kind"`
		} `json:"medals"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	bronze := 0
	for _, medal := range body.Medals {
		if medal.Kind == "bronze" {
			bronze++
		}
	}
	if bronze != 2 {
		t.Fatalf("bronze count = %d, want 2", bronze)
	}

	empty := postJSON(t, srv, "/api/v1/demo/medals/award", map[string]any{})
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("empty award = %d", empty.Code)
	}
}

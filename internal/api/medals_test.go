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
		"category": "cat-adult-76", "discipline": "disc-fight", "eventType": "fight",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("award = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Medals []struct {
			Kind         string `json:"kind"`
			Category     string `json:"category"`
			DisciplineID string `json:"disciplineId"`
			Discipline   string `json:"discipline"`
			EventType    string `json:"eventType"`
			AwardKey     string `json:"awardKey"`
			Source       string `json:"source"`
		} `json:"medals"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	bronze := 0
	labelled := false
	for _, medal := range body.Medals {
		if medal.Kind == "bronze" {
			bronze++
		}
		if medal.Source == "manual" && medal.DisciplineID == "disc-fight" && medal.EventType == "fight" && medal.AwardKey != "" {
			labelled = true
		}
	}
	if bronze < 2 || !labelled {
		t.Fatalf("expected dual bronze with category/discipline context, got bronze=%d labelled=%v", bronze, labelled)
	}

	empty := postJSON(t, srv, "/api/v1/demo/medals/award", map[string]any{})
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("empty award = %d", empty.Code)
	}
}

func TestRecomputePreservesManualAwardsAndClearsStaleAuto(t *testing.T) {
	srv := seededServer(t)
	awarded := postJSON(t, srv, "/api/v1/demo/medals/award", map[string]any{
		"gold": "f01", "category": "cat-adult-76", "discipline": "disc-fight", "eventType": "fight",
	})
	if awarded.Code != http.StatusOK {
		t.Fatalf("manual award = %d: %s", awarded.Code, awarded.Body.String())
	}

	var eventID string
	if err := srv.Store.DB.QueryRow(`SELECT id FROM events LIMIT 1`).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := srv.Store.DB.Exec(`
		INSERT INTO medals (id, event_id, fighter_id, kind, event_type, award_key, source)
		VALUES (?, ?, ?, 'gold', 'fight', ?, 'auto')`,
		"medal-stale-auto", eventID, "f02", "stale|auto|key"); err != nil {
		t.Fatal(err)
	}

	rec := postJSON(t, srv, "/api/v1/demo/medals/recompute", map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("recompute = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Medals []struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		} `json:"medals"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	manual := false
	for _, medal := range body.Medals {
		if medal.ID == "medal-stale-auto" {
			t.Fatal("stale derived medal survived recompute")
		}
		if medal.Source == "manual" {
			manual = true
		}
	}
	if !manual {
		t.Fatal("manual medal override was lost during recompute")
	}
}

func TestRecomputeMedalsIsIdempotentAndFull(t *testing.T) {
	srv := seededServer(t)
	var first, second struct {
		Medals []struct {
			ID           string `json:"id"`
			CategoryID   string `json:"categoryId"`
			DisciplineID string `json:"disciplineId"`
			EventType    string `json:"eventType"`
			Source       string `json:"source"`
		} `json:"medals"`
	}
	for i, target := range []*struct {
		Medals []struct {
			ID           string `json:"id"`
			CategoryID   string `json:"categoryId"`
			DisciplineID string `json:"disciplineId"`
			EventType    string `json:"eventType"`
			Source       string `json:"source"`
		} `json:"medals"`
	}{&first, &second} {
		rec := postJSON(t, srv, "/api/v1/demo/medals/recompute", map[string]any{})
		if rec.Code != http.StatusOK {
			t.Fatalf("recompute %d = %d: %s", i, rec.Code, rec.Body.String())
		}
		if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
			t.Fatal(err)
		}
	}
	if len(first.Medals) < 20 || len(first.Medals) != len(second.Medals) {
		t.Fatalf("expected a full idempotent medal set, got %d then %d", len(first.Medals), len(second.Medals))
	}
	groups := map[string]bool{}
	kata := false
	for _, medal := range first.Medals {
		groups[medal.CategoryID+"|"+medal.DisciplineID+"|"+medal.EventType] = true
		if medal.EventType == "kata" {
			kata = true
		}
	}
	if len(groups) < 3 || !kata {
		t.Fatalf("expected multiple divisions including kata, got groups=%d kata=%v", len(groups), kata)
	}
}

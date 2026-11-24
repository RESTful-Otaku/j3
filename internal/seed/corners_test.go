package seed_test

import (
	"encoding/json"
	"testing"

	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func TestSeededEventsDefaultToRedAndWhiteCorners(t *testing.T) {
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureOkami(store); err != nil {
		t.Fatal(err)
	}

	type cornerConfig struct {
		CornerColours map[string]string `json:"cornerColours"`
		CornerLabels  map[string]string `json:"cornerLabels"`
	}
	for _, eventID := range []string{"ev-demo", "ev-okami"} {
		var configJSON string
		if err := store.QueryRow(`SELECT rs.config_json FROM events e
			JOIN rule_sets rs ON rs.id = e.rule_set_id WHERE e.id = ?`, eventID).Scan(&configJSON); err != nil {
			t.Fatalf("%s rule set: %v", eventID, err)
		}
		var config cornerConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			t.Fatalf("%s config: %v", eventID, err)
		}
		if config.CornerColours["red"] != "#b91c1c" || config.CornerColours["blue"] != "#f8fafc" {
			t.Fatalf("%s colours = %+v", eventID, config.CornerColours)
		}
		if config.CornerLabels["red"] != "RED" || config.CornerLabels["blue"] != "WHITE" {
			t.Fatalf("%s labels = %+v", eventID, config.CornerLabels)
		}
	}
}

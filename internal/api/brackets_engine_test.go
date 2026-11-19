package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestBracketBuildAndPromote(t *testing.T) {
	srv := seededServer(t)
	built := postJSON(t, srv, "/api/v1/demo/brackets/build", map[string]any{
		"eventId": "ev-demo", "categoryId": "cat-adult-76",
		"format": "Double elimination", "targetPoolSize": 4, "maxPools": 8,
		"seed": 7, "thirdPlace": true,
	})
	if built.Code != http.StatusCreated {
		t.Fatalf("build = %d: %s", built.Code, built.Body.String())
	}
	var body struct {
		Format string `json:"format"`
		Stages []struct {
			Name  string `json:"name"`
			Kind  string `json:"kind"`
			Slots []struct {
				ID         string `json:"id"`
				FighterID  string `json:"fighterId"`
				State      string `json:"state"`
				ResultIcon string `json:"resultIcon"`
			} `json:"slots"`
		} `json:"stages"`
	}
	if err := json.Unmarshal(built.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Stages) < 2 {
		t.Fatalf("stages = %d", len(body.Stages))
	}

	var firstSlot string
	var hadFighter bool
	for _, stage := range body.Stages {
		for _, slot := range stage.Slots {
			if slot.FighterID != "" {
				firstSlot, hadFighter = slot.ID, true
				break
			}
		}
		if hadFighter {
			break
		}
	}
	if !hadFighter {
		t.Fatal("no fighter placed in any bracket slot")
	}

	promoted := postJSON(t, srv, "/api/v1/demo/brackets/slots/"+firstSlot+"/promote", map[string]any{
		"winner": true, "resultIcon": "WS",
	})
	if promoted.Code != http.StatusOK {
		t.Fatalf("promote = %d: %s", promoted.Code, promoted.Body.String())
	}
	var after struct {
		Stages []struct {
			Slots []struct {
				ID         string `json:"id"`
				FighterID  string `json:"fighterId"`
				State      string `json:"state"`
				ResultIcon string `json:"resultIcon"`
			} `json:"slots"`
		} `json:"stages"`
	}
	if err := json.Unmarshal(promoted.Body.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, stage := range after.Stages {
		for _, slot := range stage.Slots {
			if slot.ID == firstSlot && slot.State == "advanced" && slot.ResultIcon == "WS" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("promotion result not retained in history")
	}
}

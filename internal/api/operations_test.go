package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCheckInAndTeamRankings(t *testing.T) {
	srv := seededServer(t)
	rec := postJSON(t, srv, "/api/v1/demo/entrants/ent-02/checkin", map[string]any{"weightKg": 200})
	if rec.Code != http.StatusOK {
		t.Fatalf("check-in = %d: %s", rec.Code, rec.Body.String())
	}
	var checked struct {
		OutOfClass bool `json:"outOfClass"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &checked); err != nil {
		t.Fatal(err)
	}
	if !checked.OutOfClass {
		t.Fatal("out-of-class competitor not flagged")
	}

	postJSON(t, srv, "/api/v1/demo/medals/award", map[string]any{
		"gold": "f01", "silver": "f02", "bronze": []string{"f03"},
	})
	teams := httptestRecorder(srv, http.MethodGet, "/api/v1/demo/team-rankings")
	if teams.Code != http.StatusOK {
		t.Fatalf("team rankings = %d", teams.Code)
	}
	var body struct {
		Teams []struct {
			Club   string `json:"club"`
			Points int    `json:"points"`
		} `json:"teams"`
	}
	if err := json.Unmarshal(teams.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Teams) == 0 {
		t.Fatal("no team rankings")
	}
}

func TestCSVExportAndImport(t *testing.T) {
	srv := seededServer(t)
	export := httptestRecorder(srv, http.MethodGet, "/api/v1/demo/export?type=fights")
	if export.Code != http.StatusOK {
		t.Fatalf("export = %d", export.Code)
	}
	if export.Header().Get("Content-Type") != "text/csv; charset=utf-8" {
		t.Fatal("missing CSV content type")
	}
	if export.Body.Len() < 20 {
		t.Fatal("CSV export is empty")
	}

	imported := postJSON(t, srv, "/api/v1/demo/import", map[string]any{
		"rows": []map[string]any{{"name": "Imported Fighter", "club": "New Club", "weightKg": 72, "gender": "male"}},
	})
	if imported.Code != http.StatusCreated {
		t.Fatalf("import = %d: %s", imported.Code, imported.Body.String())
	}
}

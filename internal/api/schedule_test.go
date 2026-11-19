package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
)

func httptestRecorder(srv *api.Server, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestScheduleDriftAndCompletion(t *testing.T) {
	srv := seededServer(t)
	rec := httptestRecorder(srv, http.MethodGet, "/api/v1/demo/schedule?eventId=ev-demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d", rec.Code)
	}
	var body struct {
		Items []struct {
			ID             string `json:"id"`
			ProjectedStart string `json:"projectedStart"`
			OffsetMinutes  int    `json:"offsetMinutes"`
			Status         string `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) == 0 {
		t.Fatal("no schedule items")
	}
	original := body.Items[0].ProjectedStart

	drifted := postJSON(t, srv, "/api/v1/demo/schedule/drift", map[string]any{"minutes": 15})
	if drifted.Code != http.StatusOK {
		t.Fatalf("drift = %d", drifted.Code)
	}
	if err := json.Unmarshal(drifted.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Items[0].OffsetMinutes != 15 || body.Items[0].ProjectedStart == original {
		t.Fatalf("drift not applied: %+v", body.Items[0])
	}

	completed := postJSON(t, srv, "/api/v1/demo/schedule/"+body.Items[0].ID+"/complete", map[string]any{})
	if completed.Code != http.StatusOK {
		t.Fatalf("complete = %d", completed.Code)
	}
	if err := json.Unmarshal(completed.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Items[0].Status != "complete" {
		t.Fatal("item not marked complete")
	}
}

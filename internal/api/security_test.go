package api_test

import (
	"net/http"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func openUnauthenticatedServer(t *testing.T) *api.Server {
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	return api.New(store, "test")
}

func TestMutatingRoutesRequireAuthentication(t *testing.T) {
	srv := openUnauthenticatedServer(t)
	mutations := []struct {
		method, path, body string
	}{
		{"POST", "/api/v1/demo/brackets/build", `{}`},
		{"POST", "/api/v1/demo/pools/generate", `{}`},
		{"POST", "/api/v1/demo/auto-run", `{}`},
		{"POST", "/api/v1/demo/announcements", `{"title":"x"}`},
		{"POST", "/api/v1/demo/admin/users/u1/status", `{"status":"locked"}`},
		{"POST", "/api/v1/demo/permissions", `{"role":"fighter","resource":"fights","flag":"delete","value":true}`},
		{"POST", "/api/v1/fights/fight-01/start", `{}`},
		{"POST", "/api/v1/uploads", `{}`},
	}
	for _, mutation := range mutations {
		rec := httptestRecorder(srv, mutation.method, mutation.path)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without token = %d, want 401", mutation.method, mutation.path, rec.Code)
		}
	}
}

func TestPublicReadsRemainOpen(t *testing.T) {
	srv := openUnauthenticatedServer(t)
	for _, path := range []string{"/api/v1/demo", "/api/v1/demo/fights", "/api/v1/demo/standings", "/api/v1/demo/belts"} {
		rec := httptestRecorder(srv, http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
		}
	}
}

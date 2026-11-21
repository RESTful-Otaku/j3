package api_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestOpenAPIRouteParity guards against undocumented endpoints. Every route
// registered in server.go must appear as a path and method in the OpenAPI spec.
func TestOpenAPIRouteParity(t *testing.T) {
	server, err := os.ReadFile("../../internal/api/server.go")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	specText := string(spec)
	pattern := regexp.MustCompile(`mux\.HandleFunc\("(GET|POST|PATCH|DELETE) ([^"]+)"`)
	routes := pattern.FindAllStringSubmatch(string(server), -1)
	if len(routes) < 60 {
		t.Fatalf("only found %d registered routes", len(routes))
	}
	for _, route := range routes {
		method, path := strings.ToLower(route[1]), route[2]
		pathBlock := "  " + path + ":\n"
		if !strings.Contains(specText, pathBlock) {
			t.Errorf("path %s is not documented", path)
			continue
		}
		start := strings.Index(specText, pathBlock)
		tail := specText[start:]
		if end := strings.Index(tail, "\n  /"); end > 0 {
			tail = tail[:end]
		}
		if !strings.Contains(tail, "    "+method+":") {
			t.Errorf("%s %s is not documented", strings.ToUpper(method), path)
		}
	}
}

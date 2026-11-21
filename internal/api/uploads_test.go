package api_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/restful-otaku/j3/internal/api"
)

func uploadRequest(t *testing.T, srv *api.Server, token, kind string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("kind", kind)
	part, err := writer.CreateFormFile("file", "brand.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestUploadColoursAndRoleLimits(t *testing.T) {
	srv := seededAuthServer(t)
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 220, G: 30, B: 40, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}

	denied := uploadRequest(t, srv,
		sessionToken(t, srv, "u-fighter-01", "fighter"), "banner", encoded.Bytes())
	if denied.Code != http.StatusForbidden {
		t.Fatalf("fighter banner status = %d, want 403", denied.Code)
	}
	allowed := uploadRequest(t, srv,
		sessionToken(t, srv, "u-leader", "club_leader"), "logo", encoded.Bytes())
	if allowed.Code != http.StatusCreated {
		t.Fatalf("club logo status = %d: %s", allowed.Code, allowed.Body.String())
	}
	var result struct {
		URL       string `json:"url"`
		Colour    string `json:"colour"`
		Highlight string `json:"highlight"`
	}
	if err := json.Unmarshal(allowed.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.URL == "" || len(result.Colour) != 7 || len(result.Highlight) != 7 {
		t.Fatalf("unexpected upload result: %+v", result)
	}
}

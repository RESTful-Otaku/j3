package api

import (
	"bytes"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/restful-otaku/j3/internal/auth"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var allowedUploadKinds = map[string]bool{
	"avatar": true, "photo": true, "club": true, "logo": true,
	"banner": true,
}

// Upload rights follow the RBAC hierarchy: fighters and officials may upload a
// photo; clubs and runners may upload a logo and banner.
var uploadRights = map[string]map[string]bool{
	"fighter":           {"photo": true, "avatar": true},
	"official":          {"photo": true, "avatar": true},
	"club_leader":       {"club": true, "logo": true, "banner": true},
	"tournament_runner": {"logo": true, "banner": true},
}

var allowedExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true,
}

const maxUploadBytes = 5 << 20

// upload stores media locally with the venue data so white-label branding,
// fighter photos, club symbols and banners survive offline events.
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "settings", "create") {
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeBadRequest(w, "invalid upload")
		return
	}
	kind := r.FormValue("kind")
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	role := string(identity.Role)
	if !allowedUploadKinds[kind] {
		writeBadRequest(w, "unsupported upload kind")
		return
	}
	if allowed, known := uploadRights[role]; !known || !allowed[kind] {
		writeBadRequest(w, "your account type cannot upload this kind of image")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeBadRequest(w, "file is required")
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		writeBadRequest(w, "use PNG, JPG or WEBP images")
		return
	}

	dir := filepath.Join("data", "uploads", kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, err)
		return
	}
	name := newID(kind) + ext
	path := filepath.Join(dir, name)
	buffer, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		writeError(w, err)
		return
	}
	if len(buffer) > maxUploadBytes {
		writeBadRequest(w, "files must be under 5 MB")
		return
	}
	if err := os.WriteFile(path, buffer, 0o644); err != nil {
		writeError(w, err)
		return
	}

	main, highlight := dominantColours(buffer)
	url := fmt.Sprintf("/uploads/%s/%s", kind, name)
	writeJSON(w, http.StatusCreated, map[string]any{
		"url": url, "path": url, "kind": kind,
		"colour": main, "highlight": highlight,
	})
}

// dominantColours samples the decoded image and returns a representative main
// colour plus a lighter highlight. Unsupported formats return neutral defaults.
func dominantColours(data []byte) (string, string) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "#2563eb", "#93c5fd"
	}
	bounds := img.Bounds()
	buckets := map[[3]int]int{}
	stepX := (bounds.Dx() / 40) + 1
	stepY := (bounds.Dy() / 40) + 1
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, a := img.At(x, y).RGBA()
			if a < 128 {
				continue
			}
			key := [3]int{int(r>>8) / 24, int(g>>8) / 24, int(b>>8) / 24}
			buckets[key]++
		}
	}
	best := [3]int{2, 3, 5}
	bestCount := 0
	for key, count := range buckets {
		if count > bestCount {
			best, bestCount = key, count
		}
	}
	base := color.RGBA{
		R: uint8(best[0]*24 + 12), G: uint8(best[1]*24 + 12), B: uint8(best[2]*24 + 12), A: 255,
	}
	highlight := color.RGBA{
		R: lighten(base.R), G: lighten(base.G), B: lighten(base.B), A: 255,
	}
	return fmt.Sprintf("#%02x%02x%02x", base.R, base.G, base.B),
		fmt.Sprintf("#%02x%02x%02x", highlight.R, highlight.G, highlight.B)
}

func lighten(value uint8) uint8 {
	next := int(value) + (255-int(value))*45/100
	if next > 255 {
		return 255
	}
	return uint8(next)
}

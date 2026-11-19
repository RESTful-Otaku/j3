package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var allowedUploadKinds = map[string]bool{
	"avatar": true, "photo": true, "club": true, "logo": true,
	"banner": true, "belt": true,
}

var allowedExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".svg": true,
}

// upload stores media locally with the venue data so white-label branding,
// fighter photos, club symbols and banners survive offline events.
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeBadRequest(w, "invalid upload")
		return
	}
	kind := r.FormValue("kind")
	if !allowedUploadKinds[kind] {
		writeBadRequest(w, "unsupported upload kind")
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
		writeBadRequest(w, "use PNG, JPG, WEBP or SVG images")
		return
	}
	dir := filepath.Join("data", "uploads", kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, err)
		return
	}
	name := newID(kind) + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		writeError(w, err)
		return
	}
	defer dst.Close()
	written, err := io.Copy(dst, io.LimitReader(file, 5<<20+1))
	if err != nil {
		writeError(w, err)
		return
	}
	if written > 5<<20 {
		os.Remove(filepath.Join(dir, name))
		writeBadRequest(w, "files must be under 5 MB")
		return
	}
	url := fmt.Sprintf("/uploads/%s/%s", kind, name)
	writeJSON(w, http.StatusCreated, map[string]any{"url": url, "path": url, "kind": kind})
}

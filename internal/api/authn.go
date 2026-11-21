package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/auth"
)

type authChallenge struct {
	Email   string
	Method  string
	Code    string
	Expires time.Time
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Provider string `json:"provider"`
}

var challenges = struct {
	mu sync.Mutex
	m  map[string]authChallenge
}{m: map[string]authChallenge{}}

func randomCode() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	code := fmt.Sprintf("%06d", (uint32(buf[0])<<24|uint32(buf[1])<<16|uint32(buf[2])<<8|uint32(buf[3]))%1000000)
	return code
}

// startLogin verifies credentials where a password is supplied, then issues an
// MFA challenge. External providers require their own credentials; in demo
// mode the code is returned so the full flow can be QA'd locally.
func (s *Server) startLogin(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	provider := strings.ToLower(body.Provider)
	if provider == "" {
		provider = "password"
	}
	if provider != "password" && provider != "google" && provider != "apple" && provider != "passkey" {
		writeBadRequest(w, "unknown sign-in provider")
		return
	}
	if body.Email == "" || !strings.Contains(body.Email, "@") {
		writeBadRequest(w, "a valid email is required")
		return
	}
	if provider == "password" {
		var hash string
		var status string
		if err := s.Store.QueryRowContext(r.Context(),
			`SELECT password_hash, status FROM users WHERE email = ?`, body.Email).Scan(&hash, &status); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		if status != "active" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "account is not active"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
	}
	if provider != "password" && !demoMode() {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "external sign-in is not configured"})
		return
	}

	code := randomCode()
	challenges.mu.Lock()
	challenges.m[body.Email] = authChallenge{Email: body.Email, Method: provider, Code: code, Expires: time.Now().Add(5 * time.Minute)}
	challenges.mu.Unlock()
	response := map[string]any{
		"challengeId": body.Email, "mfaRequired": true, "provider": provider,
	}
	if demoMode() {
		response["demoCode"] = code
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) verifyLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	challenges.mu.Lock()
	challenge, ok := challenges.m[body.Email]
	challenges.mu.Unlock()
	if !ok || time.Now().After(challenge.Expires) || challenge.Code != body.Code {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired code"})
		return
	}
	challenges.mu.Lock()
	delete(challenges.m, body.Email)
	challenges.mu.Unlock()

	var userID, role, status string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT id, role, status FROM users WHERE email = ?`, body.Email).
		Scan(&userID, &role, &status); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unknown account"})
		return
	}
	if status != "active" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "account is not active"})
		return
	}
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO sessions (token, user_id, role, expires_at)
		VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now','+12 hours'))`, token, userID, role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "email": body.Email,
		"role": auth.CanonicalRole(role), "expiresAt": time.Now().Add(12 * time.Hour).Format(time.RFC3339),
	})
}

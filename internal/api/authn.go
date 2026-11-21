package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
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

var mailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

var challenges = struct {
	mu sync.Mutex
	m  map[string]authChallenge
}{m: map[string]authChallenge{}}

func randomCode() string {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", value.Int64())
}

var mfaAttempts = struct {
	mu sync.Mutex
	m  map[string]*mfaAttempt
}{m: map[string]*mfaAttempt{}}

type mfaAttempt struct {
	count       int
	lockedUntil time.Time
}

func mfaAllowed(email string) bool {
	mfaAttempts.mu.Lock()
	defer mfaAttempts.mu.Unlock()
	attempt, ok := mfaAttempts.m[email]
	if !ok {
		return true
	}
	if time.Now().Before(attempt.lockedUntil) {
		return false
	}
	if time.Now().After(attempt.lockedUntil) && !attempt.lockedUntil.IsZero() {
		delete(mfaAttempts.m, email)
	}
	return true
}

func recordMFAFailure(email string) {
	mfaAttempts.mu.Lock()
	defer mfaAttempts.mu.Unlock()
	attempt := mfaAttempts.m[email]
	if attempt == nil {
		attempt = &mfaAttempt{}
		mfaAttempts.m[email] = attempt
	}
	attempt.count++
	if attempt.count >= 5 {
		attempt.lockedUntil = time.Now().Add(10 * time.Minute)
		attempt.count = 0
	}
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
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	if !mailPattern.MatchString(body.Email) {
		writeBadRequest(w, "a valid email is required")
		return
	}
	if provider == "password" {
		var hash string
		var status string
		if err := s.Store.QueryRowContext(r.Context(),
			`SELECT password_hash, status FROM users WHERE lower(email) = ?`, body.Email).Scan(&hash, &status); err != nil {
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
	email := strings.ToLower(strings.TrimSpace(body.Email))
	if !mfaAllowed(email) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts; try again later"})
		return
	}
	challenges.mu.Lock()
	challenge, ok := challenges.m[email]
	challenges.mu.Unlock()
	if !ok || time.Now().After(challenge.Expires) ||
		subtle.ConstantTimeCompare([]byte(challenge.Code), []byte(body.Code)) != 1 {
		recordMFAFailure(email)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired code"})
		return
	}
	challenges.mu.Lock()
	delete(challenges.m, email)
	mfaAttempts.mu.Lock()
	delete(mfaAttempts.m, email)
	mfaAttempts.mu.Unlock()
	challenges.mu.Unlock()

	var userID, role, status string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT id, role, status FROM users WHERE lower(email) = ?`, email).
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

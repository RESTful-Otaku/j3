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
	var mustChangePassword bool
	if provider == "password" {
		var hash string
		var status string
		var mustChange int
		if err := s.Store.QueryRowContext(r.Context(),
			`SELECT password_hash, status, must_change_password FROM users WHERE lower(email) = ?`, body.Email).
			Scan(&hash, &status, &mustChange); err != nil {
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
		mustChangePassword = mustChange == 1
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
		"mustChangePassword": mustChangePassword,
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

	var userID, persisted, status string
	var admin, mustChange int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT id, role, status, is_admin, must_change_password FROM users WHERE lower(email) = ?`, email).
		Scan(&userID, &persisted, &status, &admin, &mustChange); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unknown account"})
		return
	}
	role := auth.VirtualRole(persisted, admin == 1)
	if status != "active" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "account is not active"})
		return
	}
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO sessions (token, user_id, role, expires_at)
		VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now','+12 hours'))`, token, userID, string(role))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "email": body.Email,
		"role": string(role), "userId": userID,
		"mustChangePassword": mustChange == 1,
		"expiresAt":          time.Now().Add(12 * time.Hour).Format(time.RFC3339),
	})
}

// changeOwnPassword lets a freshly bootstrapped owner replace the one-time
// password printed in the server log without an administrator round-trip.
func (s *Server) changeOwnPassword(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.FromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if len(body.NewPassword) < 8 {
		writeBadRequest(w, "new password must be at least 8 characters")
		return
	}
	var hash string
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT password_hash FROM users WHERE id = ?`, identity.UserID).Scan(&hash); err != nil {
		writeError(w, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.CurrentPassword)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is incorrect"})
		return
	}
	next, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, err)
		return
	}
	if _, err := s.Store.ExecContext(r.Context(), `
		UPDATE users SET password_hash = ?, must_change_password = 0,
		       updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`,
		string(next), identity.UserID); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "auth.password_change", "user", identity.UserID, "")
	writeJSON(w, http.StatusOK, map[string]any{"changed": true, "mustChangePassword": false})
}

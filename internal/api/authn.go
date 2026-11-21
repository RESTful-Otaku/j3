package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
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
	buf := make([]byte, 3)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%06d", (int(buf[0])<<16|int(buf[1])<<8|int(buf[2]))%1000000)
}

// startLogin supports password, Google/Apple demo providers and passkeys.
// Real OAuth/WebAuthn verification requires provider credentials and a WebAuthn
// relying-party config; this endpoint models the complete UX and MFA flow.
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
	if provider == "password" && body.Password == "" {
		writeBadRequest(w, "password is required")
		return
	}
	if provider != "password" && provider != "google" && provider != "apple" && provider != "passkey" {
		writeBadRequest(w, "unknown sign-in provider")
		return
	}
	if provider != "password" && body.Email == "" {
		writeBadRequest(w, "email is required")
		return
	}
	code := randomCode()
	challenges.mu.Lock()
	challenges.m[body.Email] = authChallenge{Email: body.Email, Method: provider, Code: code, Expires: time.Now().Add(5 * time.Minute)}
	challenges.mu.Unlock()
	// Demo mode surfaces the code so the flow can be QA'd without email/SMS.
	writeJSON(w, http.StatusOK, map[string]any{
		"challengeId": body.Email, "mfaRequired": true, "demoCode": code, "provider": provider,
	})
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
	delete(challenges.m, body.Email)
	challenges.mu.Unlock()
	if !ok || time.Now().After(challenge.Expires) || challenge.Code != body.Code {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired code"})
		return
	}
	tokenBuf := make([]byte, 16)
	_, _ = rand.Read(tokenBuf)
	writeJSON(w, http.StatusOK, map[string]any{
		"token": hex.EncodeToString(tokenBuf), "email": body.Email,
		"role": "runner", "expiresAt": time.Now().Add(12 * time.Hour).Format(time.RFC3339),
	})
}

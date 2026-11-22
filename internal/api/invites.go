package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/auth"
)

const inviteLifetime = 7 * 24 * time.Hour

type createInviteRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	ClubID   string `json:"clubId"`
	ClubName string `json:"clubName"`
	Phone    string `json:"phone"`
	EventID  string `json:"eventId"`
}

type inviteView struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Club      string `json:"club"`
	State     string `json:"state"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
	InviteURL string `json:"inviteUrl,omitempty"`
	Token     string `json:"token,omitempty"`
}

func newInviteToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func parseStoredTime(value string) (time.Time, error) {
	for _, layout := range []string{sessionTimeLayout, "2006-01-02T15:04:05.999Z", time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid stored time %q", value)
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func (s *Server) ensureClub(ctx context.Context, tx *sql.Tx, tenantID, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("club name is required")
	}
	var id string
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM clubs WHERE tenant_id = ? AND lower(name) = lower(?) LIMIT 1`, tenantID, name).
		Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	id = newID("club")
	_, err = tx.ExecContext(ctx, `
		INSERT INTO clubs (id, tenant_id, name, created_at, updated_at)
		VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now'))`,
		id, tenantID, name)
	return id, err
}

// createInvite implements the invite-only authority chain. The inviter's
// identity comes from the verified session; a forged role header or a club
// owner trying to invite outside their club is rejected.
func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "invites", "create") {
		return
	}
	identity, actorClub, ok := s.actorScope(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	actorRole, ok := s.effectiveRole(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	var body createInviteRequest
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	if !mailPattern.MatchString(email) {
		writeBadRequest(w, "a valid email is required")
		return
	}
	target := auth.CanonicalRole(body.Role)
	if target == "" {
		writeBadRequest(w, "unknown role")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = email
	}

	targetClub := strings.TrimSpace(body.ClubID)
	if actorRole == auth.RoleClubLeader {
		if targetClub != "" && targetClub != actorClub {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "club owners may only invite participants to their own club",
			})
			return
		}
		targetClub = actorClub
	}
	if !auth.CanInvite(actorRole, target, actorClub, targetClub) {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "your role cannot invite this account type or club",
		})
		return
	}

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()

	const tenantID = "demo-tenant"
	if strings.TrimSpace(body.ClubName) != "" && targetClub == "" {
		targetClub, err = s.ensureClub(r.Context(), tx, tenantID, body.ClubName)
		if err != nil {
			writeError(w, err)
			return
		}
		if !auth.CanInvite(actorRole, target, actorClub, targetClub) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot create that club"})
			return
		}
	}

	var existingID, existingStatus string
	err = tx.QueryRowContext(r.Context(),
		`SELECT id, status FROM users WHERE tenant_id = ? AND lower(email) = ?`, tenantID, email).
		Scan(&existingID, &existingStatus)
	if err != nil && err != sql.ErrNoRows {
		writeError(w, err)
		return
	}
	if err == nil && existingStatus != "invited" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "an active account already uses this email"})
		return
	}

	userID := existingID
	persisted, isAdmin := auth.PersistedRole(target)
	if userID == "" {
		userID = newID("user")
		source := "independent"
		if targetClub != "" {
			source = "club_invite"
		}
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO users (id, tenant_id, club_id, role, is_admin, display_name, email,
			                   password_hash, source, status, invited_by, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, '', ?, 'invited', ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))`,
			userID, tenantID, nullable(targetClub), persisted, boolInt(isAdmin), name, email, source, identity.UserID)
	} else {
		_, err = tx.ExecContext(r.Context(), `
			UPDATE users SET club_id = ?, role = ?, is_admin = ?, display_name = ?,
			       invited_by = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
			WHERE id = ?`, nullable(targetClub), persisted, boolInt(isAdmin), name, identity.UserID, userID)
	}
	if err != nil {
		writeError(w, err)
		return
	}

	if target == auth.RoleFighter {
		var fighterID string
		err = tx.QueryRowContext(r.Context(), `SELECT id FROM fighters WHERE user_id = ?`, userID).Scan(&fighterID)
		if err == sql.ErrNoRows {
			_, err = tx.ExecContext(r.Context(), `
				INSERT INTO fighters (id, tenant_id, user_id, club_id, display_name)
				VALUES (?, ?, ?, ?, ?)`, newID("fighter"), tenantID, userID, nullable(targetClub), name)
		}
		if err != nil {
			writeError(w, err)
			return
		}
	}

	if _, err := tx.ExecContext(r.Context(),
		`UPDATE invitations SET state = 'revoked' WHERE tenant_id = ? AND lower(email) = ? AND state = 'pending'`,
		tenantID, email); err != nil {
		writeError(w, err)
		return
	}

	token, err := newInviteToken()
	if err != nil {
		writeError(w, err)
		return
	}
	inviteID := newID("invite")
	expiresAt := time.Now().UTC().Add(inviteLifetime).Format(sessionTimeLayout)
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO invitations (id, tenant_id, event_id, club_id, role, email, token_hash,
		                         state, expires_at, invited_by, actor_role, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))`,
		inviteID, tenantID, nullable(body.EventID), nullable(targetClub), string(target), email,
		hashInviteToken(token), expiresAt, identity.UserID, string(actorRole))
	if err != nil {
		writeError(w, err)
		return
	}
	after, _ := json.Marshal(map[string]any{
		"email": email, "role": target, "clubId": targetClub, "inviteId": inviteID,
	})
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "invite.create", "user", userID, string(after))
	writeJSON(w, http.StatusCreated, inviteView{
		ID: inviteID, Email: email, Name: name, Role: string(target),
		Club: strings.TrimSpace(body.ClubName), State: "pending", ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC().Format(sessionTimeLayout),
		Token:     token, InviteURL: "/invite/" + token,
	})
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// listInvites is scoped so club owners only see invitations for their club.
func (s *Server) listInvites(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "invites", "view") {
		return
	}
	identity, actorClub, ok := s.actorScope(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	actorRole, _ := s.effectiveRole(r)
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT i.id, i.email, COALESCE(u.display_name, i.email), i.role,
		       COALESCE(c.name, ''), i.state, i.expires_at, i.created_at
		FROM invitations i
		LEFT JOIN users u ON u.email = i.email AND u.tenant_id = i.tenant_id
		LEFT JOIN clubs c ON c.id = i.club_id
		WHERE i.tenant_id = 'demo-tenant' AND (
		      ? IN ('admin','tournament_runner') OR i.invited_by = ?
		      OR (? = 'club_leader' AND i.club_id = ?))
		ORDER BY i.created_at DESC`, string(actorRole), identity.UserID, string(actorRole), actorClub)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	invites := []inviteView{}
	for rows.Next() {
		var item inviteView
		if err := rows.Scan(&item.ID, &item.Email, &item.Name, &item.Role, &item.Club,
			&item.State, &item.ExpiresAt, &item.CreatedAt); err != nil {
			writeError(w, err)
			return
		}
		invites = append(invites, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": invites})
}

// acceptInvite activates an invited account. It is public because the invite
// token is the credential; no session or role header is trusted here.
func (s *Server) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if len(body.Password) < 8 {
		writeBadRequest(w, "password must be at least 8 characters")
		return
	}
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()

	var inviteID, email, role, state, expiresAt string
	var clubID sql.NullString
	err = tx.QueryRowContext(r.Context(), `
		SELECT id, email, role, state, expires_at, club_id
		FROM invitations WHERE token_hash = ? AND tenant_id = 'demo-tenant'`, hashInviteToken(body.Token)).
		Scan(&inviteID, &email, &role, &state, &expiresAt, &clubID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invitation"})
		return
	}
	if state != "pending" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invitation has already been used"})
		return
	}
	expiry, err := parseStoredTime(expiresAt)
	if err != nil || time.Now().After(expiry) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invitation has expired"})
		return
	}
	hash, err := bcryptHash(body.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	var userID string
	if err := tx.QueryRowContext(r.Context(),
		`SELECT id FROM users WHERE tenant_id = 'demo-tenant' AND lower(email) = lower(?)`, email).
		Scan(&userID); err != nil {
		writeError(w, err)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = email
	}
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE users SET display_name = ?, password_hash = ?, status = 'active',
		       must_change_password = 0, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?`, name, hash, userID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE invitations SET state = 'accepted', accepted_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?`, inviteID); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	after, _ := json.Marshal(map[string]any{"email": email, "role": role, "clubId": clubID.String})
	actorReq := r.WithContext(auth.WithIdentity(r.Context(), auth.Identity{
		UserID: userID, Role: auth.CanonicalRole(role),
	}))
	s.audit(actorReq, "invite.accept", "user", userID, string(after))
	writeJSON(w, http.StatusOK, map[string]any{
		"email": email, "role": role, "userId": userID, "status": "active",
	})
}

func (s *Server) revokeInvite(w http.ResponseWriter, r *http.Request) {
	if !s.requireAction(w, r, "invites", "edit") {
		return
	}
	identity, actorClub, ok := s.actorScope(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	actorRole, _ := s.effectiveRole(r)
	result, err := s.Store.ExecContext(r.Context(), `
		UPDATE invitations SET state = 'revoked'
		WHERE id = ? AND tenant_id = 'demo-tenant' AND state = 'pending' AND (
		      ? IN ('admin','tournament_runner') OR invited_by = ?
		      OR (? = 'club_leader' AND club_id = ?))`,
		r.PathValue("id"), string(actorRole), identity.UserID, string(actorRole), actorClub)
	if err != nil {
		writeError(w, err)
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invitation not found"})
		return
	}
	s.audit(r, "invite.revoke", "invite", r.PathValue("id"), "")
	writeJSON(w, http.StatusOK, map[string]any{"revoked": true})
}

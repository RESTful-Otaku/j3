package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/restful-otaku/j3/internal/auth"
	"github.com/restful-otaku/j3/internal/rules"
	"github.com/restful-otaku/j3/internal/units"
)

type signupRequest struct {
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	DOB           string   `json:"dateOfBirth"`
	UnitSystem    string   `json:"unitSystem"`
	HeightCm      float64  `json:"heightCm"`
	WeightKg      float64  `json:"weightKg"`
	HeightFt      float64  `json:"heightFt"`
	HeightIn      float64  `json:"heightIn"`
	WeightLb      float64  `json:"weightLb"`
	BeltRankID    string   `json:"beltRankId"`
	Gender        string   `json:"gender"`
	AccountType   string   `json:"accountType"`
	ClubID        string   `json:"clubId"`
	CategoryIDs   []string `json:"categoryIds"`
	ConsentMoveUp bool     `json:"consentMoveUp"`
}

type signupResult struct {
	UserID        string              `json:"userId"`
	Status        string              `json:"status"`
	Role          string              `json:"role"`
	RoleRequestID string              `json:"roleRequestId,omitempty"`
	HeightCm      float64             `json:"heightCm"`
	WeightKg      float64             `json:"weightKg"`
	Eligible      []rules.Eligibility `json:"eligible"`
	Selected      []string            `json:"selected"`
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func canonicalStats(req signupRequest) (heightCm, weightKg float64, err error) {
	switch strings.ToLower(req.UnitSystem) {
	case "imperial":
		heightCm, err = units.FeetInchesToCm(req.HeightFt, req.HeightIn)
		if err != nil {
			return 0, 0, err
		}
		weightKg, err = units.ToKg(req.WeightLb, "lb")
	default:
		heightCm, err = units.ToCm(req.HeightCm, "cm")
		if err != nil {
			return 0, 0, err
		}
		weightKg, err = units.ToKg(req.WeightKg, "kg")
	}
	if err != nil {
		return 0, 0, err
	}
	if heightCm < 80 || heightCm > 250 {
		return 0, 0, errors.New("height must be between 80 and 250 cm")
	}
	if weightKg < 15 || weightKg > 350 {
		return 0, 0, errors.New("weight must be between 15 and 350 kg")
	}
	return heightCm, weightKg, nil
}

func (s *Server) categoriesForSignup(r *http.Request, eventID string) ([]rules.Category, error) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT c.id, c.name, c.kind, c.gender, c.min_value, c.max_value,
		       COALESCE(br.sort_order, 0)
		FROM categories c
		LEFT JOIN belt_ranks br ON br.id = c.belt_rank_id
		WHERE c.event_id = ? ORDER BY c.sort_order`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories := []rules.Category{}
	for rows.Next() {
		var c rules.Category
		var min, max, order sql.NullFloat64
		if err := rows.Scan(&c.ID, &c.Name, &c.Kind, &c.Gender, &min, &max, &order); err != nil {
			return nil, err
		}
		c.Min, c.Max = min.Float64, max.Float64
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *Server) beltOrder(r *http.Request, beltRankID string) int {
	if beltRankID == "" {
		return 0
	}
	var order int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT sort_order FROM belt_ranks WHERE id = ?`, beltRankID).Scan(&order); err != nil {
		return 0
	}
	return order
}

func (s *Server) signupPreview(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	heightCm, weightKg, err := canonicalStats(req)
	if err != nil {
		writeError(w, err)
		return
	}
	categories, err := s.categoriesForSignup(r, "ev-demo")
	if err != nil {
		writeError(w, err)
		return
	}
	stats := rules.Stats{
		DOB: req.DOB, WeightKg: weightKg, HeightCm: heightCm,
		BeltOrder: s.beltOrder(r, req.BeltRankID), Gender: req.Gender,
	}
	eligible := rules.Compatible(categories, stats, time.Now(), req.ConsentMoveUp)
	writeJSON(w, http.StatusOK, map[string]any{
		"heightCm": heightCm, "weightKg": weightKg, "eligible": eligible,
	})
}

func (s *Server) signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, errors.New("name is required"))
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		writeError(w, errors.New("a valid email is required"))
		return
	}
	if _, err := time.Parse("2006-01-02", req.DOB); err != nil {
		writeError(w, errors.New("date of birth must be YYYY-MM-DD"))
		return
	}
	heightCm, weightKg, err := canonicalStats(req)
	if err != nil {
		writeError(w, err)
		return
	}

	role := auth.Role(req.AccountType)
	if !auth.Valid(role) {
		role = auth.RoleFighter
	}
	privileged := role == auth.RoleOfficial || role == auth.RoleTournamentRunner ||
		role == auth.RoleClubLeader

	gender := req.Gender
	if gender == "" {
		gender = "open"
	}
	userID := newID("u")
	status := "active"
	roleRequestID := ""
	roleRequestStatus := "none"
	if privileged {
		status = "invited"
		roleRequestStatus = "pending"
		roleRequestID = newID("req")
	}

	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO users
		 (id, tenant_id, club_id, role, display_name, email, phone, date_of_birth,
		  height_cm, weight_kg, height_unit, weight_unit, belt_rank_id, gender,
		  source, status, role_request, role_request_status)
		VALUES (?, 'demo-tenant', NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, req.ClubID, string(role), req.Name, req.Email, req.Phone, req.DOB,
		heightCm, weightKg, req.UnitSystem, req.UnitSystem, nullIfEmpty(req.BeltRankID),
		gender, "independent", status,
		nullIfEmpty(string(role)), roleRequestStatus)
	if err != nil {
		writeError(w, err)
		return
	}

	if roleRequestID != "" {
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO role_requests (id, tenant_id, user_id, requested_role, status)
			VALUES (?, 'demo-tenant', ?, ?, 'pending')`,
			roleRequestID, userID, string(role)); err != nil {
			writeError(w, err)
			return
		}
	}

	if role == auth.RoleFighter {
		fighterID := newID("f")
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO fighters
			 (id, tenant_id, user_id, club_id, belt_rank_id, display_name, date_of_birth,
			  weight_kg, gender)
			VALUES (?, 'demo-tenant', ?, NULLIF(?, ''), ?, ?, ?, ?, ?)`,
			fighterID, userID, req.ClubID, nullIfEmpty(req.BeltRankID), req.Name,
			req.DOB, weightKg, gender); err != nil {
			writeError(w, err)
			return
		}
		for _, categoryID := range req.CategoryIDs {
			if _, err := tx.ExecContext(r.Context(), `
				INSERT OR IGNORE INTO entrants (id, event_id, fighter_id, category_id, club_id)
				VALUES (?, 'ev-demo', ?, ?, NULLIF(?, ''))`,
				newID("ent"), fighterID, categoryID, req.ClubID); err != nil {
				writeError(w, err)
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}

	result := signupResult{
		UserID: userID, Status: status, Role: string(role), RoleRequestID: roleRequestID,
		HeightCm: heightCm, WeightKg: weightKg, Selected: req.CategoryIDs, Eligible: []rules.Eligibility{},
	}
	writeJSON(w, http.StatusCreated, result)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *Server) roleRequests(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT rr.id, rr.requested_role, rr.status, u.display_name, u.email
		FROM role_requests rr JOIN users u ON u.id = rr.user_id
		ORDER BY rr.created_at DESC`)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	type request struct {
		ID, Role, Status, Name, Email string
	}
	requests := []request{}
	for rows.Next() {
		var item request
		if err := rows.Scan(&item.ID, &item.Role, &item.Status, &item.Name, &item.Email); err != nil {
			writeError(w, err)
			return
		}
		requests = append(requests, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": requests})
}

func (s *Server) approveRoleRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tx, err := s.Store.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	defer tx.Rollback()
	var userID, role string
	if err := tx.QueryRowContext(r.Context(),
		`SELECT user_id, requested_role FROM role_requests WHERE id = ? AND status = 'pending'`, id).
		Scan(&userID, &role); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE role_requests SET status = 'approved', reviewed_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?`, id); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
		UPDATE users SET role = ?, status = 'active', role_request_status = 'approved',
		  updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id = ?`, role, userID); err != nil {
		writeError(w, err)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `
		INSERT INTO audit_log (tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		VALUES ('demo-tenant', 'u-runner', 'tournament_runner', 'role.approve', 'role_request', ?, ?)`,
		id, fmt.Sprintf(`{"role":%q}`, role)); err != nil {
		writeError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "approved", "role": role})
}

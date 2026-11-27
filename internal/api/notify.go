package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// notifyKinds are feedback/confirmation + marketing/info notification kinds.
var notifyKinds = map[string]bool{
	"announcement": true, "info": true, "feedback": true, "confirmation": true,
	"account": true, "preference": true, "invite": true, "marketing": true,
}

// fanoutNotifications records per-recipient, per-channel deliveries for every
// relevant user type. External providers are optional via webhook; without
// credentials it still records intent + local delivery (and demo payloads).
func (s *Server) fanoutNotifications(eventID, kind, title, body, channels, audience string) []string {
	if !notifyKinds[kind] {
		kind = "announcement"
	}
	recipients := s.recipientsForAudience(audience)
	channelList := []string{}
	for _, c := range strings.Split(channels, ",") {
		if t := strings.TrimSpace(c); t != "" {
			channelList = append(channelList, t)
		}
	}
	if len(channelList) == 0 {
		channelList = []string{"push"}
	}
	delivered := []string{"local"}
	webhook := strings.TrimSpace(os.Getenv("J3_NOTIFY_WEBHOOK_URL"))
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, recipient := range recipients {
		for _, channel := range channelList {
			status := "recorded"
			errMsg := ""
			if webhook != "" {
				payload := map[string]any{"recipient": recipient, "channel": channel, "kind": kind, "title": title, "body": body, "audience": audience, "eventId": eventID}
				if err := postWebhook(webhook, payload); err != nil {
					status = "failed"; errMsg = err.Error()
				} else {
					status = "webhook-sent"; delivered = appendOnce(delivered, channel+":"+recipient)
				}
			} else {
				delivered = appendOnce(delivered, channel)
			}
			_, _ = s.Store.Exec(
				`INSERT INTO notification_deliveries (id, event_id, recipient, channel, kind, title, body, audience, status, error, created_at, delivered_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				newID("nd"), eventID, recipient, channel, kind, title, body, audience, status, errMsg, now, now)
		}
	}
	return delivered
}

func (s *Server) recipientsForAudience(audience string) []string {
	roles := map[string]bool{"all": true, "fighters": true, "officials": true, "clubs": true,
		"guardians": true, "participants": true, "spectators": true, "admins": true}
	if !roles[audience] {
		audience = "all"
	}
	roleFilter := []string{}
	switch audience {
	case "fighters", "participants":
		roleFilter = []string{"fighter"}
	case "officials":
		roleFilter = []string{"official", "tournament_runner"}
	case "clubs":
		roleFilter = []string{"club_leader"}
	case "admins":
		roleFilter = []string{"tournament_runner"}
	case "spectators":
		roleFilter = []string{"spectator"}
	}
	recipients := []string{}
	query := `SELECT COALESCE(email,'') FROM users WHERE status='active'`
	if len(roleFilter) == 0 {
		// all users
		if rs, err := s.Store.Query(query); err == nil {
			defer rs.Close()
			for rs.Next() {
				var email string
				if rs.Scan(&email) == nil && email != "" {
					recipients = append(recipients, email)
				}
			}
		}
		return recipients
	}
	// role-filtered
	ph := make([]string, 0, len(roleFilter))
	args := []any{}
	for _, r := range roleFilter { ph = append(ph, "?"); args = append(args, r) }
	q := query + ` AND role IN (` + strings.Join(ph, ",") + `)`
	if rs, err := s.Store.Query(q, args...); err == nil {
		defer rs.Close()
		for rs.Next() {
			var email string
			if rs.Scan(&email) == nil && email != "" {
				recipients = append(recipients, email)
			}
		}
	}
	return recipients
}

func appendOnce(list []string, value string) []string {
	for _, v := range list { if v == value { return list } }
	return append(list, value)
}

func postWebhook(url string, payload map[string]any) error {
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { return &webhookError{status: resp.StatusCode} }
	return nil
}

type webhookError struct{ status int }
func (e *webhookError) Error() string { return "webhook status " + itoa(e.status) }
func itoa(n int) string { if n == 0 { return "0" }; var b []byte; for n > 0 { b = append([]byte{byte('0' + n%10)}, b...); n /= 10 }; return string(b) }

// Package seed creates a realistic demo tournament for local QA.
package seed

import (
	"fmt"

	"github.com/restful-otaku/j3/internal/db"
)

const tenantID = "demo-tenant"

// Ensure inserts demo data idempotently; it is safe to call on every boot.
func Ensure(store *db.DB) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT OR IGNORE INTO tenants (id, name, slug, locale) VALUES (?, ?, ?, ?)`,
			[]any{tenantID, "J3 Demo Dojo", "demo", "en"}},
		{`INSERT OR IGNORE INTO themes (id, tenant_id, name, tokens_json) VALUES (?, ?, ?, ?)`,
			[]any{"demo-theme", tenantID, "J3 Default", `{"primary":"#2563eb","accent":"#f59e0b","ink":"#0f172a"}`}},
		{`INSERT OR IGNORE INTO clubs (id, tenant_id, name, primary_colour) VALUES (?, ?, ?, ?)`,
			[]any{"club-riverside", tenantID, "Riverside Jiu Jitsu", "#2563eb"}},
		{`INSERT OR IGNORE INTO clubs (id, tenant_id, name, primary_colour) VALUES (?, ?, ?, ?)`,
			[]any{"club-hilltop", tenantID, "Hilltop Dojo", "#dc2626"}},
		{`INSERT OR IGNORE INTO clubs (id, tenant_id, name, primary_colour) VALUES (?, ?, ?, ?)`,
			[]any{"club-city", tenantID, "City Warriors", "#16a34a"}},
	}

	for i, rank := range []struct {
		id, label, colour, kind string
		order                   int
	}{
		{"br-white", "White", "#f8fafc", "kyu", 1},
		{"br-grey", "Grey", "#94a3b8", "youth", 2},
		{"br-yellow", "Yellow", "#facc15", "kyu", 3},
		{"br-orange", "Orange", "#fb923c", "kyu", 4},
		{"br-green", "Green", "#22c55e", "kyu", 5},
		{"br-blue", "Blue", "#3b82f6", "kyu", 6},
		{"br-purple", "Purple", "#a855f7", "kyu", 7},
		{"br-brown", "Brown", "#92400e", "kyu", 8},
		{"br-black", "Black", "#0f172a", "dan", 9},
		{"br-redwhite", "Red & White (mon)", "#ef4444", "dan", 10},
		{"br-red", "Red", "#b91c1c", "dan", 11},
		{"br-custom", "Junior Samurai (custom)", "#14b8a6", "custom", 12},
	} {
		isCustom := 0
		if rank.kind == "custom" {
			isCustom = 1
		}
		statements = append(statements, struct {
			query string
			args  []any
		}{
			`INSERT OR IGNORE INTO belt_ranks
			 (id, tenant_id, label, colour, sort_order, grade_kind, is_custom)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			[]any{rank.id, tenantID, rank.label, rank.colour, i + 1, rank.kind, isCustom},
		})
	}

	type demoUser struct{ id, name, email, role, clubID, source string }
	for _, u := range []demoUser{
		{"u-runner", "Aiko Tanaka (Runner)", "runner@demo.j3", "tournament_runner", "", "independent"},
		{"u-official", "Kenji Sato (Official)", "official@demo.j3", "official", "", "independent"},
		{"u-official-2", "Nina Sutherland (Official)", "official2@demo.j3", "official", "", "independent"},
		{"u-sensie", "Shihan H. Yamamoto (Guest)", "sensie@demo.j3", "official", "", "independent"},
		{"u-leader", "Laura Mills (Riverside)", "leader@demo.j3", "club_leader", "club-riverside", "independent"},
		{"u-leader-hilltop", "Marcus Reed (Hilltop)", "leader-hilltop@demo.j3", "club_leader", "club-hilltop", "independent"},
		{"u-leader-city", "Priya Shah (City)", "leader-city@demo.j3", "club_leader", "club-city", "independent"},
		{"u-spectator", "Sam Viewer (Spectator)", "viewer@demo.j3", "spectator", "", "independent"},
		{"u-f1", "Yuki Mori", "yuki@demo.j3", "fighter", "club-riverside", "club_invite"},
		{"u-f2", "Daniel Cruz", "daniel@demo.j3", "fighter", "club-hilltop", "club_invite"},
		{"u-f3", "Aisha Khan", "aisha@demo.j3", "fighter", "club-city", "independent"},
		{"u-f4", "Tom Nakamura", "tom@demo.j3", "fighter", "club-riverside", "club_invite"},
		{"u-f5", "Sofia Rossi", "sofia@demo.j3", "fighter", "club-hilltop", "club_invite"},
		{"u-f6", "Haruto Sato", "haruto@demo.j3", "fighter", "club-city", "independent"},
		{"u-f7", "Mia Brown", "mia@demo.j3", "fighter", "club-riverside", "club_invite"},
		{"u-f8", "Ravi Patel", "ravi@demo.j3", "fighter", "club-hilltop", "club_invite"},
	} {
		statements = append(statements, struct {
			query string
			args  []any
		}{
			`INSERT OR IGNORE INTO users
			 (id, tenant_id, club_id, role, display_name, email, source, status)
			 VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, 'active')`,
			[]any{u.id, tenantID, u.clubID, u.role, u.name, u.email, u.source},
		})
	}

	type demoFighter struct {
		id, name, userID, clubID, beltID, dob string
		weight                                float64
		gender                                string
	}
	for _, f := range []demoFighter{
		{"f1", "Yuki Mori", "u-f1", "club-riverside", "br-blue", "1998-04-12", 74.2, "male"},
		{"f2", "Daniel Cruz", "u-f2", "club-hilltop", "br-purple", "1996-09-03", 75.8, "male"},
		{"f3", "Aisha Khan", "u-f3", "club-city", "br-brown", "1999-01-21", 62.4, "female"},
		{"f4", "Tom Nakamura", "u-f4", "club-riverside", "br-white", "2000-06-30", 80.1, "male"},
		{"f5", "Sofia Rossi", "u-f5", "club-hilltop", "br-blue", "2001-11-14", 58.9, "female"},
		{"f6", "Haruto Sato", "u-f6", "club-city", "br-green", "2011-02-08", 52.0, "male"},
		{"f7", "Mia Brown", "u-f7", "club-riverside", "br-yellow", "2012-08-19", 49.5, "female"},
		{"f8", "Ravi Patel", "u-f8", "club-hilltop", "br-black", "1988-03-25", 89.0, "male"},
	} {
		statements = append(statements, struct {
			query string
			args  []any
		}{
			`INSERT OR IGNORE INTO fighters
			 (id, tenant_id, user_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			[]any{f.id, tenantID, f.userID, f.clubID, f.beltID, f.name, f.dob, f.weight, f.gender},
		})
	}

	statements = append(statements,
		stmt(`INSERT OR IGNORE INTO rule_sets (id, tenant_id, name, version, config_json) VALUES (?, ?, ?, ?, ?)`,
			"rs-demo", tenantID, "Generic J3 Rules", 1,
			`{"matchSeconds":300,"ippon":3,"wazaAri":1,"shidoLimit":3,"holdWazaAriSec":10,"holdIpponSec":20,"submissionInstantWin":true,"goldenScore":false,"bestOf":1}`),
		stmt(`INSERT OR IGNORE INTO format_policies (id, tenant_id, name, version, config_json) VALUES (?, ?, ?, ?, ?)`,
			"fp-demo", tenantID, "Generic Format", 1,
			`{"one":"no_contest","two":"best_of_three","three":"round_robin","four":"single_elim_third_place","five":"round_robin","six":"two_pools","sevenPlus":"double_elim","thirdPlace":"if_not_met","dualBronze":true}`),
		stmt(`INSERT OR IGNORE INTO events
		 (id, tenant_id, theme_id, rule_set_id, format_policy_id, name, starts_at, ends_at, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"ev-demo", tenantID, "demo-theme", "rs-demo", "fp-demo",
			"Spring Challenge 2026", "2026-11-19T09:00:00Z", "2026-11-19T17:00:00Z", "published"),
		stmt(`INSERT OR IGNORE INTO categories (id, event_id, name, kind, min_value, max_value, gender, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"cat-adult-76", "ev-demo", "Adult -76 kg", "combined", 0, 76, "male", 1),
		stmt(`INSERT OR IGNORE INTO categories (id, event_id, name, kind, min_value, max_value, gender, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"cat-adult-82", "ev-demo", "Adult -82 kg", "combined", 76, 82, "male", 2),
		stmt(`INSERT OR IGNORE INTO categories (id, event_id, name, kind, min_value, max_value, gender, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"cat-youth-60", "ev-demo", "Youth -60 kg", "combined", 0, 60, "open", 3),
		stmt(`INSERT OR IGNORE INTO categories (id, event_id, name, kind, gender, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
			"cat-kata", "ev-demo", "Kata Open", "combined", "open", 4),
		stmt(`INSERT OR IGNORE INTO mats (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
			"mat-1", "ev-demo", "Mat 1", 1),
		stmt(`INSERT OR IGNORE INTO mats (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
			"mat-2", "ev-demo", "Mat 2", 2),
		stmt(`INSERT OR IGNORE INTO mats (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
			"mat-kata", "ev-demo", "Kata Area", 3),
	)

	for i, entrant := range []struct {
		fighter, category, club string
		seed                    int
	}{
		{"f1", "cat-adult-76", "club-riverside", 1},
		{"f2", "cat-adult-76", "club-hilltop", 2},
		{"f3", "cat-adult-82", "club-city", 1},
		{"f4", "cat-adult-82", "club-riverside", 2},
		{"f6", "cat-youth-60", "club-city", 1},
		{"f7", "cat-youth-60", "club-riverside", 2},
		{"f5", "cat-kata", "club-hilltop", 1},
		{"f8", "cat-kata", "club-hilltop", 2},
	} {
		statements = append(statements, stmt(
			`INSERT OR IGNORE INTO entrants (id, event_id, fighter_id, category_id, club_id, seed)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("ent-%d", i+1), "ev-demo", entrant.fighter, entrant.category, entrant.club, entrant.seed,
		))
	}

	statements = append(statements,
		stmt(`INSERT OR IGNORE INTO fights
		 (id, event_id, mat_id, category_id, round_label, red_fighter_id, blue_fighter_id, state, outcome)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"fight-1", "ev-demo", "mat-1", "cat-adult-76", "Semi-final", "f1", "f2", "complete", "win_red"),
		stmt(`INSERT OR IGNORE INTO fights
		 (id, event_id, mat_id, category_id, round_label, red_fighter_id, blue_fighter_id, state, outcome)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"fight-2", "ev-demo", "mat-1", "cat-adult-82", "Semi-final", "f3", "f4", "complete", "draw"),
		stmt(`INSERT OR IGNORE INTO fights
		 (id, event_id, mat_id, category_id, round_label, red_fighter_id, blue_fighter_id, state, outcome)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"fight-3", "ev-demo", "mat-2", "cat-youth-60", "Final", "f6", "f7", "live", ""),
		stmt(`INSERT OR IGNORE INTO fights
		 (id, event_id, mat_id, category_id, round_label, red_fighter_id, blue_fighter_id, state, outcome)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"fight-4", "ev-demo", "mat-kata", "cat-kata", "Kata round 1", "f5", "f8", "pending", ""),
		stmt(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
			"fight-3", 1, "score", "u-official", `{"corner":"red","points":2,"technique":"Takedown"}`),
		stmt(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
			"fight-3", 2, "score", "u-official", `{"corner":"blue","points":1,"technique":"Sweep"}`),
		stmt(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
			"fight-3", 3, "penalty", "u-official", `{"corner":"red","card":"yellow","reason":"Passivity"}`),
		stmt(`INSERT OR IGNORE INTO invitations
		 (id, tenant_id, event_id, club_id, role, email, token_hash, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			"inv-1", tenantID, "ev-demo", "club-riverside", "fighter", "new@demo.j3", "demo-token-1", "2026-12-31T00:00:00Z"),
		stmt(`INSERT OR IGNORE INTO audit_log
		 (tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			tenantID, "u-runner", "tournament_runner", "demo.seed", "event", "ev-demo", `{"status":"published"}`),
	)

	for _, s := range statements {
		if _, err := tx.Exec(s.query, s.args...); err != nil {
			return fmt.Errorf("seed %q: %w", s.query, err)
		}
	}
	return tx.Commit()
}

type statement struct {
	query string
	args  []any
}

func stmt(query string, args ...any) statement { return statement{query, args} }

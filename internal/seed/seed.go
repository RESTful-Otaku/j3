// Package seed generates a realistic demo tournament for local QA.
//
// Scale (see AGENTS.md D-014): 2 runners, 8 officials, 4 club leaders,
// 48 fighters across 4 clubs, plus a spectator — roughly 60 accounts.
package seed

import (
	"fmt"
	"sort"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/db"
)

const tenantID = "demo-tenant"

var (
	clubs = []struct{ id, name, colour string }{
		{"club-riverside", "Riverside Jiu Jitsu", "#2563eb"},
		{"club-hilltop", "Hilltop Dojo", "#dc2626"},
		{"club-city", "City Warriors", "#16a34a"},
		{"club-sakura", "Sakura Jiu Jitsu Academy", "#7c3aed"},
	}
	givenNames = []string{
		"Yuki", "Daniel", "Aisha", "Tom", "Sofia", "Haruto", "Mia", "Ravi",
		"Emma", "Kenji", "Layla", "Noah", "Hana", "Carlos", "Zara", "Liam",
		"Aiko", "Omar", "Elena", "Jun", "Nina", "Mateo", "Sakura", "Kai",
	}
	surnames = []string{
		"Mori", "Cruz", "Khan", "Nakamura", "Rossi", "Sato", "Brown", "Patel",
		"Wilson", "Watanabe", "Haddad", "Meyer", "Ito", "Silva", "Ahmed", "Novak",
		"Tanaka", "Farouk", "Petrov", "Kim", "Okafor", "Diaz", "Yamamoto", "Chen",
	}
	beltCycle = []string{
		"br-white", "br-white", "br-white", "br-yellow", "br-orange", "br-green",
		"br-blue", "br-blue", "br-purple", "br-brown", "br-black", "br-custom",
	}
	categoryByIndex = []string{
		"cat-adult-76", "cat-adult-76", "cat-adult-82", "cat-adult-90",
		"cat-women-58", "cat-youth-60", "cat-kata",
	}
)

// Ensure inserts demo data idempotently; safe to call on every boot.
func Ensure(store *db.DB) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var stmts []statement
	add := func(query string, args ...any) {
		stmts = append(stmts, stmt(query, args...))
	}

	add(`INSERT OR IGNORE INTO tenants (id, name, slug, locale) VALUES (?, ?, ?, ?)`,
		tenantID, "J3 Demo Dojo", "demo", "en")
	add(`INSERT OR IGNORE INTO themes (id, tenant_id, name, tokens_json, logo_path, banner_path) VALUES (?, ?, ?, ?, ?, ?)`,
		"demo-theme", tenantID, "J3 Default",
		`{"primary":"#2563eb","accent":"#f59e0b","ink":"#0f172a"}`,
		"/media/logo.png", "/media/banner.png")

	for i, c := range clubs {
		add(`INSERT OR IGNORE INTO clubs (id, tenant_id, name, logo_path, primary_colour) VALUES (?, ?, ?, ?, ?)`,
			c.id, tenantID, c.name, fmt.Sprintf("/media/club_logo_%d.jpg", i+1), c.colour)
	}

	for i, rank := range []struct{ id, label, colour, kind string }{
		{"br-white", "White", "#f8fafc", "kyu"},
		{"br-grey", "Grey", "#94a3b8", "youth"},
		{"br-yellow", "Yellow", "#facc15", "kyu"},
		{"br-orange", "Orange", "#fb923c", "kyu"},
		{"br-green", "Green", "#22c55e", "kyu"},
		{"br-blue", "Blue", "#3b82f6", "kyu"},
		{"br-purple", "Purple", "#a855f7", "kyu"},
		{"br-brown", "Brown", "#92400e", "kyu"},
		{"br-black", "Black", "#0f172a", "dan"},
		{"br-redwhite", "Red & White (mon)", "#ef4444", "dan"},
		{"br-red", "Red", "#b91c1c", "dan"},
		{"br-custom", "Junior Samurai (custom)", "#14b8a6", "custom"},
	} {
		custom := 0
		if rank.kind == "custom" {
			custom = 1
		}
		add(`INSERT OR IGNORE INTO belt_ranks
			(id, tenant_id, label, colour, sort_order, grade_kind, is_custom)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			rank.id, tenantID, rank.label, rank.colour, i+1, rank.kind, custom)
	}

	// --- Users: 1 admin, 2 runners, 8 officials, 4 club leaders, 48 fighters, 1 spectator.
	// One shared demo password makes every seeded account easy to test.
	demoHash, _ := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	addUser := func(id, name, email, role, clubID, source string) {
		add(`INSERT OR IGNORE INTO users
			(id, tenant_id, club_id, role, display_name, email, password_hash, source, status)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, 'active')`,
			id, tenantID, clubID, role, name, email, string(demoHash), source)
	}
	// Owner/admin: a flagged runner so the original role CHECK stays valid.
	// This account can invite other admins, runners, officials and club owners.
	add(`INSERT OR IGNORE INTO users
		(id, tenant_id, role, is_admin, display_name, email, password_hash, source, status)
		VALUES ('u-owner', ?, 'tournament_runner', 1, 'J3 Owner (Admin)',
		        'owner@demo.j3', ?, 'independent', 'active')`, tenantID, string(demoHash))

	addUser("u-runner", "Aiko Tanaka (Runner)", "runner@demo.j3", "tournament_runner", "", "independent")
	addUser("u-runner-2", "Marcus Bell (Runner)", "runner2@demo.j3", "tournament_runner", "", "independent")

	officialNames := []string{
		"Kenji Sato", "Nina Sutherland", "Diego Alvarez", "Shihan H. Yamamoto",
		"Grace Okafor", "Tobias Reinhardt", "Mei Lin", "Owen Bradley",
	}
	for i, name := range officialNames {
		id := "u-official"
		if i > 0 {
			id = fmt.Sprintf("u-official-%d", i+1)
		}
		addUser(id, fmt.Sprintf("%s (Official)", name), fmt.Sprintf("official%d@demo.j3", i+1),
			"official", "", "independent")
	}

	leaderNames := []string{"Laura Mills", "Marcus Reed", "Priya Shah", "Yoshi Tanaka"}
	for i, name := range leaderNames {
		addUser(fmt.Sprintf("u-leader-%d", i+1), fmt.Sprintf("%s (%s)", name, clubs[i].name),
			fmt.Sprintf("leader%d@demo.j3", i+1), "club_leader", clubs[i].id, "independent")
	}
	// Compatibility alias used by QA scripts and tests.
	addUser("u-leader", "Laura Mills (Riverside)", "leader@demo.j3", "club_leader", clubs[0].id, "independent")

	addUser("u-spectator", "Sam Viewer (Spectator)", "viewer@demo.j3", "spectator", "", "independent")

	fighters := make([]struct {
		id, userID, name, clubID, beltID, dob, gender string
		weight                                        float64
	}, 0, 48)
	for i := 0; i < 48; i++ {
		name := fmt.Sprintf("%s %s", givenNames[i%len(givenNames)], surnames[(i+i/24)%len(surnames)])
		id := fmt.Sprintf("u-fighter-%02d", i+1)
		clubID := clubs[i%len(clubs)].id
		source := "club_invite"
		if i%7 == 0 {
			clubID = ""
			source = "independent"
		}
		addUser(id, fmt.Sprintf("%s (Fighter)", name), fmt.Sprintf("fighter-%02d@demo.j3", i+1),
			"fighter", clubID, source)
		year := 1990 + i%28
		if i >= 42 {
			year = 2010 + i%4
		}
		gender := "male"
		if i%3 == 0 {
			gender = "female"
		} else if i%11 == 0 {
			gender = "open"
		}
		fighters = append(fighters, struct {
			id, userID, name, clubID, beltID, dob, gender string
			weight                                        float64
		}{
			fmt.Sprintf("f%02d", i+1), id, name, clubID, beltCycle[i%len(beltCycle)],
			fmt.Sprintf("%d-%02d-%02d", year, 1+i%12, 1+i%28), gender,
			54 + float64(i%36) + 0.4,
		})
	}

	for _, f := range fighters {
		add(`INSERT OR IGNORE INTO fighters
			(id, tenant_id, user_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender, avatar_path)
			VALUES (?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`,
			f.id, tenantID, f.userID, f.clubID, f.beltID, f.name, f.dob, f.weight, f.gender,
			fmt.Sprintf("/media/fighter_%d.jpg", ((f.id[1]-'0')-1)%6+1))
	}

	add(`INSERT OR IGNORE INTO rule_sets (id, tenant_id, name, version, config_json) VALUES (?, ?, ?, ?, ?)`,
		"rs-demo", tenantID, "Generic J3 Rules", 1,
		`{"matchSeconds":300,"ippon":3,"wazaAri":1,"shidoLimit":3,"holdWazaAriSec":10,"holdIpponSec":20,"submissionInstantWin":true,"goldenScore":false,"bestOf":1,"venue":"Riverside Sports Centre","protection":["Clean pressed Gi","Belt","Mouth guard","Head guard for youth divisions","Instep / shin guards for Randori"],`+cornerDefaultsJSON+`,"schedule":[{"time":"08:30","label":"Doors open & weigh-in","kind":"general"},{"time":"09:00","label":"Opening ceremony","kind":"ceremony"},{"time":"12:30","label":"Lunch break","kind":"break"},{"time":"16:30","label":"Medal ceremony","kind":"ceremony"}],"amenities":["Changing rooms","First aid","Café","Spectator seating"]}`)
	add(`INSERT OR IGNORE INTO format_policies (id, tenant_id, name, version, config_json) VALUES (?, ?, ?, ?, ?)`,
		"fp-demo", tenantID, "Generic Format", 1,
		`{"one":"no_contest","two":"best_of_three","three":"round_robin","four":"single_elim_third_place","five":"round_robin","six":"two_pools","sevenPlus":"double_elim","thirdPlace":"if_not_met","dualBronze":true}`)
	add(`INSERT OR IGNORE INTO events
		(id, tenant_id, theme_id, rule_set_id, format_policy_id, name, starts_at, ends_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"ev-demo", tenantID, "demo-theme", "rs-demo", "fp-demo",
		"Spring Challenge 2026", "2026-11-19T09:00:00Z", "2026-11-19T17:00:00Z", "published")

	for i, category := range []struct {
		id, name, kind, gender string
		min, max               any
	}{
		{"cat-adult-76", "Adult -76 kg", "combined", "male", 0, 76},
		{"cat-adult-82", "Adult -82 kg", "combined", "male", 76, 82},
		{"cat-adult-90", "Adult +82 kg", "combined", "male", 82, 999},
		{"cat-women-58", "Women -58 kg", "combined", "female", 0, 58},
		{"cat-youth-60", "Youth -60 kg", "combined", "open", 0, 60},
		{"cat-kata", "Kata Open", "combined", "open", nil, nil},
	} {
		add(`INSERT OR IGNORE INTO categories
			(id, event_id, name, kind, min_value, max_value, gender, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			category.id, "ev-demo", category.name, category.kind,
			category.min, category.max, category.gender, i+1)
	}
	for i, name := range []string{"Mat 1", "Mat 2", "Mat 3", "Kata Area"} {
		add(`INSERT OR IGNORE INTO mats (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("mat-%d", i+1), "ev-demo", name, i+1)
	}
	for i, discipline := range []struct{ id, name string }{
		{"disc-fight", "JJ Fighting System"},
		{"disc-newaza", "Ne-Waza / Katamewaza"},
		{"disc-kata", "Kata"},
	} {
		add(`INSERT OR IGNORE INTO disciplines (id, event_id, name, rules_json, sort_order)
			VALUES (?, 'ev-demo', ?, '{}', ?)`, discipline.id, discipline.name, i+1)
	}
	add(`INSERT OR IGNORE INTO discipline_formats
		(event_id, discipline_id, format, knockout, third_place, dual_bronze, retry_ties, order_policy)
		VALUES ('ev-demo', 'disc-fight', 'single_elimination', 'single_elimination', 'rematch', 1, 0, 'weight_then_age')`)

	// Entrants: assign by index-derived category, with kata as a second division.
	byCategory := map[string][]string{}
	for i, f := range fighters {
		cat := categoryByIndex[i%len(categoryByIndex)]
		add(`INSERT OR IGNORE INTO entrants
			(id, event_id, fighter_id, category_id, club_id, seed)
			VALUES (?, ?, ?, ?, NULLIF(?, ''), ?)`,
			fmt.Sprintf("ent-%02d", i+1), "ev-demo", f.id, cat, f.clubID, i+1)
		byCategory[cat] = append(byCategory[cat], f.id)
		if i%6 == 0 {
			add(`INSERT OR IGNORE INTO entrants
				(id, event_id, fighter_id, category_id, club_id, seed)
				VALUES (?, ?, ?, ?, NULLIF(?, ''), ?)`,
				fmt.Sprintf("ent-kata-%02d", i+1), "ev-demo", f.id, "cat-kata", f.clubID, i+1)
			byCategory["cat-kata"] = append(byCategory["cat-kata"], f.id)
		}
	}

	// Fights: pair entrants in each category; a few are complete, one live.
	fightNo := 0
	fightIDs := []string{}
	catKeys := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		catKeys = append(catKeys, cat)
	}
	sort.Strings(catKeys)
	for _, cat := range catKeys {
		ids := byCategory[cat]
		for i := 0; i+1 < len(ids) && fightNo < 30; i += 2 {
			fightNo++
			id := fmt.Sprintf("fight-%02d", fightNo)
			fightIDs = append(fightIDs, id)
			state, outcome := "pending", ""
			switch {
			case i < 4:
				state, outcome = "complete", "win_red"
				if i%2 == 1 {
					outcome = "win_blue"
				}
			case i == 4:
				state = "live"
			}
			matID := "mat-1"
			if cat == "cat-kata" {
				matID = "mat-4"
			}
			roundLabel := fmt.Sprintf("Round %d", 1+(fightNo%3))
			if cat == "cat-kata" {
				roundLabel = "Kata " + roundLabel
			} else if cat == "cat-adult-82" {
				roundLabel = "Ne-Waza " + roundLabel
			}
			add(`INSERT OR IGNORE INTO fights
				(id, event_id, mat_id, category_id, round_label,
				 red_fighter_id, blue_fighter_id, state, outcome)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				id, "ev-demo", matID, cat, roundLabel,
				ids[i], ids[i+1], state, outcome)
			if state == "live" {
				add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
					id, 1, "score", "u-official", `{"corner":"red","points":2,"technique":"Takedown"}`)
				add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
					id, 2, "score", "u-official", `{"corner":"blue","points":1,"technique":"Sweep"}`)
				add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, actor_user_id, payload_json) VALUES (?, ?, ?, ?, ?)`,
					id, 3, "penalty", "u-official", `{"corner":"red","card":"yellow","reason":"Passivity"}`)
			}
		}
	}

	// Officials assigned per contest: shown on the big-screen scoreboard.
	officialIDs := []string{"u-official", "u-official-2", "u-official-3", "u-official-4"}
	for i, fightID := range fightIDs {
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, ?)`,
			fightID, officialIDs[i%len(officialIDs)], "referee")
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, ?)`,
			fightID, officialIDs[(i+1)%len(officialIDs)], "judge")
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, ?)`,
			fightID, officialIDs[(i+2)%len(officialIDs)], "scorekeeper")
	}

	for i, item := range []struct {
		name, area, start string
		duration          int
	}{
		{"Doors open & weigh-in", "Reception", "2026-11-19T08:30:00Z", 1800},
		{"Opening ceremony", "Main hall", "2026-11-19T09:00:00Z", 900},
		{"Adult -76 kg · Pool stage", "Mat 1", "2026-11-19T09:30:00Z", 1800},
		{"Youth -60 kg · Pool stage", "Mat 2", "2026-11-19T10:30:00Z", 1800},
		{"Kata Open · Round 1", "Kata Area", "2026-11-19T11:30:00Z", 1200},
		{"Lunch break", "Café", "2026-11-19T12:30:00Z", 3600},
		{"Semi-finals", "Mat 1", "2026-11-19T14:00:00Z", 1800},
		{"Medal ceremony", "Podium", "2026-11-19T16:30:00Z", 900},
	} {
		add(`INSERT OR IGNORE INTO schedule_items
			(id, event_id, name, area, starts_at, duration_seconds, sort_order)
			VALUES (?, 'ev-demo', ?, ?, ?, ?, ?)`,
			fmt.Sprintf("sched-%02d", i+1), item.name, item.area, item.start, item.duration, i)
	}
	add(`INSERT OR IGNORE INTO schedule_state (event_id, offset_minutes) VALUES ('ev-demo', 0)`)

	for i, kata := range []struct {
		name, style, level, discipline string
		weapon                         int
	}{
		{"Hekigan-ken", "JJJ", "standard", "kata", 0},
		{"Nage-no-kata", "Judo", "standard", "kata", 0},
		{"Katame-no-kata", "Judo", "standard", "kata", 0},
		{"Seiyunchin", "Karate", "advanced", "kata", 0},
		{"Kanku-dai", "Karate", "advanced", "kata", 0},
		{"Gekusai-dai", "JJJ", "standard", "kata", 0},
		{"Gekusai-sho", "JJJ", "standard", "kata", 0},
		{"Pinan sono ichi", "Karate", "beginner", "kata", 0},
		{"Taikyoku sono ichi", "Karate", "beginner", "kata", 0},
		{"Koryu-no-kata", "JJJ", "weapon", "weapon kata", 1},
		{"Tanto-no-kata", "JJJ", "weapon", "weapon kata", 1},
		{"Bo-no-kata", "JJJ", "weapon", "weapon kata", 1},
		{"Aikidoiken", "JJJ", "standard", "kata", 0},
		{"Shinjitsu-ryu", "JJJ", "advanced", "kata", 0},
		{"Broken ladder form", "Custom", "custom", "kata", 0},
	} {
		add(`INSERT OR IGNORE INTO kata_catalogue (id, name, style, level, discipline, is_weapon)
			VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("kata-%02d", i+1), kata.name, kata.style, kata.level, kata.discipline, kata.weapon)
	}
	for i, session := range []struct {
		title, kind, leader string
	}{
		{"Guest Sensei · Randori drills", "drill", "Shihan H. Yamamoto"},
		{"Free sparring rotation · Mat 3", "sparring", "Kenji Sato"},
		{"Kata clinic with visiting Shihan", "seminar", "Aiko Tanaka"},
	} {
		add(`INSERT OR IGNORE INTO training_sessions
			(id, event_id, title, kind, leader, starts_at, duration_seconds)
			VALUES (?, 'ev-demo', ?, ?, ?, ?, ?)`,
			fmt.Sprintf("session-%02d", i+1), session.title, session.kind, session.leader,
			fmt.Sprintf("2026-11-19T%02d:00:00Z", 13+i), 1800)
	}

	add(`INSERT OR IGNORE INTO invitations
		(id, tenant_id, event_id, club_id, role, email, token_hash, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"inv-1", tenantID, "ev-demo", clubs[0].id, "fighter", "new@demo.j3", "demo-token-1", "2026-12-31T00:00:00Z")
	add(`INSERT OR IGNORE INTO audit_log
		(tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tenantID, "u-runner", "tournament_runner", "demo.seed", "event", "ev-demo", `{"status":"published"}`)

	for _, s := range stmts {
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

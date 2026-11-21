// Package seed's Okami tournament: a complete, realistic event configuration.
package seed

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/db"
)

const okamiEvent = "ev-okami"

// EnsureOkami creates the "Okami Jiu Jitsu Challenger" demo tournament with
// all people, clubs, categories, rules, pools, schedule and display data.
func EnsureOkami(store *db.DB) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []struct {
		query string
		args  []any
	}{
		{`INSERT OR IGNORE INTO tenants (id, name, slug) VALUES ('demo-tenant','J3 Demo','demo')`, nil},
		{`INSERT OR IGNORE INTO themes (id, tenant_id, name, tokens_json)
		  VALUES ('okami-theme','demo-tenant','Okami', ?)`,
			[]any{`{"primary":"#b91c1c","accent":"#fbbf24","ink":"#f8fafc","style":"classic"}`}},
		{`INSERT OR IGNORE INTO events
		  (id, tenant_id, theme_id, name, starts_at, ends_at, status)
		  VALUES (?, 'demo-tenant', 'okami-theme', ?, '2026-11-21T09:15:00Z','2026-11-21T21:00:00Z','published')`,
			[]any{okamiEvent, "Okami Jiu Jitsu Challenger"}},
		{`INSERT OR IGNORE INTO event_info
		  (event_id, organiser, venue, costs_json, refreshments, parking, hotel_json, evening_json, course_json, attendance_rule)
		  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []any{
			okamiEvent, "Tony Jamieson", "Skegness Grammar School Gymnasium, Skegness",
			`[{"label":"Student","amount":30,"currency":"GBP"},{"label":"Spectator","amount":5,"currency":"GBP"}]`,
			"Students must bring a packed lunch. No fizzy pop allowed.",
			"Ample car parking available on site at the school.",
			`[{"name":"The Vine Hotel","phone":"01754 610611"}]`,
			`{"time":"19:00","venue":"The Vine Hotel event room","note":"Evening event with food and drinks"}`,
			`{"required":true,"start":"09:30","groups":["11-17 coloured belts","18+ coloured belts","Junior black belts","Senior black belts","6-10 dojo fitness challenges"]}`,
			"All students must attend the morning course to compete. Leaving the mats early or arriving after the course means exclusion from the tournament.",
		}},
	}
	add := func(query string, args ...any) {
		stmts = append(stmts, struct {
			query string
			args  []any
		}{query, args})
	}

	// Full belt spectrum with correct ordering for minimum-grade categories.
	for i, rank := range []struct{ id, label, colour, kind string }{
		{"br-white", "White", "#f8fafc", "kyu"},
		{"br-yellow", "Yellow", "#facc15", "kyu"},
		{"br-orange", "Orange", "#fb923c", "kyu"},
		{"br-green", "Green", "#22c55e", "kyu"},
		{"br-blue", "Blue", "#3b82f6", "kyu"},
		{"br-purple", "Purple", "#a855f7", "kyu"},
		{"br-brown", "Brown", "#92400e", "kyu"},
		{"br-black", "Black", "#0f172a", "dan"},
	} {
		add(`INSERT OR IGNORE INTO belt_ranks
			(id, tenant_id, label, colour, sort_order, grade_kind)
			VALUES (?, 'demo-tenant', ?, ?, ?, ?)`,
			rank.id, rank.label, rank.colour, i+1, rank.kind)
	}

	// Clubs, including Okami Jiu Jitsu.
	for _, club := range []struct{ id, name, colour string }{
		{"okami", "Okami Jiu Jitsu", "#b91c1c"},
		{"club-b", "Skegness Dojo", "#2563eb"},
		{"club-c", "Lincoln Jiu Jitsu", "#16a34a"},
		{"club-d", "Coastal Martial Arts", "#7c3aed"},
	} {
		add(`INSERT OR IGNORE INTO clubs (id, tenant_id, name, primary_colour) VALUES (?, 'demo-tenant', ?, ?)`,
			club.id, club.name, club.colour)
	}

	// Organiser and staff: 1 runner, 6 judges (3 per mat), 2 scorekeepers, 8 volunteers.
	add(`INSERT OR IGNORE INTO users (id, tenant_id, club_id, role, display_name, email, status)
	     VALUES ('u-tony','demo-tenant','okami','tournament_runner','Tony Jamieson','tony@okami.j3','active')`)
	for i := 1; i <= 6; i++ {
		add(`INSERT OR IGNORE INTO users (id, tenant_id, role, display_name, email, status)
		     VALUES (?, 'demo-tenant','official',?,?,'active')`,
			fmt.Sprintf("u-judge-%d", i), fmt.Sprintf("Judge %d", i), fmt.Sprintf("judge%d@okami.j3", i))
	}
	for i := 1; i <= 2; i++ {
		add(`INSERT OR IGNORE INTO users (id, tenant_id, role, display_name, email, status)
		     VALUES (?, 'demo-tenant','official',?,?,'active')`,
			fmt.Sprintf("u-keeper-%d", i), fmt.Sprintf("keeper%d@okami.j3", i), fmt.Sprintf("Scorekeeper Mat %d", i))
	}
	volunteerRoles := []string{"Finding participants", "Belt tying", "Directing attendees", "Awarding medals", "Mat assistant", "Course assistant", "Weigh-in helper", "General duties"}
	for i, duty := range volunteerRoles {
		add(`INSERT OR IGNORE INTO users (id, tenant_id, role, display_name, email, status)
		     VALUES (?, 'demo-tenant','official',?,?,'active')`,
			fmt.Sprintf("u-volunteer-%d", i+1), fmt.Sprintf("volunteer%d@okami.j3", i+1), fmt.Sprintf("Volunteer %d · %s", i+1, duty))
	}

	// 40 fighters across four clubs and four age bands, with believable belts.
	belts := []string{"br-white", "br-yellow", "br-orange", "br-green", "br-blue", "br-purple", "br-brown", "br-black"}
	clubs := []string{"okami", "club-b", "club-c", "club-d"}
	firstNames := []string{"Haruto", "Mika", "Aiden", "Yui", "Kai", "Ella", "Riku", "Sofia", "Noah", "Aiko", "Leo", "Nina", "Taro", "Amara", "Owen", "Mei", "JUnit", "Zara", "Hiro", "Lily", "Ravi", "Emma", "Sota", "Grace", "Ali", "Hana", "Theo", "Isla", "Ken", "Maya", "Finlay", "Rin", "Adam", "Suki", "Milo", "Freya", "Tobi", "Ivy", "Ethan", "Nao"}
	for i := 0; i < 40; i++ {
		id := fmt.Sprintf("ok-f%02d", i+1)
		name := fmt.Sprintf("%s %s", firstNames[i], []string{"Tanaka", "Smith", "Mori", "Brown", "Sato", "Wilson", "Khan", "Davis", "Ito", "Moore"}[i%10])
		club := clubs[i%4]
		band := "adult"
		age := 18 + i%12
		if i < 10 {
			band, age = "minor", 6+i
		}
		if i >= 10 && i < 20 {
			band, age = "teen", 11+(i-10)
		}
		if i >= 20 && i < 26 {
			band, age = "adolescent", 16+(i-20)
		}
		belt := belts[i%len(belts)]
		weight := 26 + float64(i)*2
		gender := "male"
		if i%3 == 0 {
			gender = "female"
		}
		add(`INSERT OR IGNORE INTO users (id, tenant_id, club_id, role, display_name, email, status)
		     VALUES (?, 'demo-tenant', ?, 'fighter', ?, ?, 'active')`,
			"ok-u"+id[2:], club, name, fmt.Sprintf("fighter%s@okami.j3", id[2:]))
		add(`INSERT OR IGNORE INTO fighters
		     (id, tenant_id, user_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender)
		     VALUES (?, 'demo-tenant', ?, ?, ?, ?, ?, ?, ?)`,
			id, "ok-u"+id[2:], club, belt, name,
			fmt.Sprintf("%d-05-01", 2026-age), weight, gender)
		_ = band
	}

	// Categories: age, weight and minimum belt grade.
	for i, category := range []struct {
		id, name, kind string
		min, max       float64
	}{
		{"ok-age-minor", "Minors 6-10", "age", 6, 10},
		{"ok-age-teen", "Teens 11-15", "age", 11, 15},
		{"ok-age-adol", "Adolescents 16-17", "age", 16, 17},
		{"ok-age-adult", "Adults 18+", "age", 18, 99},
		{"ok-weight-light", "Light -55kg", "weight", 0, 55},
		{"ok-weight-mid", "Middle 55-75kg", "weight", 55, 75},
		{"ok-weight-heavy", "Heavy +75kg", "weight", 75, 200},
		{"ok-belt-green", "Green belt and above", "belt", 5, 12},
	} {
		add(`INSERT OR IGNORE INTO categories (id, event_id, name, kind, min_value, max_value, sort_order)
		     VALUES (?, ?, ?, ?, ?, ?, ?)`,
			category.id, okamiEvent, category.name, category.kind, category.min, category.max, i)
	}

	// Kit requirements and mat staffing.
	for kitIndex, item := range []struct{ name, note string }{
		{"Clean pressed Gi", "Ironed and clean for every session"},
		{"Head Guard", "Required for sparring and Randori"},
		{"Hand and feet sparring pads", "MMA hybrid gloves and instep guards"},
		{"Groin guard", "Males; females not required to wear one"},
		{"Clean flip flops / slip-ons", "For all mat areas"},
		{"Gum shield", "Mandatory for all contact sessions"},
		{"Club licence book", "Must be presented at registration"},
	} {
		add(`INSERT OR IGNORE INTO kit_requirements (id, event_id, item, note) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("kit-%02d", kitIndex+1), okamiEvent, item.name, item.note)
	}
	for _, mat := range []int{1, 2} {
		add(`INSERT OR IGNORE INTO mats (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("ok-mat-%d", mat), okamiEvent, fmt.Sprintf("Mat %d", mat), mat)
		for judge := 0; judge < 3; judge++ {
			role := "support judge"
			if judge == 0 {
				role = "main judge"
			}
			add(`INSERT OR IGNORE INTO mat_staff (id, event_id, mat_name, role, user_id, sort_order)
			     VALUES (?, ?, ?, ?, ?, ?)`,
				fmt.Sprintf("staff-m%d-j%d", mat, judge), okamiEvent, fmt.Sprintf("Mat %d", mat), role,
				fmt.Sprintf("u-judge-%d", (mat-1)*3+judge+1), judge)
		}
		add(`INSERT OR IGNORE INTO mat_staff (id, event_id, mat_name, role, user_id, sort_order)
		     VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("staff-m%d-keeper", mat), okamiEvent, fmt.Sprintf("Mat %d", mat),
			"scorekeeper / timekeeper", fmt.Sprintf("u-keeper-%d", mat), 3)
	}

	// Disciplines with exact timings and point values from the organiser.
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-randori", okamiEvent, "Randori / Katamewaza",
		`{"ages":"6-9 and 10-14","holdSeconds":3,"maxHolds":3,"holdPoints":5,"refCountOutLoud":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-ground", okamiEvent, "Ground Fighting",
		`{"ages":"10-14","holdSeconds":3,"maxHolds":3,"holdPoints":5,"tapOrSubmission":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-cont10", okamiEvent, "Continuous Fighting 10-15",
		`{"roundSeconds":165,"phases":[{"name":"Sparring","seconds":60,"target":"hips to shoulders"},{"name":"Randori","seconds":45,"target":"first to be thrown"},{"name":"Ground","seconds":60,"target":"hold downs"}],"scoring":{"strike":1,"throwDefendedKick":3,"takedown":5,"hold":5},"fullEquipment":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-cont16", okamiEvent, "Continuous Fighting 16-17",
		`{"roundSeconds":165,"phases":[{"name":"Sparring","seconds":60,"target":"head to hips"},{"name":"Randori","seconds":45},{"name":"Ground","seconds":60}],"scoring":{"strike":1,"throwDefendedKick":3,"takedown":5,"hold":5},"mmaGloves":true,"locksAndChokes":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-cont18", okamiEvent, "Continuous Fighting 18+",
		`{"roundSeconds":165,"phases":[{"name":"Sparring","seconds":60},{"name":"Randori","seconds":45},{"name":"Ground","seconds":60}],"scoring":{"strike":1,"throwDefendedKick":3,"takedown":5,"hold":5},"mmaGloves":true,"locksAndChokes":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-spar", okamiEvent, "Sparring 16+",
		`{"roundSeconds":60,"lightContact":true,"target":"head to hip, no back, kicks shoulder to hip","scoring":{"strike":1,"throwDefendedKick":3},"headGuard":true}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-kata", okamiEvent, "Open Hand Kata",
		`{"grades":"any","maxSeconds":120,"panel":3,"sections":["Etiquette","Demonstration","Would the application work"],"maxPerSection":10}`)
	add(`INSERT OR IGNORE INTO discipline_rules (id, event_id, name, rules_json) VALUES (?, ?, ?, ?)`,
		"ok-disc-weapons", okamiEvent, "Weapons Kata",
		`{"minimumBelt":"green","maxSeconds":120,"demoOnly":true,"noSharpBlades":true,"panel":3,"sections":["Etiquette","Demonstration","Would the application work"],"maxPerSection":10}`)

	// Competition formats with fair progression and medal policies.
	for _, format := range []struct {
		discipline, format, knockout, third string
		dualBronze, retryTies               int
		orderPolicy                         string
	}{
		{"ok-disc-kata", "individual_panel", "none", "retry", 0, 1, "age_then_rank"},
		{"ok-disc-weapons", "individual_panel", "none", "retry", 0, 1, "age_then_rank"},
		{"ok-disc-randori", "round_robin", "none", "dual_bronze", 1, 1, "age_then_rank"},
		{"ok-disc-ground", "round_robin", "single_elimination", "rematch", 1, 1, "weight_then_age"},
		{"ok-disc-cont10", "round_robin", "single_elimination", "rematch", 1, 1, "weight_then_age"},
		{"ok-disc-cont16", "round_robin", "single_elimination", "rematch", 1, 1, "weight_then_age"},
		{"ok-disc-cont18", "round_robin", "single_elimination", "rematch", 1, 1, "weight_then_age"},
		{"ok-disc-spar", "round_robin", "single_elimination", "dual_bronze", 1, 1, "weight_then_age"},
	} {
		add(`INSERT OR IGNORE INTO discipline_formats
			(event_id, discipline_id, format, knockout, third_place, dual_bronze, retry_ties, order_policy)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			okamiEvent, format.discipline, format.format, format.knockout, format.third,
			format.dualBronze, format.retryTies, format.orderPolicy)
	}

	for i, section := range []string{"Etiquette", "Demonstration", "Would the application work"} {
		add(`INSERT OR IGNORE INTO kata_sections (id, event_id, name, max_score, sort_order) VALUES (?, ?, ?, 10, ?)`,
			fmt.Sprintf("ok-section-%d", i+1), okamiEvent, section, i)
	}

	// Schedule from the organiser.
	for i, item := range []struct {
		time, name, kind string
		duration         int
	}{
		{"09:15", "Registration", "general", 900},
		{"09:30", "Course (Drills and Training) — age 6+", "course", 9000},
		{"12:00", "Break / Lunch", "break", 1800},
		{"12:30", "Open Hand Kata — youngest to oldest, lowest to highest grade", "kata", 1800},
		{"13:00", "Weapons Kata — green belt and above, age order", "kata", 1800},
		{"13:30", "Junior Randori / Katamewaza (6-9) — round robin", "fight", 1800},
		{"14:00", "Junior Ground Fighting (10-14) — round robin into elimination", "fight", 1800},
		{"14:30", "Teens Ground Fighting & Continuous (10-15) — round robin into elimination", "fight", 1800},
		{"15:00", "Teens Continuous Fighting (16-17) — round robin into elimination", "fight", 1800},
		{"15:30", "Adults Continuous Fighting (18+) — round robin into elimination", "fight", 1800},
		{"19:00", "Evening event — The Vine Hotel", "social", 7200},
	} {
		add(`INSERT OR IGNORE INTO schedule_items
			(id, event_id, name, area, starts_at, duration_seconds, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("ok-sched-%02d", i+1), okamiEvent, item.name, "Main hall",
			fmt.Sprintf("2026-11-21T%s:00Z", item.time), item.duration, i)
	}

	// Fights with results, officials and scoring events.
	for i := 0; i < 18; i++ {
		fightID := fmt.Sprintf("ok-fight-%02d", i+1)
		red, blue := fmt.Sprintf("ok-f%02d", i+1), fmt.Sprintf("ok-f%02d", (i+1)%40+1)
		outcome := "win_red"
		if i%3 == 2 {
			outcome = "draw"
		}
		if i%4 == 3 {
			outcome = "win_blue"
		}
		add(`INSERT OR IGNORE INTO fights
			(id, event_id, mat_id, round_label, red_fighter_id, blue_fighter_id, state, outcome)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			fightID, okamiEvent, fmt.Sprintf("ok-mat-%d", i%2+1),
			[]string{"Randori", "Ground", "Continuous", "Sparring"}[i%4],
			red, blue, "complete", outcome)
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, 'referee')`,
			fightID, fmt.Sprintf("u-judge-%d", i%6+1))
		add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, payload_json) VALUES (?,1,'score',?)`,
			fightID, `{"corner":"red","points":5,"technique":"Hold down"}`)
		add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, payload_json) VALUES (?,2,'score',?)`,
			fightID, `{"corner":"blue","points":3,"technique":"Throw from kick"}`)
	}

	// Discipline participation: 20 kata, 40 continuous ground, 40 continuous,
	// and all minors/teens in Randori/Katamewaza.
	for i := 0; i < 40; i++ {
		fighterID := fmt.Sprintf("ok-f%02d", i+1)
		club := clubs[i%4]
		groups := []string{"ok-disc-ground", "ok-disc-cont"}
		if i < 10 {
			groups = append(groups, "ok-disc-randori")
		}
		if i < 20 {
			groups = append(groups, "ok-disc-kata")
		}
		if i >= 20 && i < 26 {
			groups = append(groups, "ok-disc-cont16")
		}
		if i >= 26 {
			groups = append(groups, "ok-disc-cont18", "ok-disc-spar")
		}
		if i >= 16 {
			groups = append(groups, "ok-disc-weapons")
		}
		for groupIndex, group := range groups {
			add(`INSERT OR IGNORE INTO entrants (id, event_id, fighter_id, category_id, club_id, seed)
			     VALUES (?, ?, ?, NULL, ?, ?)`,
				fmt.Sprintf("ok-ent-%s-%02d", group, i+1), okamiEvent, fighterID,
				club, groupIndex+1)
		}
	}

	// Kata performances with three judges × three sections.
	for i := 0; i < 10; i++ {
		performanceID := fmt.Sprintf("ok-kata-%02d", i+1)
		add(`INSERT OR IGNORE INTO kata_performances
			(id, event_id, fighter_id, kata_name, panel_size, round_number, kata_type, status, total)
			VALUES (?, ?, ?, ?, 3, 1, ?, 'complete', ?)`,
			performanceID, okamiEvent, fmt.Sprintf("ok-f%02d", i+1),
			[]string{"Hekigan-ken", "Seiyunchin", "Gekusai-dai", "Pinan sono ichi", "Kanku-dai"}[i%5],
			"tokuyi", float64(24+i%4))
		for judge := 1; judge <= 3; judge++ {
			for section := 1; section <= 3; section++ {
				add(`INSERT OR IGNORE INTO kata_section_scores (performance_id, section_id, judge, score)
				     VALUES (?, ?, ?, ?)`,
					performanceID, fmt.Sprintf("ok-section-%d", section), judge, 7+float64((i+judge+section)%4))
			}
		}
	}

	// Pools generated from the entrants in each discipline group.
	for group, label := range map[string]string{
		"randori": "Randori 6-9", "ground": "Ground 10-14", "cont": "Continuous Fighting", "spar": "Sparring 16+",
	} {
		for p := 0; p < 3; p++ {
			poolID := fmt.Sprintf("ok-pool-%s-%d", group, p+1)
			add(`INSERT OR IGNORE INTO pools (id, event_id, name, sort_order) VALUES (?, ?, ?, ?)`,
				poolID, okamiEvent, fmt.Sprintf("%s · Pool %c", label, 'A'+p), p)
			for m := 0; m < 8; m++ {
				add(`INSERT OR IGNORE INTO pool_members (pool_id, fighter_id, seed) VALUES (?, ?, ?)`,
					poolID, fmt.Sprintf("ok-f%02d", (p*8+m)%40+1), m+1)
			}
		}
	}

	// Course attendance: everyone attended except two excluded late arrivals.
	for i := 0; i < 40; i++ {
		attended := 1
		excluded := 0
		if i >= 38 {
			attended, excluded = 0, 1
		}
		add(`INSERT OR IGNORE INTO course_attendance (entrant_id, attended, excluded) VALUES (?, ?, ?)`,
			fmt.Sprintf("ok-ent-ok-disc-ground-%02d", i+1), attended, excluded)
	}

	// Medals and audit trail.
	for _, award := range []struct{ id, kind string }{
		{"ok-f01", "gold"}, {"ok-f02", "silver"}, {"ok-f03", "bronze"}, {"ok-f04", "bronze"},
	} {
		add(`INSERT OR IGNORE INTO medals (id, event_id, fighter_id, kind) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("ok-medal-%s-%s", award.kind, award.id), okamiEvent, award.id, award.kind)
	}
	add(`INSERT INTO audit_log (tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
	     VALUES ('demo-tenant','u-tony','tournament_runner','okami.seed','event',?,'{"seeded":true}')`, okamiEvent)

	for _, statement := range stmts {
		if _, err := tx.Exec(statement.query, statement.args...); err != nil {
			return fmt.Errorf("okami seed %q: %w", statement.query, err)
		}
	}
	// Ready-to-clone templates: a full example and three empty starting points.
	for _, template := range []struct{ id, name, config string }{
		{"tpl-okami", "Okami Challenger (full)",
			`{"name":"Okami Jiu Jitsu Challenger","disciplines":[{"name":"Open Hand Kata"}],"categories":[{"name":"Adults","kind":"age","min":18,"max":99}],"saveTemplate":true}`},
		{"tpl-club", "Club open day (empty)",
			`{"name":"Club Open Day","disciplines":[],"categories":[],"constraints":{"cap":50,"maxPools":8},"saveTemplate":true}`},
		{"tpl-kata", "Kata only (empty)",
			`{"name":"Kata Challenge","disciplines":[{"name":"Open Hand Kata"},{"name":"Weapons Kata"}],"categories":[],"saveTemplate":true}`},
		{"tpl-newhaza", "Newaza tournament (empty)",
			`{"name":"Newaza Tournament","disciplines":[{"name":"Randori / Katamewaza"},{"name":"Ground Fighting"}],"categories":[],"saveTemplate":true}`},
	} {
		add(`INSERT OR IGNORE INTO tournament_templates (id, tenant_id, name, config_json) VALUES (?, 'demo-tenant', ?, ?)`,
			template.id, template.name, template.config)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE users SET password_hash = ? WHERE tenant_id = 'demo-tenant' AND password_hash = ''`, string(passwordHash)); err != nil {
		return err
	}
	return tx.Commit()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

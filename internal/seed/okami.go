// Package seed's Okami tournament: a complete, realistic event configuration.
package seed

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/restful-otaku/j3/internal/awards"
	"github.com/restful-otaku/j3/internal/db"
)

const okamiEvent = "ev-okami"

// EnsureOkami creates the "Okami Jiu Jitsu Challengers Tournament 2027" demo
// tournament: a complete, mid-event fixture with live, paused, complete,
// forfeit and upcoming contests so RBAC and public views can be QA'd quickly.
// sqlOrNull keeps generated SQL expressions literal while leaving empty values
// as real NULLs. All inputs are seed-owned strings, never user supplied.
func sqlOrNull(value string) string {
	if value == "" {
		return "NULL"
	}
	return value
}

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
		{`INSERT OR IGNORE INTO themes (id, tenant_id, name, tokens_json, logo_path, banner_path)
		  VALUES ('okami-theme','demo-tenant','Okami', ?, ?, ?)`,
			[]any{`{"primary":"#b91c1c","accent":"#fbbf24","ink":"#f8fafc","style":"classic"}`, "/media/logo.png", "/media/banner.png"}},
		{`INSERT OR IGNORE INTO rule_sets (id, tenant_id, name, version, config_json)
		  VALUES ('rs-okami', 'demo-tenant', 'Okami 2027 Rules', 1, ?)`,
			[]any{`{"venue":"Skegness Grammar School Gymnasium, Skegness","protection":["Clean pressed Gi","Belt","Mouth guard","Head guard for youth divisions","Instep / shin guards for Randori"],` + cornerDefaultsJSON + `,"matchSeconds":300}`}},
		{`INSERT OR IGNORE INTO events
		  (id, tenant_id, theme_id, rule_set_id, name, starts_at, ends_at, status)
		  VALUES (?, 'demo-tenant', 'okami-theme', 'rs-okami', ?, '2027-11-20T09:15:00Z','2027-11-20T21:00:00Z','published')`,
			[]any{okamiEvent, "Okami Jiu Jitsu Challengers Tournament 2027"}},
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
	for i, club := range []struct{ id, name, colour string }{
		{"okami", "Okami Jiu Jitsu", "#b91c1c"},
		{"club-b", "Skegness Dojo", "#2563eb"},
		{"club-c", "Lincoln Jiu Jitsu", "#16a34a"},
		{"club-d", "Coastal Martial Arts", "#7c3aed"},
	} {
		add(`INSERT OR IGNORE INTO clubs (id, tenant_id, name, logo_path, primary_colour) VALUES (?, 'demo-tenant', ?, ?, ?)`,
			club.id, club.name, fmt.Sprintf("/media/club_logo_%d.jpg", i+1), club.colour)
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
			fmt.Sprintf("u-keeper-%d", i), fmt.Sprintf("Scorekeeper Mat %d", i), fmt.Sprintf("keeper%d@okami.j3", i))
	}
	volunteerRoles := []string{"Finding participants", "Belt tying", "Directing attendees", "Awarding medals", "Mat assistant", "Course assistant", "Weigh-in helper", "General duties"}
	for i, duty := range volunteerRoles {
		add(`INSERT OR IGNORE INTO users (id, tenant_id, role, display_name, email, status)
		     VALUES (?, 'demo-tenant','official',?,?,'active')`,
			fmt.Sprintf("u-volunteer-%d", i+1), fmt.Sprintf("Volunteer %d · %s", i+1, duty), fmt.Sprintf("volunteer%d@okami.j3", i+1))
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
		     (id, tenant_id, user_id, club_id, belt_rank_id, display_name, date_of_birth, weight_kg, gender, avatar_path)
		     VALUES (?, 'demo-tenant', ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, "ok-u"+id[2:], club, belt, name,
			fmt.Sprintf("%d-05-01", 2027-age), weight, gender,
			fmt.Sprintf("/media/fighter_%d.jpg", (i%6)+1))
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
		{"Gi", "Clean pressed Gi required for fighting disciplines"},
		{"Belt", "Correctly tied; no adjustments during contest"},
		{"Mouth guard", "Mandatory for all contact sessions"},
		{"Shin guards", "Instep / shin guards for Randori and sparring"},
		{"Groin protection", "Males; females not required"},
		{"Head guard", "Required for sparring and Randori; youth protection"},
		{"Course booklet", "Provided at registration for morning course"},
		{"Flip flops", "Clean slip-ons for mat areas"},
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

	// The fixture is deliberately mid-event: opening items are complete, the
	// afternoon fighting block is running and the evening social is upcoming.
	for i, item := range []struct {
		time, name, kind, status string
		duration                 int
	}{
		{"09:15", "Registration", "general", "complete", 900},
		{"09:30", "Course (Drills and Training) — age 6+", "course", "complete", 9000},
		{"12:00", "Break / Lunch", "break", "complete", 1800},
		{"12:30", "Open Hand Kata — youngest to oldest, lowest to highest grade", "kata", "complete", 1800},
		{"13:00", "Weapons Kata — green belt and above, age order", "kata", "live", 1800},
		{"13:30", "Junior Randori / Katamewaza (6-9) — round robin", "fight", "live", 1800},
		{"14:00", "Junior Ground Fighting (10-14) — round robin into elimination", "fight", "upcoming", 1800},
		{"14:30", "Teens Ground Fighting & Continuous (10-15) — round robin into elimination", "fight", "upcoming", 1800},
		{"15:00", "Teens Continuous Fighting (16-17) — round robin into elimination", "fight", "upcoming", 1800},
		{"15:30", "Adults Continuous Fighting (18+) — round robin into elimination", "fight", "upcoming", 1800},
		{"19:00", "Evening event — The Vine Hotel", "social", "upcoming", 7200},
	} {
		add(`INSERT OR IGNORE INTO schedule_items
			(id, event_id, name, area, starts_at, duration_seconds, sort_order, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("ok-sched-%02d", i+1), okamiEvent, item.name, "Main hall",
			fmt.Sprintf("2027-11-20T%s:00Z", item.time), item.duration, i, item.status)
	}
	add(`INSERT OR IGNORE INTO schedule_state (event_id, offset_minutes) VALUES (?, 7)`, okamiEvent)

	// A deliberately mid-event fixture: completed bouts alongside live, paused,
	// technical and upcoming contests so every role and public board has
	// something realistic to QA.
	type scoreRow struct {
		kind, corner, technique, card string
		points                        int
	}
	type fightFixture struct {
		state, outcome string
		seconds        int
		running        bool
		events         []scoreRow
	}
	fixtures := []fightFixture{
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Ippon · seoi nage", "", 3}, {"score", "blue", "Waza-ari · koshi guruma", "", 1}}},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"score", "red", "Throw from kick", "", 2}, {"score", "blue", "Submission · cross armbar", "", 3}}},
		{state: "complete", outcome: "draw", events: []scoreRow{{"score", "red", "Waza-ari · tomoe nage", "", 1}, {"score", "blue", "Waza-ari · osae-komi", "", 1}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Hold down · kesa-gatame", "", 2}, {"penalty", "blue", "Shido · passivity", "shido", -1}}},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"score", "blue", "Sweep · harai-goshi", "", 2}, {"score", "red", "Palm strike", "", 1}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"submission", "red", "Triangle choke", "", 0}}},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"score", "blue", "Takedown · morote gari", "", 2}, {"penalty", "red", "Yellow card", "yellow", -1}}},
		{state: "complete", outcome: "draw", events: []scoreRow{{"score", "red", "Knee strike", "", 1}, {"score", "blue", "Throw · haraigoshi", "", 1}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Mount control", "", 2}, {"score", "blue", "Guard pass", "", 1}}},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"score", "blue", "Ippon · ippon seoi nage", "", 3}, {"score", "red", "Waza-ari", "", 1}}},
		{state: "live", outcome: "", seconds: 95, running: true, events: []scoreRow{{"score", "red", "Throw · seoi nage", "", 2}, {"score", "blue", "Palm strike", "", 1}}},
		{state: "paused", outcome: "", seconds: 47, events: []scoreRow{{"score", "blue", "Hold · yoko-shiho-gatame", "", 2}, {"penalty", "red", "Shido", "shido", -1}}},
		{state: "live", outcome: "", seconds: 142, running: true, events: []scoreRow{{"score", "red", "Reap · osoto-gari", "", 2}}},
		{state: "forfeit", outcome: "win_blue", events: []scoreRow{{"forfeit", "blue", "Technical result · Injury", "", 0}}},
		{state: "forfeit", outcome: "dq_red", events: []scoreRow{{"technical", "blue", "Technical result · Cheating / unsportsmanlike conduct", "red", 0}}},
		{state: "complete", outcome: "draw", events: []scoreRow{{"score", "red", "Throw · koshi-guruma", "", 2}, {"score", "blue", "Sweep · tomoe-nage", "", 2}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Hold down twenty seconds (Ippon)", "", 5}, {"score", "blue", "Kick · mawashi geri", "", 1}}},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"score", "blue", "Submission · triangle choke", "", 0}, {"score", "red", "Throw · seoi nage", "", 2}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Takedown · morote gari", "", 3}, {"score", "blue", "Shido · passivity", "shido", -1}}},
		{state: "pending", outcome: "", seconds: 180},
		{state: "complete", outcome: "win_blue", events: []scoreRow{{"submission", "blue", "Submission · rear naked choke", "", 0}}},
		{state: "complete", outcome: "win_red", events: []scoreRow{{"score", "red", "Ippon · ura-nage", "", 3}}},
		{state: "live", outcome: "", seconds: 61, running: true, events: []scoreRow{{"score", "red", "Kick · mawashi geri", "", 1}, {"score", "blue", "Takedown · kata-guruma", "", 2}}},
		{state: "paused", outcome: "", seconds: 24, events: []scoreRow{{"penalty", "blue", "Yellow card", "yellow", -1}}},
	}
	fightDisciplines := []string{
		"ok-disc-randori", "ok-disc-ground", "ok-disc-cont10", "ok-disc-cont16",
		"ok-disc-cont18", "ok-disc-spar", "ok-disc-kata", "ok-disc-weapons",
	}
	fightCategories := []string{
		"ok-age-minor", "ok-weight-light", "ok-age-teen", "ok-age-adult",
		"ok-weight-light", "ok-weight-mid", "ok-age-teen", "ok-age-adult",
	}
	for i, fixture := range fixtures {
		fightID := fmt.Sprintf("ok-fight-%02d", i+1)
		red, blue := fmt.Sprintf("ok-f%02d", i+1), fmt.Sprintf("ok-f%02d", (i+1)%40+1)
		mat := fmt.Sprintf("ok-mat-%d", i%2+1)
		completed := ""
		if fixture.state == "complete" || fixture.state == "forfeit" {
			completed = fmt.Sprintf("strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ','now','-%d minutes')", len(fixtures)-i)
		}
		add(`INSERT OR IGNORE INTO fights
			(id, event_id, mat_id, category_id, discipline_id, round_label, red_fighter_id, blue_fighter_id, state, outcome, completed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+sqlOrNull(completed)+`)`,
			fightID, okamiEvent, mat, fightCategories[i%len(fightCategories)],
			fightDisciplines[i%len(fightDisciplines)],
			[]string{"Randori", "Ground", "Continuous", "Sparring"}[i%4],
			red, blue, fixture.state, fixture.outcome)
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, 'referee')`,
			fightID, fmt.Sprintf("u-judge-%d", i%6+1))
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, 'judge')`,
			fightID, fmt.Sprintf("u-judge-%d", (i+2)%6+1))
		add(`INSERT OR IGNORE INTO fight_officials (fight_id, user_id, role) VALUES (?, ?, 'scorekeeper')`,
			fightID, fmt.Sprintf("u-keeper-%d", i%2+1))
		if fixture.state != "pending" {
			endsAt := "''"
			if fixture.running {
				endsAt = fmt.Sprintf("strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ','now','+%d seconds')", fixture.seconds)
			}
			running := 0
			if fixture.running {
				running = 1
			}
			add(`INSERT OR IGNORE INTO fight_state
				(fight_id, seconds_remaining, running, round, ends_at)
				SELECT ?, ?, ?, ?, `+endsAt+`
				WHERE NOT EXISTS (SELECT 1 FROM fight_state WHERE fight_id = ?)`,
				fightID, fixture.seconds, running, 1+i%3, fightID)
		}
		for seq, event := range fixture.events {
			add(`INSERT OR IGNORE INTO score_events
				(fight_id, seq, kind, corner, points, technique, card, actor_user_id, payload_json)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				fightID, seq+1, event.kind, event.corner, event.points, event.technique, event.card,
				fmt.Sprintf("u-keeper-%d", i%2+1), `{"source":"seeded-mid-event"}`)
		}
		if i < 10 {
			add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, payload_json) VALUES (?,1,'score',?)`,
				fightID, `{"corner":"red","points":3,"technique":"Ippon · seoi nage"}`)
			add(`INSERT OR IGNORE INTO fight_events (fight_id, seq, kind, payload_json) VALUES (?,2,'score',?)`,
				fightID, `{"corner":"blue","points":1,"technique":"Waza-ari · osae-komi"}`)
		}
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

	// Kata: first round complete, weapons round partly judged and the last
	// competitors still waiting. Three judges score each of three sections.
	for i := 0; i < 10; i++ {
		performanceID := fmt.Sprintf("ok-kata-%02d", i+1)
		status, round := "complete", 1
		switch {
		case i >= 8:
			status, round = "pending", 2
		case i >= 6:
			status, round = "live", 2
		}
		kataCategory := []string{"ok-age-minor", "ok-age-teen", "ok-age-adult"}[i%3]
		kataDiscipline := "ok-disc-kata"
		if round > 1 {
			kataDiscipline = "ok-disc-weapons"
		}
		add(`INSERT OR IGNORE INTO kata_performances
			(id, event_id, fighter_id, kata_name, panel_size, round_number, kata_type, status, total, category_id, discipline_id)
			VALUES (?, ?, ?, ?, 3, ?, ?, ?, ?, ?, ?)`,
			performanceID, okamiEvent, fmt.Sprintf("ok-f%02d", i+1),
			[]string{"Hekigan-ken", "Seiyunchin", "Gekusai-dai", "Pinan sono ichi", "Kanku-dai"}[i%5],
			round, "tokuyi", status, float64(24+i%4), kataCategory, kataDiscipline)
		for judge := 1; judge <= 3; judge++ {
			for section := 1; section <= 3; section++ {
				if status == "pending" || (status == "live" && section > 1) {
					continue
				}
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

	// Medals for completed fight and kata blocks are derived below from every
	// category/discipline/event-type podium (including dual-bronze policies).
	// Idempotent audit trail entries describing the mid-event state.
	for _, entry := range []struct{ action, target, actor, role, after string }{
		{"okami.seed", okamiEvent, "u-tony", "tournament_runner", `{"seeded":true,"year":2027}`},
		{"fight.start", "ok-fight-11", "u-keeper-1", "official", `{"running":true,"mat":"Mat 1"}`},
		{"fight.penalty", "ok-fight-12", "u-keeper-2", "official", `{"card":"shido","corner":"red"}`},
		{"fight.technical", "ok-fight-14", "u-keeper-2", "official", `{"outcome":"win_blue","reason":"injury"}`},
		{"fight.technical", "ok-fight-15", "u-keeper-1", "official", `{"outcome":"dq_red","reason":"cheating"}`},
		{"schedule.drift", okamiEvent, "u-tony", "tournament_runner", `{"offsetMinutes":7}`},
	} {
		add(`INSERT INTO audit_log
			(tenant_id, actor_user_id, actor_role, action, target_type, target_id, after_json)
			SELECT 'demo-tenant', ?, ?, ?, 'fight', ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM audit_log WHERE target_id = ? AND action = ?)`,
			entry.actor, entry.role, entry.action, entry.target, entry.after, entry.target, entry.action)
	}

	// Ready-to-clone templates: a full example and three empty starting points.
	for _, template := range []struct{ id, name, config string }{
		{"tpl-okami", "Okami Challengers 2027 (full)",
			`{"name":"Okami Jiu Jitsu Challengers Tournament 2027","disciplines":[{"name":"Open Hand Kata"}],"categories":[{"name":"Adults","kind":"age","min":18,"max":99}],"saveTemplate":true}`},
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

	for _, statement := range stmts {
		if _, err := tx.Exec(statement.query, statement.args...); err != nil {
			return fmt.Errorf("okami seed %q: %w", statement.query, err)
		}
	}
	if _, err := awards.ReconcileTx(context.Background(), tx, okamiEvent); err != nil {
		return fmt.Errorf("okami seed medals: %w", err)
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

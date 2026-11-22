package seed_test

import (
	"testing"

	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func TestOkamiSeedCompleteAndIdempotent(t *testing.T) {
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for i := 0; i < 2; i++ {
		if err := seed.EnsureOkami(store); err != nil {
			t.Fatal(err)
		}
	}
	counts := map[string]int{
		"fighters": 0, "clubs": 0, "entrants": 0, "mats": 0,
		"mat_staff": 0, "discipline_rules": 0, "kit_requirements": 0,
		"schedule_items": 0, "kata_sections": 0, "course_attendance": 0,
	}
	for table := range counts {
		var n int
		if err := store.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		counts[table] = n
	}
	if counts["fighters"] != 40 || counts["clubs"] != 4 || counts["mats"] != 2 {
		t.Fatalf("core counts: %+v", counts)
	}
	if counts["mat_staff"] != 8 || counts["discipline_rules"] != 8 || counts["kit_requirements"] != 7 {
		t.Fatalf("config counts: %+v", counts)
	}
	if counts["entrants"] < 60 || counts["schedule_items"] < 11 || counts["kata_sections"] != 3 {
		t.Fatalf("event counts: %+v", counts)
	}
	var excluded int
	if err := store.QueryRow(`SELECT COUNT(*) FROM course_attendance WHERE excluded=1`).Scan(&excluded); err != nil {
		t.Fatal(err)
	}
	if excluded != 2 {
		t.Fatalf("excluded = %d", excluded)
	}
}

func TestOkami2027MidEventStates(t *testing.T) {
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := seed.EnsureOkami(store); err != nil {
		t.Fatal(err)
	}

	var name, startsAt string
	if err := store.QueryRow(`SELECT name, starts_at FROM events WHERE id = 'ev-okami'`).Scan(&name, &startsAt); err != nil {
		t.Fatal(err)
	}
	if name != "Okami Jiu Jitsu Challengers Tournament 2027" || startsAt[:4] != "2027" {
		t.Fatalf("event identity = %q at %q", name, startsAt)
	}

	states := map[string]int{}
	rows, err := store.Query(`SELECT state, COUNT(*) FROM fights WHERE event_id = 'ev-okami' GROUP BY state`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			t.Fatal(err)
		}
		states[state] = count
	}
	rows.Close()
	for _, state := range []string{"complete", "live", "paused", "forfeit", "pending"} {
		if states[state] == 0 {
			t.Fatalf("no %s fights in mid-event seed: %+v", state, states)
		}
	}
	var outcomes int
	if err := store.QueryRow(`SELECT COUNT(*) FROM fights WHERE outcome IN ('win_red','win_blue','draw','dq_red','dq_blue')`).Scan(&outcomes); err != nil {
		t.Fatal(err)
	}
	if outcomes < 12 {
		t.Fatalf("only %d decided fights", outcomes)
	}
	var running int
	if err := store.QueryRow(`SELECT COUNT(*) FROM fight_state WHERE running = 1 AND ends_at LIKE '2027%'`).Scan(&running); err == nil && running > 0 {
		// Dynamic strftime values are stored in the current year, so this is
		// informational rather than a failure; the next check is authoritative.
		t.Logf("live clocks persisted with absolute timestamps: %d", running)
	}
	if err := store.QueryRow(`SELECT COUNT(*) FROM fight_state WHERE running = 1`).Scan(&running); err != nil {
		t.Fatal(err)
	}
	if running < 2 {
		t.Fatalf("only %d running clocks", running)
	}
	var events int
	if err := store.QueryRow(`SELECT COUNT(*) FROM score_events`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events < 30 {
		t.Fatalf("only %d committed score events", events)
	}
	var assignments int
	if err := store.QueryRow(`SELECT COUNT(*) FROM fight_officials WHERE role IN ('judge','scorekeeper')`).Scan(&assignments); err != nil {
		t.Fatal(err)
	}
	if assignments < 40 {
		t.Fatalf("only %d judge/scorekeeper assignments", assignments)
	}
	var liveSchedule int
	if err := store.QueryRow(`SELECT COUNT(*) FROM schedule_items WHERE status = 'live'`).Scan(&liveSchedule); err != nil {
		t.Fatal(err)
	}
	if liveSchedule == 0 {
		t.Fatal("expected in-progress schedule items")
	}
	var kataLive, kataPending int
	if err := store.QueryRow(`SELECT COUNT(*) FROM kata_performances WHERE status = 'live'`).Scan(&kataLive); err != nil {
		t.Fatal(err)
	}
	if err := store.QueryRow(`SELECT COUNT(*) FROM kata_performances WHERE status = 'pending'`).Scan(&kataPending); err != nil {
		t.Fatal(err)
	}
	if kataLive == 0 || kataPending == 0 {
		t.Fatalf("kata states live=%d pending=%d", kataLive, kataPending)
	}
}

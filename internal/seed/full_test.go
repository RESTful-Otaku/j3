package seed_test

import (
	"testing"

	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
)

func TestFullTournamentSeedIsCompleteAndIdempotent(t *testing.T) {
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := seed.Ensure(store); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureFull(store); err != nil {
		t.Fatal(err)
	}
	if err := seed.EnsureFull(store); err != nil {
		t.Fatal(err)
	}

	counts := map[string]int{
		"fights": 0, "bracket_stages": 0, "bracket_slots": 0,
		"pools": 0, "kata_performances": 0, "medals": 0,
	}
	for table := range counts {
		var n int
		if err := store.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		counts[table] = n
	}
	if counts["bracket_stages"] < 5 || counts["bracket_slots"] < 30 {
		t.Fatalf("bracket not populated: %+v", counts)
	}
	if counts["fights"] < 30 || counts["pools"] < 6 || counts["kata_performances"] < 6 || counts["medals"] < 4 {
		t.Fatalf("demo data incomplete: %+v", counts)
	}

	var bronze, kataMedals, groups int
	if err := store.QueryRow(`SELECT COUNT(*) FROM medals WHERE kind = 'bronze'`).Scan(&bronze); err != nil {
		t.Fatal(err)
	}
	if err := store.QueryRow(`SELECT COUNT(*) FROM medals WHERE event_type = 'kata'`).Scan(&kataMedals); err != nil {
		t.Fatal(err)
	}
	if err := store.QueryRow(`SELECT COUNT(DISTINCT category_id || '|' || discipline_id || '|' || event_type) FROM medals`).Scan(&groups); err != nil {
		t.Fatal(err)
	}
	if bronze < 4 || kataMedals < 3 || groups < 3 {
		t.Fatalf("expected medals across multiple divisions and kata, got bronze=%d kata=%d groups=%d", bronze, kataMedals, groups)
	}

	var promoted int
	if err := store.QueryRow(`SELECT COUNT(*) FROM bracket_slots WHERE state = 'advanced'`).Scan(&promoted); err != nil {
		t.Fatal(err)
	}
	if promoted < 10 {
		t.Fatalf("promotion history missing: %d", promoted)
	}
}

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

package techniques

import (
	"strings"
	"testing"
)

func TestCatalogueSizeIDsAndShape(t *testing.T) {
	all := All()
	if len(all) < 300 {
		t.Fatalf("expected at least 300 techniques, got %d", len(all))
	}

	seen := make(map[string]bool, len(all))
	groups := make(map[string]bool)
	for _, tech := range all {
		if tech.ID == "" {
			t.Fatal("found an empty technique ID")
		}
		if seen[tech.ID] {
			t.Errorf("duplicate technique ID %q", tech.ID)
		}
		seen[tech.ID] = true
		if tech.ID != strings.ToLower(tech.ID) || strings.ContainsAny(tech.ID, " -/.") {
			t.Errorf("ID %q is not lowercase snake_case", tech.ID)
		}
		if tech.English == "" || tech.Romaji == "" {
			t.Errorf("technique %q is missing an English or romaji name", tech.ID)
		}
		if tech.Group == "" || tech.PointKey == "" {
			t.Errorf("technique %q is missing a group or point key", tech.ID)
		}
		switch tech.Kind {
		case "score", "penalty", "submission":
		default:
			t.Errorf("technique %q has invalid kind %q", tech.ID, tech.Kind)
		}
		if tech.Kind == "penalty" {
			switch tech.Card {
			case "yellow", "red", "shido", "hansoku", "":
			default:
				t.Errorf("penalty %q has invalid card %q", tech.ID, tech.Card)
			}
		}
		groups[tech.Group] = true
	}

	for _, group := range Groups() {
		if !groups[group.ID] {
			t.Errorf("group %q has no techniques", group.ID)
		}
		if group.ID == "" || group.Label == "" || group.PointKey == "" {
			t.Errorf("group %+v is missing an id, label or point key", group)
		}
	}
}

func TestAllIsStableAndDefensive(t *testing.T) {
	first := All()
	second := All()
	if len(first) != len(second) {
		t.Fatalf("All() changed length: %d then %d", len(first), len(second))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("catalogue order changed at %d: %q then %q", i, first[i].ID, second[i].ID)
		}
	}

	first[0].English = "mutated"
	if second[0].English == "mutated" {
		t.Fatal("All() returned a shared Technique value")
	}
	for _, tech := range first {
		if len(tech.Tags) == 0 {
			continue
		}
		tech.Tags[0] = "mutated tag"
		break
	}
	for _, tech := range All() {
		for _, tag := range tech.Tags {
			if tag == "mutated tag" {
				t.Fatal("All() returned shared tag slices")
			}
		}
	}
}

func TestByID(t *testing.T) {
	want, ok := ByID("seoi_nage")
	if !ok || want.English != "Shoulder throw" {
		t.Fatalf("ByID(seoi_nage) = %+v, %v", want, ok)
	}
	if got, ok := ByID("SEOI_NAGE"); !ok || got.ID != "seoi_nage" {
		t.Fatalf("ByID is not case-insensitive: %+v, %v", got, ok)
	}
	if _, ok := ByID("definitely_not_a_technique"); ok {
		t.Fatal("ByID found a missing technique")
	}
	if _, ok := ByID(""); ok {
		t.Fatal("ByID found an empty ID")
	}

	// The returned value must not let callers mutate the catalogue.
	want.Tags = append(want.Tags, "extra")
	again, _ := ByID("seoi_nage")
	for _, tag := range again.Tags {
		if tag == "extra" {
			t.Fatal("ByID returned a shared tag slice")
		}
	}
}

func TestSearchByNameRomajiAndTags(t *testing.T) {
	byEnglish := Search("shoulder throw", "", 0)
	if len(byEnglish) == 0 || byEnglish[0].ID != "seoi_nage" {
		t.Fatalf("English search did not find seoi_nage: %+v", byEnglish)
	}
	byRomaji := Search("SEOI", "", 0)
	found := false
	for _, tech := range byRomaji {
		if tech.ID == "seoi_nage" {
			found = true
		}
	}
	if !found {
		t.Fatalf("case-insensitive romaji search missed seoi_nage: %+v", byRomaji)
	}
	byJapanese := Search("背負い投げ", "", 0)
	if len(byJapanese) == 0 {
		t.Fatal("Japanese script search returned nothing")
	}
	byTag := Search("kyusho", "", 0)
	if len(byTag) < 5 {
		t.Fatalf("tag search returned only %d results", len(byTag))
	}
	if got := Search("no-such-technique-in-catalogue", "", 0); len(got) != 0 {
		t.Fatalf("unmatched search returned %d results", len(got))
	}
}

func TestSearchGroupFilterAndLimit(t *testing.T) {
	throws := Search("", "throws", 0)
	if len(throws) == 0 {
		t.Fatal("group filter returned no throws")
	}
	for _, tech := range throws {
		if tech.Group != GroupThrows {
			t.Errorf("group filter returned %q from %q", tech.ID, tech.Group)
		}
	}
	byLabel := Search("", "Throws", 0)
	if len(byLabel) != len(throws) {
		t.Fatalf("group label filter returned %d, id filter returned %d", len(byLabel), len(throws))
	}
	if got := Search("", "definitely-not-a-group", 0); len(got) != 0 {
		t.Fatalf("unknown group returned %d results", len(got))
	}
	if got := Search("", "", 5); len(got) != 5 {
		t.Fatalf("limit 5 returned %d results", len(got))
	}
	if got := Search("", "", 0); len(got) != len(All()) {
		t.Fatalf("unlimited empty query returned %d of %d", len(got), len(All()))
	}
	if got := Search("", "", -3); len(got) != len(All()) {
		t.Fatalf("negative limit returned %d of %d", len(got), len(All()))
	}
}

func TestEveryGroupReturnsResults(t *testing.T) {
	for _, group := range Groups() {
		results := Search("", group.ID, 0)
		if len(results) == 0 {
			t.Errorf("group %q returned no results", group.ID)
			continue
		}
		matchesGroup := false
		for _, tech := range results {
			if tech.Group == group.ID {
				matchesGroup = true
			}
		}
		if !matchesGroup {
			t.Errorf("group %q filter returned unrelated techniques", group.ID)
		}
	}
}

func TestInstantWinSubmissions(t *testing.T) {
	instant := 0
	for _, tech := range All() {
		if tech.InstantWin && tech.Kind == "submission" {
			instant++
		}
	}
	if instant < 20 {
		t.Fatalf("expected at least 20 instant-win submissions, got %d", instant)
	}
	for _, id := range []string{
		"armbar", "triangle_choke", "rear_naked_choke", "heel_hook", "kneebar",
		"kimura", "americana", "guillotine_choke", "darce_choke", "anaconda_choke",
		"omoplata", "inside_heel_hook", "outside_heel_hook",
	} {
		tech, ok := ByID(id)
		if !ok {
			t.Errorf("missing instant-win technique %q", id)
			continue
		}
		if !tech.InstantWin {
			t.Errorf("technique %q is not marked instant win", id)
		}
	}
}

func TestPenaltyCards(t *testing.T) {
	cards := map[string]bool{}
	penalties := 0
	for _, tech := range All() {
		if tech.Group != GroupPenalties || tech.Kind != "penalty" {
			continue
		}
		penalties++
		if tech.Card != "" {
			cards[tech.Card] = true
		}
	}
	if penalties < 10 {
		t.Fatalf("expected at least 10 penalties, got %d", penalties)
	}
	for _, card := range []string{"yellow", "red", "shido", "hansoku"} {
		if !cards[card] {
			t.Errorf("no penalty entry with card %q", card)
		}
	}
	for _, id := range []string{"yellow_card", "red_card", "shido_penalty", "hansoku_make", "forfeit", "cheating"} {
		if _, ok := ByID(id); !ok {
			t.Errorf("missing penalty %q", id)
		}
	}
}

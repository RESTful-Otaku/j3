// Package awards derives a complete, deterministic medal set from completed
// fights and kata results. Medals are keyed per category, discipline and event
// type so the same fighter can win in more than one division.
package awards

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type category struct {
	id  string
	name string
}

type discipline struct {
	id, name, kind string
}

type standing struct {
	fighterID string
	contests  int
	wins      int
	points    int
	score     float64
}

type fightRow struct {
	id, categoryID, label, disciplineID, redID, blueID, outcome string
	pointsRed, pointsBlue                                        int
}

type kataRow struct {
	id, fighterID, name, kind, disciplineID, categoryID string
	round                                              int
	status                                             string
	total                                              float64
}

type slot struct {
	categoryID, disciplineID, eventType, fighterID, kind, key string
}

// Reconcile computes and upserts the full medal set for an event.
func Reconcile(ctx context.Context, db *sql.DB, eventID string) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	count, err := ReconcileTx(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	return count, tx.Commit()
}

// ReconcileTx is the transaction-aware variant used by seeders.
func ReconcileTx(ctx context.Context, tx *sql.Tx, eventID string) (int, error) {
	categories, err := loadCategories(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	disciplines, err := loadDisciplines(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	entrantCategory, err := loadEntrantCategories(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	dualBronze, err := loadBronzePolicy(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	fights, err := loadFights(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}
	kata, err := loadKata(ctx, tx, eventID)
	if err != nil {
		return 0, err
	}

	slots := []slot{}
	seenGroup := map[string]bool{}

	// Fighting and sparring disciplines: rank fighters by wins, then catalogue
	// points scored. Each category/discipline pair gets its own podium.
	groupFighters := map[string]map[string]*standing{}
	for _, fight := range fights {
		disciplineID, eventType := matchDiscipline(fight.label, fight.disciplineID, disciplines)
		categoryID := firstNonEmpty(fight.categoryID, entrantCategory[fight.redID], entrantCategory[fight.blueID], fallbackCategory(categories))
		groupKey := categoryID + "|" + disciplineID + "|" + eventType
		seenGroup[groupKey] = true
		if groupFighters[groupKey] == nil {
			groupFighters[groupKey] = map[string]*standing{}
		}
		table := groupFighters[groupKey]
		red := ensureStanding(table, fight.redID)
		blue := ensureStanding(table, fight.blueID)
		red.contests++
		blue.contests++
		red.points += fight.pointsRed
		blue.points += fight.pointsBlue
		switch fight.outcome {
		case "win_red", "dq_blue":
			red.wins++
		case "win_blue", "dq_red":
			blue.wins++
		}
	}
	for key, table := range groupFighters {
		parts := strings.Split(key, "|")
		bronze := 1
		if dualBronze[parts[1]] {
			bronze = 2
		}
		slots = append(slots, podiumSlots(parts[0], parts[1], parts[2], sortedStandings(table), bronze)...)
	}

	// Kata: final-round panel totals decide the podium for each kata discipline
	// and category. A fighter's entrant category is used when present; otherwise
	// categories are spread deterministically so every division is represented.
	kataGroups := map[string][]standing{}
	kataIndex := map[string]int{}
	for _, performance := range kata {
		name := performance.name
		if performance.kind != "" {
			name = performance.kind + " " + name
		}
		disciplineID := matchKataDiscipline(name, performance.disciplineID, disciplines)
		eventType := "kata"
		categoryID := firstNonEmpty(performance.categoryID, entrantCategory[performance.fighterID], fallbackKataCategory(categories, kataIndex, disciplineID))
		groupKey := categoryID + "|" + disciplineID + "|" + eventType
		seenGroup[groupKey] = true
		score := standing{fighterID: performance.fighterID, score: performance.total}
		kataGroups[groupKey] = append(kataGroups[groupKey], score)
	}
	for key, rows := range kataGroups {
		parts := strings.Split(key, "|")
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].score != rows[j].score {
				return rows[i].score > rows[j].score
			}
			return rows[i].fighterID < rows[j].fighterID
		})
		deduped := []standing{}
		used := map[string]bool{}
		for _, row := range rows {
			if !used[row.fighterID] {
				used[row.fighterID] = true
				deduped = append(deduped, row)
			}
		}
		bronze := 1
		if dualBronze[parts[1]] {
			bronze = 2
		}
		slots = append(slots, podiumSlots(parts[0], parts[1], parts[2], deduped, bronze)...)
	}

	// Publish one podium per configured discipline/category even when no result
	// has landed yet: the ceremony shows a labelled "to be decided" division.
	for _, category := range categories {
		for _, discipline := range disciplines {
			key := category.id + "|" + discipline.id + "|" + classify(discipline.name)
			if seenGroup[key] {
				continue
			}
			seenGroup[key] = true
			slots = append(slots, slot{categoryID: category.id, disciplineID: discipline.id, eventType: classify(discipline.name), kind: "pending"})
		}
	}

	for i := range slots {
		if slots[i].key == "" {
			slots[i].key = fmt.Sprintf("%s|%s|%s|pending|%d", slots[i].categoryID,
				slots[i].disciplineID, slots[i].eventType, i)
		}
		slots[i].key = eventID + "|" + slots[i].key
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].key < slots[j].key
	})

	// Drop derived rows that a previous reconciliation produced. Manual awards
	// are keyed separately and stay untouched, so organiser overrides survive.
	if _, err := tx.ExecContext(ctx, `DELETE FROM medals WHERE event_id = ? AND source = 'auto'`, eventID); err != nil {
		return 0, fmt.Errorf("clear derived medals: %w", err)
	}

	for _, award := range slots {
		if award.kind == "pending" {
			continue // placeholders are not database rows
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO medals
				(id, event_id, category_id, fighter_id, kind, discipline_id, event_type, award_key, source, awarded_at)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?, ?, 'auto', strftime('%Y-%m-%dT%H:%M:%fZ','now'))
			ON CONFLICT(event_id, award_key)
				WHERE award_key IS NOT NULL AND award_key <> ''
			DO UPDATE SET fighter_id = excluded.fighter_id, category_id = excluded.category_id,
				discipline_id = excluded.discipline_id, event_type = excluded.event_type, source = 'auto'`,
			"medal-"+award.key, eventID, nullOrEmpty(award.categoryID), award.fighterID, award.kind,
			nullOrEmpty(award.disciplineID), award.eventType, award.key)
		if err != nil {
			return 0, fmt.Errorf("upsert medal %s: %w", award.key, err)
		}
	}
	return len(slots), nil
}

func loadCategories(ctx context.Context, tx *sql.Tx, eventID string) ([]category, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, name FROM categories WHERE event_id = ? ORDER BY sort_order, id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []category
	for rows.Next() {
		var item category
		if err := rows.Scan(&item.id, &item.name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func loadDisciplines(ctx context.Context, tx *sql.Tx, eventID string) ([]discipline, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, kind FROM (
			SELECT id, name, kind, 0 AS sort FROM discipline_rules WHERE event_id = ?
			UNION ALL
			SELECT id, name, 'fight' AS kind, sort_order AS sort FROM disciplines
				WHERE event_id = ? AND id NOT IN (SELECT id FROM discipline_rules WHERE event_id = ?)
		) ORDER BY sort, name`, eventID, eventID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []discipline
	for rows.Next() {
		var item discipline
		if err := rows.Scan(&item.id, &item.name, &item.kind); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func loadEntrantCategories(ctx context.Context, tx *sql.Tx, eventID string) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT fighter_id, category_id FROM entrants
		WHERE event_id = ? AND category_id IS NOT NULL AND category_id <> ''
		GROUP BY fighter_id, category_id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var fighterID, categoryID string
		if err := rows.Scan(&fighterID, &categoryID); err != nil {
			return nil, err
		}
		if _, ok := out[fighterID]; !ok {
			out[fighterID] = categoryID
		}
	}
	return out, rows.Err()
}

func loadBronzePolicy(ctx context.Context, tx *sql.Tx, eventID string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT discipline_id, dual_bronze FROM discipline_formats WHERE event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var disciplineID string
		var dual int
		if err := rows.Scan(&disciplineID, &dual); err != nil {
			return nil, err
		}
		out[disciplineID] = dual == 1
	}
	return out, rows.Err()
}

func loadFights(ctx context.Context, tx *sql.Tx, eventID string) ([]fightRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT f.id, COALESCE(f.category_id, ''), COALESCE(f.discipline_id, ''), f.round_label,
		       COALESCE(f.red_fighter_id, ''), COALESCE(f.blue_fighter_id, ''), f.outcome,
		       COALESCE(SUM(CASE WHEN se.corner = 'red' THEN se.points ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN se.corner = 'blue' THEN se.points ELSE 0 END), 0)
		FROM fights f LEFT JOIN score_events se ON se.fight_id = f.id
		WHERE f.event_id = ? AND f.state IN ('complete', 'forfeit')
		GROUP BY f.id ORDER BY f.completed_at, f.id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []fightRow
	for rows.Next() {
		var item fightRow
		if err := rows.Scan(&item.id, &item.categoryID, &item.disciplineID, &item.label,
			&item.redID, &item.blueID, &item.outcome, &item.pointsRed, &item.pointsBlue); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func loadKata(ctx context.Context, tx *sql.Tx, eventID string) ([]kataRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, COALESCE(fighter_id, ''), COALESCE(kata_name, ''), COALESCE(kata_type, ''),
		       COALESCE(discipline_id, ''), COALESCE(category_id, ''), round_number, status, COALESCE(total, 0)
		FROM kata_performances WHERE event_id = ? AND COALESCE(total, 0) > 0
		ORDER BY round_number, id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []kataRow
	for rows.Next() {
		var item kataRow
		if err := rows.Scan(&item.id, &item.fighterID, &item.name, &item.kind,
			&item.disciplineID, &item.categoryID, &item.round, &item.status, &item.total); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func matchDiscipline(label, existing string, disciplines []discipline) (string, string) {
	for _, item := range disciplines {
		if item.id == existing {
			return item.id, classify(item.name)
		}
	}
	lower := strings.ToLower(label)
	for _, item := range disciplines {
		name := strings.ToLower(item.name)
		if lower != "" && (strings.Contains(name, lower) || strings.Contains(lower, name)) {
			return item.id, classify(item.name)
		}
	}
	for _, item := range disciplines {
		if classify(item.name) == classify(label) {
			return item.id, classify(item.name)
		}
	}
	return "auto:" + slug(label), classify(label)
}

func matchKataDiscipline(label, existing string, disciplines []discipline) string {
	for _, item := range disciplines {
		if item.id == existing && classify(item.name) == "kata" {
			return item.id
		}
	}
	lower := strings.ToLower(label)
	for _, item := range disciplines {
		name := strings.ToLower(item.name)
		if classify(item.name) == "kata" && (strings.Contains(name, lower) || strings.Contains(lower, "kata") || strings.Contains(lower, "weapon")) {
			return item.id
		}
	}
	for _, item := range disciplines {
		if classify(item.name) == "kata" {
			return item.id
		}
	}
	return "auto:kata-" + slug(label)
}

func classify(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "randori"), strings.Contains(lower, "katame"):
		return "randori"
	case strings.Contains(lower, "weapon"), strings.Contains(lower, "kata"):
		return "kata"
	case strings.Contains(lower, "ground"), strings.Contains(lower, "ne-waza"), strings.Contains(lower, "newaza"):
		return "ne_waza"
	case strings.Contains(lower, "continuous"):
		return "continuous"
	case strings.Contains(lower, "spar"):
		return "spar"
	default:
		return "fight"
	}
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out.WriteRune(r)
		default:
			out.WriteByte('-')
		}
	}
	result := strings.Trim(out.String(), "-")
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	if result == "" {
		return "general"
	}
	return result
}

func ensureStanding(table map[string]*standing, fighterID string) *standing {
	if fighterID == "" {
		return &standing{}
	}
	if table[fighterID] == nil {
		table[fighterID] = &standing{fighterID: fighterID}
	}
	return table[fighterID]
}

func sortedStandings(table map[string]*standing) []standing {
	out := make([]standing, 0, len(table))
	for _, row := range table {
		if row.fighterID != "" && row.contests > 0 {
			out = append(out, *row)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].wins != out[j].wins {
			return out[i].wins > out[j].wins
		}
		if out[i].points != out[j].points {
			return out[i].points > out[j].points
		}
		return out[i].fighterID < out[j].fighterID
	})
	return out
}

func podiumSlots(categoryID, disciplineID, eventType string, rows []standing, bronze int) []slot {
	out := []slot{}
	kinds := []string{"gold", "silver"}
	for i := 0; i < bronze; i++ {
		kinds = append(kinds, "bronze")
	}
	for i, kind := range kinds {
		if i >= len(rows) {
			break
		}
		out = append(out, slot{categoryID: categoryID, disciplineID: disciplineID,
			eventType: eventType, fighterID: rows[i].fighterID, kind: kind,
			key: fmt.Sprintf("%s|%s|%s|%s|%d", categoryID, disciplineID, eventType, kind, i)})
	}
	return out
}

func fallbackCategory(categories []category) string {
	if len(categories) == 0 {
		return ""
	}
	return categories[0].id
}

func fallbackKataCategory(categories []category, used map[string]int, disciplineID string) string {
	if len(categories) == 0 {
		return ""
	}
	index := used[disciplineID]
	used[disciplineID] = index + 1
	return categories[index%len(categories)].id
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func nullOrEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

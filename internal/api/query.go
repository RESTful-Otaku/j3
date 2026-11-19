package api

import (
	"net/http"
	"strconv"
	"strings"
)

type querySpec struct {
	columns []string
	search  []string
}

// Only these tables and columns may be queried. Values are always bound as
// parameters, so user input can never become SQL.
var querySpecs = map[string]querySpec{
	"fighters":   {[]string{"id", "display_name", "date_of_birth", "weight_kg", "gender", "club_id", "belt_rank_id"}, []string{"display_name"}},
	"clubs":      {[]string{"id", "name", "primary_colour", "logo_path"}, []string{"name"}},
	"events":     {[]string{"id", "name", "starts_at", "ends_at", "status"}, []string{"name"}},
	"categories": {[]string{"id", "event_id", "name", "kind", "gender", "min_value", "max_value", "sort_order"}, []string{"name"}},
	"fights":     {[]string{"id", "event_id", "mat_id", "category_id", "round_label", "red_fighter_id", "blue_fighter_id", "state", "outcome"}, []string{"id"}},
	"pools":      {[]string{"id", "event_id", "category_id", "name", "sort_order"}, []string{"name"}},
	"belt_ranks": {[]string{"id", "label", "colour", "sort_order", "grade_kind", "is_custom"}, []string{"label"}},
	"users":      {[]string{"id", "display_name", "role", "club_id", "status"}, []string{"display_name", "email"}},
	"audit_log":  {[]string{"id", "actor_user_id", "actor_role", "action", "target_type", "target_id", "created_at"}, []string{"action"}},
}

var reservedQueryParams = map[string]bool{
	"table": true, "search": true, "sort": true, "order": true,
	"limit": true, "offset": true,
}

// query exposes an allow-listed, sortable, filterable, searchable data source
// for custom views, charts and tables.
func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	table := r.URL.Query().Get("table")
	spec, ok := querySpecs[table]
	if !ok {
		writeBadRequest(w, "unknown or private table")
		return
	}

	where := []string{}
	args := []any{}
	params := r.URL.Query()
	for key, values := range params {
		if reservedQueryParams[key] || len(values) == 0 {
			continue
		}
		if contains(spec.columns, key) {
			where = append(where, key+" = ?")
			args = append(args, values[0])
		}
	}
	if search := strings.TrimSpace(params.Get("search")); search != "" {
		clauses := []string{}
		for _, col := range spec.search {
			clauses = append(clauses, "LOWER("+col+") LIKE ?")
			args = append(args, "%"+strings.ToLower(search)+"%")
		}
		if len(clauses) > 0 {
			where = append(where, "("+strings.Join(clauses, " OR ")+")")
		}
	}

	sortCol := params.Get("sort")
	if !contains(spec.columns, sortCol) {
		sortCol = spec.columns[0]
	}
	order := "ASC"
	if strings.EqualFold(params.Get("order"), "desc") {
		order = "DESC"
	}
	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil || limit < 1 || limit > 500 {
		limit = 100
	}
	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	statement := "SELECT " + strings.Join(spec.columns, ", ") + " FROM " + table
	if len(where) > 0 {
		statement += " WHERE " + strings.Join(where, " AND ")
	}
	statement += " ORDER BY " + sortCol + " " + order + " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.Store.QueryContext(r.Context(), statement, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		writeError(w, err)
		return
	}
	result := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			writeError(w, err)
			return
		}
		row := map[string]any{}
		for i, col := range columns {
			switch v := values[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		result = append(result, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"table": table, "columns": columns, "rows": result, "count": len(result),
	})
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

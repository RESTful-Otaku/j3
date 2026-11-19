package api

import (
	"fmt"
	"net/http"
	"sort"
)

type kataScoreView struct {
	Judge int     `json:"judge"`
	Score float64 `json:"score"`
}

type kataPerformanceView struct {
	ID        string          `json:"id"`
	FighterID string          `json:"fighterId"`
	Name      string          `json:"name"`
	Club      string          `json:"club"`
	Belt      string          `json:"belt"`
	KataName  string          `json:"kataName"`
	PanelSize int             `json:"panelSize"`
	Status    string          `json:"status"`
	Total     float64         `json:"total"`
	Scores    []kataScoreView `json:"scores"`
}

func (s *Server) createKataPerformance(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FighterID string `json:"fighterId"`
		KataName  string `json:"kataName"`
		PanelSize int    `json:"panelSize"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.FighterID == "" {
		writeBadRequest(w, "fighterId is required")
		return
	}
	if body.PanelSize < 3 || body.PanelSize > 9 {
		body.PanelSize = 5
	}
	id := newID("kata")
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO kata_performances (id, event_id, fighter_id, kata_name, panel_size)
		VALUES (?, 'ev-demo', ?, ?, ?)`, id, body.FighterID, body.KataName, body.PanelSize)
	if err != nil {
		writeError(w, err)
		return
	}
	s.audit(r, "kata.create", "kata_performance", id, fmt.Sprintf(`{"fighter":%q}`, body.FighterID))
	items, err := s.listKata(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"performances": items})
}

func (s *Server) submitKataScore(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Judge int     `json:"judge"`
		Score float64 `json:"score"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if body.Judge < 1 || body.Score < 0 || body.Score > 10 {
		writeBadRequest(w, "judge must be >=1 and score between 0 and 10")
		return
	}
	var panel int
	if err := s.Store.QueryRowContext(r.Context(),
		`SELECT panel_size FROM kata_performances WHERE id = ?`, id).Scan(&panel); err != nil {
		writeError(w, err)
		return
	}
	if body.Judge > panel {
		writeBadRequest(w, "judge index exceeds panel size")
		return
	}
	_, err := s.Store.ExecContext(r.Context(), `
		INSERT INTO kata_scores (performance_id, judge, score) VALUES (?, ?, ?)
		ON CONFLICT(performance_id, judge) DO UPDATE SET score = excluded.score`,
		id, body.Judge, body.Score)
	if err != nil {
		writeError(w, err)
		return
	}

	rows, err := s.Store.QueryContext(r.Context(),
		`SELECT score FROM kata_scores WHERE performance_id = ? ORDER BY judge`, id)
	if err != nil {
		writeError(w, err)
		return
	}
	scores := []float64{}
	for rows.Next() {
		var score float64
		if rows.Scan(&score) == nil {
			scores = append(scores, score)
		}
	}
	rows.Close()
	total := 0.0
	status := "pending"
	if len(scores) >= panel {
		status = "complete"
		ordered := append([]float64(nil), scores...)
		sort.Float64s(ordered)
		if len(ordered) >= 5 {
			ordered = ordered[1 : len(ordered)-1]
		}
		sum := 0.0
		for _, score := range ordered {
			sum += score
		}
		total = sum / float64(len(ordered))
	}
	_, _ = s.Store.ExecContext(r.Context(),
		`UPDATE kata_performances SET total = ?, status = ? WHERE id = ?`, total, status, id)
	s.audit(r, "kata.score", "kata_performance", id,
		fmt.Sprintf(`{"judge":%d,"score":%v,"total":%v}`, body.Judge, body.Score, total))
	items, err := s.listKata(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"performances": items})
}

func (s *Server) listKataHandler(w http.ResponseWriter, r *http.Request) {
	items, err := s.listKata(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"performances": items})
}

func (s *Server) listKata(r *http.Request) ([]kataPerformanceView, error) {
	rows, err := s.Store.QueryContext(r.Context(), `
		SELECT kp.id, kp.fighter_id, COALESCE(f.display_name, kp.fighter_id),
		       COALESCE(c.name,''), COALESCE(b.label,''), kp.kata_name,
		       kp.panel_size, kp.status, kp.total
		FROM kata_performances kp
		LEFT JOIN fighters f ON f.id = kp.fighter_id
		LEFT JOIN clubs c ON c.id = f.club_id
		LEFT JOIN belt_ranks b ON b.id = f.belt_rank_id
		WHERE kp.event_id = 'ev-demo' ORDER BY kp.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []kataPerformanceView{}
	for rows.Next() {
		var item kataPerformanceView
		if err := rows.Scan(&item.ID, &item.FighterID, &item.Name, &item.Club, &item.Belt,
			&item.KataName, &item.PanelSize, &item.Status, &item.Total); err != nil {
			return nil, err
		}
		item.Scores = []kataScoreView{}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range items {
		scoreRows, err := s.Store.QueryContext(r.Context(),
			`SELECT judge, score FROM kata_scores WHERE performance_id = ? ORDER BY judge`, items[i].ID)
		if err != nil {
			return nil, err
		}
		for scoreRows.Next() {
			var score kataScoreView
			if err := scoreRows.Scan(&score.Judge, &score.Score); err != nil {
				scoreRows.Close()
				return nil, err
			}
			items[i].Scores = append(items[i].Scores, score)
		}
		scoreRows.Close()
	}
	return items, nil
}

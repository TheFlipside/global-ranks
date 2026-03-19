package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"global-ranks/internal/model"
	"global-ranks/internal/validate"
)

type leaderboardResponse struct {
	Game    string                   `json:"game"`
	Entries []model.LeaderboardEntry `json:"entries"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
}

// GetLeaderboard handles GET /api/v1/games/{slug}/leaderboard.
func (d *Deps) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if err := validate.GameSlug(slug); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	ctx := r.Context()
	game, err := model.GetGameBySlug(ctx, d.DB, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	entries, total, err := model.GetLeaderboard(ctx, d.DB, game.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}

	if entries == nil {
		entries = []model.LeaderboardEntry{}
	}

	writeJSON(w, http.StatusOK, leaderboardResponse{
		Game:    slug,
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

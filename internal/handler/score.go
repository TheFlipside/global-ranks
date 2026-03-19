package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"global-ranks/internal/model"
	"global-ranks/internal/validate"
)

type submitScoreRequest struct {
	UUID     string `json:"uuid"`
	GameSlug string `json:"game_slug"`
	Score    int64  `json:"score"`
}

type submitScoreResponse struct {
	ID   int64 `json:"id"`
	Rank int   `json:"rank"`
}

// SubmitScore handles POST /api/v1/scores.
func (d *Deps) SubmitScore(w http.ResponseWriter, r *http.Request) {
	var req submitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uid, err := uuid.Parse(req.UUID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	// Per-UUID score submission rate limit.
	if d.ScoreRateLimit != nil && !d.ScoreRateLimit.Allow(uid.String()) {
		w.Header().Set("Retry-After", "5")
		writeError(w, http.StatusTooManyRequests, "score submission rate limit exceeded")
		return
	}

	if err := validate.GameSlug(req.GameSlug); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.Score(req.Score, d.Cfg.MaxScore); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	// Auto-create user on first score submission.
	exists, err := model.UserExists(ctx, d.DB, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !exists {
		if _, err := model.CreateUser(ctx, d.DB, uid); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	}

	// Get or create the game.
	game, err := model.GetOrCreateGame(ctx, d.DB, req.GameSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve game")
		return
	}

	id, rank, err := model.SubmitScore(ctx, d.DB, uid, game.ID, req.Score)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to submit score")
		return
	}

	writeJSON(w, http.StatusCreated, submitScoreResponse{ID: id, Rank: rank})
}

// GetUserUUID extracts and validates a UUID from the request path.
func GetUserUUID(r *http.Request) (uuid.UUID, error) {
	raw := r.PathValue("uuid")
	if raw == "" {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.Parse(raw)
}


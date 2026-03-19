package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"global-ranks/internal/model"
	"global-ranks/internal/validate"
)

type submitScoreRequest struct {
	SessionID string `json:"session_id"`
	Score     int64  `json:"score"`
}

type submitScoreResponse struct {
	ID   int64 `json:"id"`
	Rank int   `json:"rank"`
}

// SubmitScore handles POST /api/v1/scores.
// Requires authentication and a valid, unused session.
func (d *Deps) SubmitScore(w http.ResponseWriter, r *http.Request) {
	uid, ok := d.authenticateRequest(w, r)
	if !ok {
		return
	}

	var req submitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session_id")
		return
	}

	if err := validate.Score(req.Score, d.Cfg.MaxScore); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Per-UUID score submission rate limit.
	if d.ScoreRateLimit != nil && !d.ScoreRateLimit.Allow(uid.String()) {
		w.Header().Set("Retry-After", "5")
		writeError(w, http.StatusTooManyRequests, "score submission rate limit exceeded")
		return
	}

	ctx := r.Context()

	session, err := model.GetSession(ctx, d.DB, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch session")
		return
	}

	if session.UserUUID != uid {
		writeError(w, http.StatusForbidden, "session does not belong to this user")
		return
	}

	if session.Used {
		writeError(w, http.StatusConflict, "session already used")
		return
	}

	// Atomically mark session as used (guards against race conditions).
	if err := model.MarkSessionUsed(ctx, d.DB, sessionID); err != nil {
		writeError(w, http.StatusConflict, "session already used")
		return
	}

	id, rank, err := model.SubmitScore(ctx, d.DB, uid, session.GameID, req.Score)
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

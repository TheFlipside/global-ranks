package handler

import (
	"encoding/json"
	"net/http"

	"global-ranks/internal/model"
	"global-ranks/internal/validate"
)

type createSessionRequest struct {
	GameSlug string `json:"game_slug"`
}

type createSessionResponse struct {
	SessionID string `json:"session_id"`
}

// CreateSession handles POST /api/v1/sessions.
// Requires authentication. Creates a new game session for the user.
func (d *Deps) CreateSession(w http.ResponseWriter, r *http.Request) {
	uid, ok := d.authenticateRequest(w, r)
	if !ok {
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.GameSlug(req.GameSlug); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	game, err := model.GetOrCreateGame(ctx, d.DB, req.GameSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve game")
		return
	}

	session, err := model.CreateSession(ctx, d.DB, uid, game.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{
		SessionID: session.ID.String(),
	})
}

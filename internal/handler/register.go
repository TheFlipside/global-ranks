package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"global-ranks/internal/model"
)

type registerRequest struct {
	UUID string `json:"uuid"`
}

type registerResponse struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	SecretToken string `json:"secret_token"`
	AvatarURL   string `json:"avatar_url"`
}

// Register handles POST /api/v1/register.
// Creates a new user and returns the secret token (returned only once).
func (d *Deps) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uid, err := uuid.Parse(req.UUID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	ctx := r.Context()

	exists, err := model.UserExists(ctx, d.DB, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "user already registered")
		return
	}

	u, err := model.CreateUser(ctx, d.DB, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{
		UUID:        u.UUID.String(),
		Username:    u.Username,
		SecretToken: u.SecretToken,
		AvatarURL:   fmt.Sprintf("/api/v1/avatars/%s.png", u.UUID),
	})
}

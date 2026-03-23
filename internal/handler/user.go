package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"global-ranks/internal/model"
	"global-ranks/internal/validate"
)

type userResponse struct {
	UUID      string `json:"uuid"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type updateUsernameRequest struct {
	Username string `json:"username"`
}

// GetUser handles GET /api/v1/users/{uuid}.
func (d *Deps) GetUser(w http.ResponseWriter, r *http.Request) {
	uid, err := GetUserUUID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	u, err := model.GetUser(r.Context(), d.DB, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		UUID:      u.UUID.String(),
		Username:  u.Username,
		AvatarURL: fmt.Sprintf("/api/v1/avatars/%s.png", u.UUID),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateUser handles PATCH /api/v1/users/{uuid}.
// Requires authentication. The authenticated user must match the path UUID.
func (d *Deps) UpdateUser(w http.ResponseWriter, r *http.Request) {
	authUID, ok := d.authenticateRequest(w, r)
	if !ok {
		return
	}

	uid, err := GetUserUUID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	if authUID != uid {
		writeError(w, http.StatusForbidden, "cannot update another user")
		return
	}

	var req updateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Username(req.Username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := model.UpdateUsername(r.Context(), d.DB, uid, req.Username)
	if err != nil {
		if errors.Is(err, model.ErrUsernameTaken) {
			writeError(w, http.StatusConflict, "username already taken")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		UUID:      u.UUID.String(),
		Username:  u.Username,
		AvatarURL: fmt.Sprintf("/api/v1/avatars/%s.png", u.UUID),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

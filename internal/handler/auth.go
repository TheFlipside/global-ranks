package handler

import (
	"net/http"

	"github.com/google/uuid"

	"global-ranks/internal/model"
)

// authenticateRequest extracts X-User-UUID and X-Auth-Token headers,
// validates them against the database, and returns the authenticated user UUID.
// On failure it writes an error response and returns uuid.Nil.
func (d *Deps) authenticateRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	rawUUID := r.Header.Get("X-User-UUID")
	token := r.Header.Get("X-Auth-Token")

	if rawUUID == "" || token == "" {
		writeError(w, http.StatusUnauthorized, "missing X-User-UUID or X-Auth-Token header")
		return uuid.Nil, false
	}

	uid, err := uuid.Parse(rawUUID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid X-User-UUID")
		return uuid.Nil, false
	}

	valid, err := model.ValidateToken(r.Context(), d.DB, uid, token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return uuid.Nil, false
	}
	if !valid {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return uuid.Nil, false
	}

	return uid, true
}

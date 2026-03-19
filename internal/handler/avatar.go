package handler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"global-ranks/internal/identicon"
)

type avatarCache struct {
	mu      sync.RWMutex
	entries map[string][]byte
	maxSize int
}

func newAvatarCache(maxSize int) *avatarCache {
	return &avatarCache{
		entries: make(map[string][]byte),
		maxSize: maxSize,
	}
}

func (c *avatarCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.entries[key]
	return data, ok
}

func (c *avatarCache) set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		// Evict a random entry when at capacity.
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
	c.entries[key] = data
}

// AvatarHandler returns an http.HandlerFunc for GET /api/v1/avatars/{id}.
// The {id} path value includes the ".png" suffix.
func (d *Deps) AvatarHandler() http.HandlerFunc {
	cache := newAvatarCache(d.Cfg.AvatarCacheSize)

	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.PathValue("id")
		uid := strings.TrimSuffix(raw, ".png")

		if _, err := uuid.Parse(uid); err != nil {
			writeError(w, http.StatusBadRequest, "invalid uuid")
			return
		}

		if data, ok := cache.get(uid); ok {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
			w.Write(data)
			return
		}

		data, err := identicon.Generate(uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate avatar")
			return
		}

		cache.set(uid, data)

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
		w.Write(data)
	}
}

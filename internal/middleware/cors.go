package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that enables cross-origin requests for an allowlist
// of origins.
//
// allowed is a comma-separated list of exact origins (e.g.
// "https://game1.example,https://game2.example"). Surrounding whitespace and
// empty entries are ignored. When the list is empty the middleware is a no-op:
// no CORS headers are emitted and same-origin behaviour is unchanged.
//
// For each request the Origin header is matched against the allowlist. On a
// match the request's own origin is echoed back in Access-Control-Allow-Origin
// (never a comma-separated list, which browsers reject). Vary: Origin is always
// set while CORS is active so shared caches key responses per origin. OPTIONS
// preflight requests are answered with 204 No Content.
func CORS(allowed string) func(http.Handler) http.Handler {
	set := parseOrigins(allowed)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(set) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")

			if origin := r.Header.Get("Origin"); origin != "" {
				if _, ok := set[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-UUID, X-Auth-Token")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// parseOrigins splits a comma-separated origin list into a lookup set,
// trimming whitespace and discarding empty entries.
func parseOrigins(raw string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = struct{}{}
		}
	}
	return set
}

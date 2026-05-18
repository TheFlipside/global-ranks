package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// okHandler is the next handler in the chain; it records whether it ran.
func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORS(t *testing.T) {
	const allowed = " https://game1.example , https://game2.example "

	tests := []struct {
		name        string
		config      string
		method      string
		origin      string
		wantACAO    string // expected Access-Control-Allow-Origin
		wantVary    bool
		wantStatus  int
		wantNextRun bool
	}{
		{
			name:        "empty config is a no-op",
			config:      "",
			method:      http.MethodGet,
			origin:      "https://game1.example",
			wantACAO:    "",
			wantVary:    false,
			wantStatus:  http.StatusOK,
			wantNextRun: true,
		},
		{
			name:        "allowed origin is echoed back",
			config:      allowed,
			method:      http.MethodGet,
			origin:      "https://game2.example",
			wantACAO:    "https://game2.example",
			wantVary:    true,
			wantStatus:  http.StatusOK,
			wantNextRun: true,
		},
		{
			name:        "disallowed origin gets no ACAO header",
			config:      allowed,
			method:      http.MethodGet,
			origin:      "https://evil.example",
			wantACAO:    "",
			wantVary:    true,
			wantStatus:  http.StatusOK,
			wantNextRun: true,
		},
		{
			name:        "preflight from allowed origin short-circuits with 204",
			config:      allowed,
			method:      http.MethodOptions,
			origin:      "https://game1.example",
			wantACAO:    "https://game1.example",
			wantVary:    true,
			wantStatus:  http.StatusNoContent,
			wantNextRun: false,
		},
		{
			name:        "preflight from disallowed origin gets 204 but no ACAO",
			config:      allowed,
			method:      http.MethodOptions,
			origin:      "https://evil.example",
			wantACAO:    "",
			wantVary:    true,
			wantStatus:  http.StatusNoContent,
			wantNextRun: false,
		},
		{
			name:        "no origin header, CORS active",
			config:      allowed,
			method:      http.MethodGet,
			origin:      "",
			wantACAO:    "",
			wantVary:    true,
			wantStatus:  http.StatusOK,
			wantNextRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nextRan bool
			h := CORS(tt.config)(okHandler(&nextRan))

			req := httptest.NewRequest(tt.method, "/api/v1/health", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantACAO {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantACAO)
			}
			gotVary := false
			for _, v := range rec.Header().Values("Vary") {
				if strings.Contains(v, "Origin") {
					gotVary = true
					break
				}
			}
			if gotVary != tt.wantVary {
				t.Errorf("Vary: Origin present = %v, want %v", gotVary, tt.wantVary)
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextRan != tt.wantNextRun {
				t.Errorf("next handler ran = %v, want %v", nextRan, tt.wantNextRun)
			}
		})
	}
}

func TestParseOrigins(t *testing.T) {
	set := parseOrigins("  https://a.example , ,https://b.example,")
	if len(set) != 2 {
		t.Fatalf("len = %d, want 2 (%v)", len(set), set)
	}
	for _, want := range []string{"https://a.example", "https://b.example"} {
		if _, ok := set[want]; !ok {
			t.Errorf("missing %q in %v", want, set)
		}
	}
	if got := parseOrigins(""); len(got) != 0 {
		t.Errorf("empty input -> len %d, want 0", len(got))
	}
}

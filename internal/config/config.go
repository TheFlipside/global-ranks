package config

import (
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

// Config holds all server configuration loaded from environment variables.
type Config struct {
	DBDSN            string
	Port             string
	RateScorePerSec  float64
	RateScoreBurst   int
	RateGeneralPerSec float64
	RateGeneralBurst int
	AvatarCacheSize  int
	MaxScore         int64
	TrustedProxies   []*net.IPNet
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		DBDSN:             envStr("GR_DB_DSN", "postgres://localhost:5432/globalranks?sslmode=disable"),
		Port:              envStr("GR_PORT", "8080"),
		RateScorePerSec:   envFloat("GR_RATE_SCORE_PER_SEC", 0.2),
		RateScoreBurst:    envInt("GR_RATE_SCORE_BURST", 3),
		RateGeneralPerSec: envFloat("GR_RATE_GENERAL_PER_SEC", 10),
		RateGeneralBurst:  envInt("GR_RATE_GENERAL_BURST", 30),
		AvatarCacheSize:   envInt("GR_AVATAR_CACHE_SIZE", 1000),
		MaxScore:          int64(envInt("GR_MAX_SCORE", 999999999)),
		TrustedProxies:    parseCIDRs(envStr("GR_TRUSTED_PROXIES", "")),
	}
}

// parseCIDRs splits a comma-separated list of CIDRs into []*net.IPNet.
// Single IPs without a mask (e.g. "127.0.0.1") are treated as /32 or /128.
func parseCIDRs(raw string) []*net.IPNet {
	if raw == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.Contains(s, "/") {
			if strings.Contains(s, ":") {
				s += "/128"
			} else {
				s += "/32"
			}
		}
		_, cidr, err := net.ParseCIDR(s)
		if err != nil {
			slog.Warn("ignoring invalid CIDR in GR_TRUSTED_PROXIES", "value", s, "error", err)
			continue
		}
		nets = append(nets, cidr)
	}
	return nets
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
		slog.Warn("invalid integer for env var, using fallback", "key", key, "value", v, "fallback", fallback)
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
		slog.Warn("invalid float for env var, using fallback", "key", key, "value", v, "fallback", fallback)
	}
	return fallback
}

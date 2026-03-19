package config

import (
	"os"
	"strconv"
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
	}
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
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}
	return fallback
}

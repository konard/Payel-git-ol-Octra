// Package config loads runtime configuration for the Octra backend from
// environment variables. The whole platform is a single monolith, so all
// settings live in one place.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds every tunable setting for the backend monolith.
type Config struct {
	// HTTPAddr is the listen address for the public HTTP API, e.g. ":8080".
	HTTPAddr string

	// GRPCAddr is the listen address for the gRPC server, e.g. ":9090".
	GRPCAddr string

	// DBDsn is the PostgreSQL DSN consumed by GORM.
	DBDsn string

	// RedisURL is the redis connection URL, e.g. "redis://localhost:6379/0".
	RedisURL string

	// EnvironmentsDir is the host directory where per-user Nix environments
	// (one profile per user) are created.
	EnvironmentsDir string

	// CLITTL is how long a started CLI subprocess is kept alive before it is
	// considered stale and force-killed on the next request.
	CLITTL time.Duration

	// JWT secrets for access and refresh tokens.
	JWTSecret        string
	JWTRefreshSecret string

	// FrontendURL is used by OAuth callbacks to redirect the browser.
	FrontendURL string

	// OAuth client credentials.
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string

	// LeFine integration.
	LeFineBaseURL          string
	LeFineRedirectURL      string
	LeFineIntegrationSecret string

	// Skills API base URL for fetching skill definitions.
	BaseURLGetSkills string

	// Typesense configuration.
	TypesenseHost   string
	TypesenseAPIKey string
}

// Load reads configuration from the environment, applying sensible defaults so
// the service can boot locally without any setup.
func Load() Config {
	return Config{
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr:        getEnv("GRPC_ADDR", ":9090"),
		DBDsn:           getEnv("DB_DNS", "host=localhost user=octra password=octra dbname=octra port=5432 sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EnvironmentsDir: getEnv("ENVIRONMENTS_DIR", "/var/lib/octra/environments"),
		CLITTL:          getEnvDuration("CLI_TTL", 30*time.Minute),

		JWTSecret:              getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		JWTRefreshSecret:       getEnv("JWT_REFRESH_SECRET", "dev-jwt-refresh-secret-change-in-production"),
		FrontendURL:            getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoogleClientID:         getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleClientSecret:     getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		GoogleRedirectURL:      getEnv("GOOGLE_REDIRECT_URL", ""),
		GitHubClientID:         getEnv("GITHUB_OAUTH_CLIENT_ID", ""),
		GitHubClientSecret:     getEnv("GITHUB_OAUTH_CLIENT_SECRET", ""),
		GitHubRedirectURL:      getEnv("GITHUB_REDIRECT_URL", ""),
		LeFineBaseURL:          getEnv("LEFINE_BASE_URL", "https://lefine.pro"),
		LeFineRedirectURL:      getEnv("LEFINE_REDIRECT_URL", ""),
		LeFineIntegrationSecret: getEnv("LEFINE_INTEGRATION_SECRET", "dev-only-lefine-octra-shared-secret-change-me"),

		BaseURLGetSkills: getEnv("BASE_URL_GET_SKILLS", "https://skills.sh"),

		TypesenseHost:   getEnv("TYPESENSE_HOST", "typesense:8108"),
		TypesenseAPIKey: getEnv("TYPESENSE_API_KEY", "octra-typesense-key"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		// Allow either a Go duration ("30m") or a plain number of seconds.
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return fallback
}

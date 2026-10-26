package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultJWTTTL      = 8 * time.Hour
	minJWTSecretLength = 32
)

var (
	ErrMissingHTTPPort    = errors.New("missing required env var: HTTP_PORT")
	ErrMissingDatabaseURL = errors.New("missing required env var: DATABASE_URL")
	ErrMissingJWTSecret   = errors.New("missing required env var: JWT_SECRET")
	ErrShortJWTSecret     = fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	ErrInvalidJWTTTL      = errors.New("JWT_TTL must be a valid duration (e.g. 8h)")
)

type Config struct {
	HTTPPort      string
	DatabaseURL   string
	JWTSecret     string
	JWTTTL        time.Duration
	AdminEmail    string
	AdminPassword string
	// OPSBot (opcional): sin estas variables el chat responde 503.
	AIServiceURL    string
	AIInternalToken string
}

func Load() (Config, error) {
	httpPort := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if httpPort == "" {
		return Config{}, ErrMissingHTTPPort
	}
	if databaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}
	if jwtSecret == "" {
		return Config{}, ErrMissingJWTSecret
	}
	if len(jwtSecret) < minJWTSecretLength {
		return Config{}, ErrShortJWTSecret
	}

	jwtTTL := defaultJWTTTL
	if raw := strings.TrimSpace(os.Getenv("JWT_TTL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("%w: %q", ErrInvalidJWTTTL, raw)
		}
		jwtTTL = parsed
	}

	return Config{
		HTTPPort:      httpPort,
		DatabaseURL:   databaseURL,
		JWTSecret:     jwtSecret,
		JWTTTL:        jwtTTL,
		AdminEmail:    strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),

		AIServiceURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("AI_SERVICE_URL")), "/"),
		AIInternalToken: strings.TrimSpace(os.Getenv("AI_INTERNAL_TOKEN")),
	}, nil
}

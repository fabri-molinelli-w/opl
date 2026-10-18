package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrMissingHTTPPort    = errors.New("missing required env var: HTTP_PORT")
	ErrMissingDatabaseURL = errors.New("missing required env var: DATABASE_URL")
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() (Config, error) {
	httpPort := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	if httpPort == "" {
		return Config{}, ErrMissingHTTPPort
	}
	if databaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}

	return Config{
		HTTPPort:    httpPort,
		DatabaseURL: databaseURL,
	}, nil
}

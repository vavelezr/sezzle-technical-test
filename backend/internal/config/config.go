// Package config loads the service configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Environment variable names read by Load.
const (
	EnvPort           = "PORT"
	EnvAllowedOrigins = "CORS_ALLOWED_ORIGINS"
	EnvLogLevel       = "LOG_LEVEL"
)

// Config holds every tunable setting of the service.
type Config struct {
	Port            int
	AllowedOrigins  []string
	LogLevel        slog.Level
	MaxBodyBytes    int64
	RequestTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Default returns the configuration used when no environment overrides exist.
func Default() Config {
	return Config{
		Port:            8080,
		AllowedOrigins:  []string{"http://localhost:5173", "http://localhost:3000"},
		LogLevel:        slog.LevelInfo,
		MaxBodyBytes:    4 << 10, // 4 KiB is plenty for a calculation payload.
		RequestTimeout:  5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// Addr returns the TCP listen address derived from Port.
func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.Port)
}

// Load starts from Default and applies overrides read through lookup
// (typically os.LookupEnv). Unset or blank variables keep their defaults.
func Load(lookup func(string) (string, bool)) (Config, error) {
	cfg := Default()

	if v, ok := nonBlank(lookup, EnvPort); ok {
		port, err := strconv.Atoi(v)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("%s must be an integer between 1 and 65535, got %q", EnvPort, v)
		}
		cfg.Port = port
	}

	if v, ok := nonBlank(lookup, EnvAllowedOrigins); ok {
		cfg.AllowedOrigins = splitList(v)
	}

	if v, ok := nonBlank(lookup, EnvLogLevel); ok {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return Config{}, fmt.Errorf("%s: %w", EnvLogLevel, err)
		}
	}

	return cfg, nil
}

func nonBlank(lookup func(string) (string, bool), key string) (string, bool) {
	v, ok := lookup(key)
	v = strings.TrimSpace(v)
	return v, ok && v != ""
}

func splitList(v string) []string {
	var items []string
	for item := range strings.SplitSeq(v, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

package config_test

import (
	"log/slog"
	"slices"
	"testing"

	"sezzle-calculator/backend/internal/config"
)

func lookupFrom(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestLoad(t *testing.T) {
	defaults := config.Default()

	tests := []struct {
		name        string
		env         map[string]string
		wantPort    int
		wantOrigins []string
		wantLevel   slog.Level
		wantErr     bool
	}{
		{
			name:        "defaults when unset",
			env:         map[string]string{},
			wantPort:    8080,
			wantOrigins: defaults.AllowedOrigins,
			wantLevel:   slog.LevelInfo,
		},
		{
			name:        "blank values keep defaults",
			env:         map[string]string{"PORT": "  ", "CORS_ALLOWED_ORIGINS": "", "LOG_LEVEL": " "},
			wantPort:    8080,
			wantOrigins: defaults.AllowedOrigins,
			wantLevel:   slog.LevelInfo,
		},
		{
			name: "overrides",
			env: map[string]string{
				"PORT":                 "9090",
				"CORS_ALLOWED_ORIGINS": " https://a.example , ,https://b.example",
				"LOG_LEVEL":            "debug",
			},
			wantPort:    9090,
			wantOrigins: []string{"https://a.example", "https://b.example"},
			wantLevel:   slog.LevelDebug,
		},
		{name: "non numeric port", env: map[string]string{"PORT": "http"}, wantErr: true},
		{name: "port out of range", env: map[string]string{"PORT": "70000"}, wantErr: true},
		{name: "zero port", env: map[string]string{"PORT": "0"}, wantErr: true},
		{name: "invalid log level", env: map[string]string{"LOG_LEVEL": "verbose"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(lookupFrom(tt.env))

			if tt.wantErr {
				if err == nil {
					t.Fatal("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", cfg.Port, tt.wantPort)
			}
			if !slices.Equal(cfg.AllowedOrigins, tt.wantOrigins) {
				t.Errorf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, tt.wantOrigins)
			}
			if cfg.LogLevel != tt.wantLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.wantLevel)
			}
		})
	}
}

func TestAddr(t *testing.T) {
	cfg := config.Default()
	cfg.Port = 3001

	if got := cfg.Addr(); got != ":3001" {
		t.Errorf("Addr() = %q, want %q", got, ":3001")
	}
}

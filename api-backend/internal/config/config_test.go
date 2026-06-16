package config

import (
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		want        Config
		wantErr     string
	}{
		{
			name: "cache enabled",
			environment: map[string]string{
				"DB_CONN": "redis://default:password@redis:6379",
			},
			want: Config{
				DBConn: "redis://default:password@redis:6379",
			},
		},
		{
			name: "cache disabled with legacy truthy value",
			environment: map[string]string{
				"CACHE_DISABLED": "on",
			},
			want: Config{
				CacheDisabled: true,
			},
		},
		{
			name: "cache disabled with legacy falsy value",
			environment: map[string]string{
				"DB_CONN":        "redis://default:password@redis:6379",
				"CACHE_DISABLED": "off",
			},
			want: Config{
				DBConn: "redis://default:password@redis:6379",
			},
		},
		{
			name:    "cache enabled requires db conn",
			wantErr: "DB_CONN is required when caching is enabled",
		},
		{
			name: "invalid bool",
			environment: map[string]string{
				"DB_CONN":        "redis://default:password@redis:6379",
				"CACHE_DISABLED": "sometimes",
			},
			wantErr: "parse error on field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := load(env.Options{Environment: tt.environment})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("load() error = nil, want %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("load() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("load() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("load() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "true", want: true},
		{value: "1", want: true},
		{value: "yes", want: true},
		{value: "on", want: true},
		{value: "false", want: false},
		{value: "0", want: false},
		{value: "no", want: false},
		{value: "off", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseBool(tt.value)
			if err != nil {
				t.Fatalf("parseBool() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

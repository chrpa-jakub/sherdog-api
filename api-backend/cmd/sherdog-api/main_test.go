package main

import "testing"

func TestCacheDisabled(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "true", want: true},
		{value: "1", want: true},
		{value: "yes", want: true},
		{value: "on", want: true},
		{value: "false", want: false},
		{value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv("CACHE_DISABLED", tt.value)
			if got := cacheDisabled(); got != tt.want {
				t.Fatalf("cacheDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

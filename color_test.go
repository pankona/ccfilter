package main

import "testing"

func TestColorize(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		color   string
		enabled bool
		want    string
	}{
		{
			name:    "green with color enabled",
			text:    "success",
			color:   "green",
			enabled: true,
			want:    "\x1b[32msuccess\x1b[0m",
		},
		{
			name:    "red with color enabled",
			text:    "error",
			color:   "red",
			enabled: true,
			want:    "\x1b[31merror\x1b[0m",
		},
		{
			name:    "blue with color enabled",
			text:    "info",
			color:   "blue",
			enabled: true,
			want:    "\x1b[34minfo\x1b[0m",
		},
		{
			name:    "cyan with color enabled",
			text:    "tool",
			color:   "cyan",
			enabled: true,
			want:    "\x1b[36mtool\x1b[0m",
		},
		{
			name:    "yellow with color enabled",
			text:    "warning",
			color:   "yellow",
			enabled: true,
			want:    "\x1b[33mwarning\x1b[0m",
		},
		{
			name:    "gray with color enabled",
			text:    "metadata",
			color:   "gray",
			enabled: true,
			want:    "\x1b[90mmetadata\x1b[0m",
		},
		{
			name:    "color disabled",
			text:    "text",
			color:   "green",
			enabled: false,
			want:    "text",
		},
		{
			name:    "unknown color",
			text:    "text",
			color:   "unknown",
			enabled: true,
			want:    "text",
		},
		{
			name:    "empty text",
			text:    "",
			color:   "green",
			enabled: true,
			want:    "\x1b[32m\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorize(tt.text, tt.color, tt.enabled)
			if got != tt.want {
				t.Errorf("colorize() = %q, want %q", got, tt.want)
			}
		})
	}
}

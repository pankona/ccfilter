package main

import (
	"testing"
)

func TestNewFilterConfig(t *testing.T) {
	config := NewFilterConfig()

	if config.ShowSystem {
		t.Error("ShowSystem should be false by default")
	}
	if !config.ShowAssistant {
		t.Error("ShowAssistant should be true by default")
	}
	if !config.ShowTools {
		t.Error("ShowTools should be true by default")
	}
	if !config.ShowResult {
		t.Error("ShowResult should be true by default")
	}
	if config.InfoLevel != "standard" {
		t.Errorf("InfoLevel = %v, want standard", config.InfoLevel)
	}
	if config.Format != "text" {
		t.Errorf("Format = %v, want text", config.Format)
	}
	if !config.UseColor {
		t.Error("UseColor should be true by default")
	}
}

func TestParseMessageType(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    string
		wantErr bool
	}{
		{
			name:    "system message",
			line:    `{"type":"system","subtype":"init"}`,
			want:    "system",
			wantErr: false,
		},
		{
			name:    "assistant message",
			line:    `{"type":"assistant","message":{"content":[]}}`,
			want:    "assistant",
			wantErr: false,
		},
		{
			name:    "user message",
			line:    `{"type":"user","message":{"role":"user"}}`,
			want:    "user",
			wantErr: false,
		},
		{
			name:    "result message",
			line:    `{"type":"result","subtype":"success"}`,
			want:    "result",
			wantErr: false,
		},
		{
			name:    "invalid json",
			line:    `{invalid json}`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessageType(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMessageType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseMessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldDisplay(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		config  FilterConfig
		want    bool
	}{
		{
			name:    "system with ShowSystem=false",
			msgType: "system",
			config:  FilterConfig{ShowSystem: false},
			want:    false,
		},
		{
			name:    "system with ShowSystem=true",
			msgType: "system",
			config:  FilterConfig{ShowSystem: true},
			want:    true,
		},
		{
			name:    "assistant with ShowAssistant=false",
			msgType: "assistant",
			config:  FilterConfig{ShowAssistant: false},
			want:    false,
		},
		{
			name:    "assistant with ShowAssistant=true",
			msgType: "assistant",
			config:  FilterConfig{ShowAssistant: true},
			want:    true,
		},
		{
			name:    "assistant with default config",
			msgType: "assistant",
			config:  *NewFilterConfig(),
			want:    true,
		},
		{
			name:    "user (tool result) with ShowTools=false",
			msgType: "user",
			config:  FilterConfig{ShowTools: false},
			want:    false,
		},
		{
			name:    "user (tool result) with ShowTools=true",
			msgType: "user",
			config:  FilterConfig{ShowTools: true},
			want:    true,
		},
		{
			name:    "result with ShowResult=false",
			msgType: "result",
			config:  FilterConfig{ShowResult: false},
			want:    false,
		},
		{
			name:    "result with ShowResult=true",
			msgType: "result",
			config:  FilterConfig{ShowResult: true},
			want:    true,
		},
		{
			name:    "unknown type",
			msgType: "unknown",
			config:  FilterConfig{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldDisplay(tt.msgType, &tt.config)
			if got != tt.want {
				t.Errorf("shouldDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldDisplayContent(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		config  FilterConfig
		want    bool
	}{
		{
			name:    "text content with ShowAssistant=true",
			content: Content{Type: "text", Text: "Hello"},
			config:  FilterConfig{ShowAssistant: true},
			want:    true,
		},
		{
			name:    "text content with ShowAssistant=false",
			content: Content{Type: "text", Text: "Hello"},
			config:  FilterConfig{ShowAssistant: false},
			want:    false,
		},
		{
			name:    "tool_use content with ShowTools=true",
			content: Content{Type: "tool_use", ID: "1", Name: "Bash"},
			config:  FilterConfig{ShowTools: true},
			want:    true,
		},
		{
			name:    "tool_use content with ShowTools=false",
			content: Content{Type: "tool_use", ID: "1", Name: "Bash"},
			config:  FilterConfig{ShowTools: false},
			want:    false,
		},
		{
			name:    "text content with default config",
			content: Content{Type: "text", Text: "Hello"},
			config:  *NewFilterConfig(),
			want:    true,
		},
		{
			name:    "tool_use content with default config",
			content: Content{Type: "tool_use", ID: "1", Name: "Read"},
			config:  *NewFilterConfig(),
			want:    true,
		},
		{
			name:    "unknown content type",
			content: Content{Type: "unknown"},
			config:  FilterConfig{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldDisplayContent(tt.content, &tt.config)
			if got != tt.want {
				t.Errorf("shouldDisplayContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

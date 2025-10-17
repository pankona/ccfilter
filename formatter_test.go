package main

import (
	"strings"
	"testing"
)

func TestFormatAssistantMessage_Text(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		config FilterConfig
		want   string
	}{
		{
			name:   "simple text",
			input:  `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, World!"}]}}`,
			config: FilterConfig{ShowAssistant: true},
			want:   "Hello, World!\n",
		},
		{
			name:   "multiline text",
			input:  `{"type":"assistant","message":{"content":[{"type":"text","text":"Line 1\nLine 2\nLine 3"}]}}`,
			config: FilterConfig{ShowAssistant: true},
			want:   "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:   "japanese text",
			input:  `{"type":"assistant","message":{"content":[{"type":"text","text":"こんにちは、世界！"}]}}`,
			config: FilterConfig{ShowAssistant: true},
			want:   "こんにちは、世界！\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatAssistantMessage([]byte(tt.input), &tt.config)
			if err != nil {
				t.Errorf("formatAssistantMessage() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("formatAssistantMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatResultMessage(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		config FilterConfig
		want   []string // 含まれるべき文字列のリスト
	}{
		{
			name:   "success with standard config",
			input:  `{"type":"result","subtype":"success","result":"完了しました","duration_ms":5000,"total_cost_usd":0.0123,"num_turns":3}`,
			config: FilterConfig{InfoLevel: "standard"},
			want: []string{
				"━━━",
				"完了しました",
				"Duration: 5.0s",
				"Cost: $0.0123",
				"Turns: 3",
			},
		},
		{
			name:   "minimal config",
			input:  `{"type":"result","subtype":"success","result":"完了しました","duration_ms":5000,"total_cost_usd":0.0123,"num_turns":3}`,
			config: FilterConfig{InfoLevel: "minimal"},
			want: []string{
				"完了しました",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatResultMessage([]byte(tt.input), &tt.config)
			if err != nil {
				t.Errorf("formatResultMessage() error = %v", err)
				return
			}
			for _, substr := range tt.want {
				if !strings.Contains(got, substr) {
					t.Errorf("formatResultMessage() does not contain %q\nGot: %s", substr, got)
				}
			}
		})
	}
}

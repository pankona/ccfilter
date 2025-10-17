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

func TestFormatToolUse(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		config  FilterConfig
		want    string
	}{
		{
			name: "Glob with standard config",
			content: Content{
				Type:  "tool_use",
				Name:  "Glob",
				Input: []byte(`{"pattern":"**/*.go"}`),
			},
			config: FilterConfig{InfoLevel: "standard"},
			want:   "→ Glob: pattern=\"**/*.go\"\n",
		},
		{
			name: "Bash with minimal config",
			content: Content{
				Type:  "tool_use",
				Name:  "Bash",
				Input: []byte(`{"command":"ls -la"}`),
			},
			config: FilterConfig{InfoLevel: "minimal"},
			want:   "→ Bash\n",
		},
		{
			name: "Read with standard config",
			content: Content{
				Type:  "tool_use",
				Name:  "Read",
				Input: []byte(`{"file_path":"/path/to/file.txt"}`),
			},
			config: FilterConfig{InfoLevel: "standard"},
			want:   "→ Read: file_path=\"/path/to/file.txt\"\n",
		},
		{
			name: "Write with standard config",
			content: Content{
				Type:  "tool_use",
				Name:  "Write",
				Input: []byte(`{"file_path":"/path/to/file.txt","content":"data"}`),
			},
			config: FilterConfig{InfoLevel: "standard"},
			want:   "→ Write: file_path=\"/path/to/file.txt\"\n",
		},
		{
			name: "Grep with standard config",
			content: Content{
				Type:  "tool_use",
				Name:  "Grep",
				Input: []byte(`{"pattern":"func.*Test"}`),
			},
			config: FilterConfig{InfoLevel: "standard"},
			want:   "→ Grep: pattern=\"func.*Test\"\n",
		},
		{
			name: "tool with no parameters",
			content: Content{
				Type:  "tool_use",
				Name:  "UnknownTool",
				Input: []byte(`{}`),
			},
			config: FilterConfig{InfoLevel: "standard"},
			want:   "→ UnknownTool\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatToolUse(tt.content, &tt.config)
			if got != tt.want {
				t.Errorf("formatToolUse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLines int
		want     string
	}{
		{
			name:     "no truncation needed",
			input:    "line1\nline2\nline3",
			maxLines: 5,
			want:     "line1\nline2\nline3",
		},
		{
			name:     "truncate with ellipsis",
			input:    "line1\nline2\nline3\nline4\nline5\nline6",
			maxLines: 3,
			want:     "line1\nline2\nline3\n... (3 more lines)",
		},
		{
			name:     "unlimited",
			input:    "line1\nline2\nline3\nline4\nline5",
			maxLines: -1,
			want:     "line1\nline2\nline3\nline4\nline5",
		},
		{
			name:     "exact match",
			input:    "line1\nline2\nline3",
			maxLines: 3,
			want:     "line1\nline2\nline3",
		},
		{
			name:     "single line",
			input:    "only one line",
			maxLines: 5,
			want:     "only one line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateOutput(tt.input, tt.maxLines)
			if got != tt.want {
				t.Errorf("truncateOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatUserMessage_ToolResult(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		config FilterConfig
		want   []string // 含まれるべき文字列
	}{
		{
			name:   "successful tool result",
			input:  `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"Success: file found"}]}}`,
			config: FilterConfig{InfoLevel: "standard"},
			want: []string{
				"← Success: file found",
			},
		},
		{
			name:   "error tool result",
			input:  `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","is_error":true,"content":"Error: permission denied"}]}}`,
			config: FilterConfig{InfoLevel: "standard"},
			want: []string{
				"← Error:",
				"Error: permission denied",
			},
		},
		{
			name:   "minimal mode",
			input:  `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"Line1\nLine2\nLine3\nLine4"}]}}`,
			config: FilterConfig{InfoLevel: "minimal"},
			want: []string{
				"← Line1",
			},
		},
		{
			name:   "long output truncated in standard mode",
			input:  `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"Line1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7"}]}}`,
			config: FilterConfig{InfoLevel: "standard"},
			want: []string{
				"← Line1",
				"Line5",
				"... (2 more lines)",
			},
		},
		{
			name:   "verbose mode shows all",
			input:  `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"Line1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7"}]}}`,
			config: FilterConfig{InfoLevel: "verbose"},
			want: []string{
				"← Line1",
				"Line7",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatUserMessage([]byte(tt.input), &tt.config)
			if err != nil {
				t.Errorf("formatUserMessage() error = %v", err)
				return
			}
			for _, substr := range tt.want {
				if !strings.Contains(got, substr) {
					t.Errorf("formatUserMessage() does not contain %q\nGot: %s", substr, got)
				}
			}
		})
	}
}

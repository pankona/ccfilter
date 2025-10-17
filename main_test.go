package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestProcessInput_Integration(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		config     FilterConfig
		wantOutput []string // 含まれるべき文字列
		wantErr    bool
	}{
		{
			name: "simple text message",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
			config: FilterConfig{
				ShowAssistant: true,
				ShowResult:    true,
				InfoLevel:     "standard",
				UseColor:      false,
			},
			wantOutput: []string{"Hello", "Done", "Duration"},
			wantErr:    false,
		},
		{
			name: "tool use and result",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Searching..."}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Glob","input":{"pattern":"*.go"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"main.go"}]}}`,
			config: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				InfoLevel:     "standard",
				UseColor:      false,
			},
			wantOutput: []string{"Searching", "→ Glob", "← main.go"},
			wantErr:    false,
		},
		{
			name: "filtering - only result",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
			config: FilterConfig{
				ShowAssistant: false,
				ShowResult:    true,
				InfoLevel:     "standard",
				UseColor:      false,
			},
			wantOutput: []string{"Done"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			var output bytes.Buffer

			err := processInput(input, &output, &tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("processInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			result := output.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(result, want) {
					t.Errorf("processInput() output does not contain %q\nGot: %s", want, result)
				}
			}
		})
	}
}

func TestProcessInput_ErrorHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid json",
			input: `{invalid json}`,
		},
		{
			name:  "empty lines",
			input: "\n\n\n",
		},
		{
			name: "mixed valid and invalid",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"OK"}]}}
{invalid}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			var output bytes.Buffer
			config := NewFilterConfig()
			config.UseColor = false

			// エラーで終了しないことを確認
			err := processInput(input, &output, config)
			if err != nil {
				t.Errorf("processInput() should not return error for invalid JSON, got: %v", err)
			}
		})
	}
}

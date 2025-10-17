package main

import (
	"bytes"
	"flag"
	"os"
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

func TestParseArgs(t *testing.T) {
	// テストのためのフラグリセット用ヘルパー
	resetFlags := func() {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	}

	tests := []struct {
		name    string
		args    []string
		want    FilterConfig
		wantErr bool
	}{
		{
			name: "default config",
			args: []string{},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name: "show all",
			args: []string{"--all"},
			want: FilterConfig{
				ShowSystem:    true,
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name: "minimal mode",
			args: []string{"--minimal"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "minimal",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name: "verbose mode",
			args: []string{"--verbose"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "verbose",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name: "no color",
			args: []string{"--no-color"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      false, // --no-color で無効
			},
			wantErr: false,
		},
		{
			name: "force color",
			args: []string{"--color"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // --color で強制有効
			},
			wantErr: false,
		},
		{
			name: "show cost",
			args: []string{"--show-cost"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     true,
				ShowResult:    true,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
				ShowCost:      true,
			},
			wantErr: false,
		},
		{
			name: "assistant only",
			args: []string{"--assistant"},
			want: FilterConfig{
				ShowAssistant: true,
				ShowTools:     false,
				ShowResult:    false,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name: "tools only",
			args: []string{"--tools"},
			want: FilterConfig{
				ShowAssistant: false,
				ShowTools:     true,
				ShowResult:    false,
				InfoLevel:     "standard",
				Format:        "text",
				UseColor:      true, // デフォルトでtrue
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			args:    []string{"--format=invalid"},
			want:    FilterConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			os.Args = append([]string{"cmd"}, tt.args...)

			got, err := parseArgs()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if got.ShowSystem != tt.want.ShowSystem {
					t.Errorf("ShowSystem = %v, want %v", got.ShowSystem, tt.want.ShowSystem)
				}
				if got.ShowAssistant != tt.want.ShowAssistant {
					t.Errorf("ShowAssistant = %v, want %v", got.ShowAssistant, tt.want.ShowAssistant)
				}
				if got.ShowTools != tt.want.ShowTools {
					t.Errorf("ShowTools = %v, want %v", got.ShowTools, tt.want.ShowTools)
				}
				if got.ShowResult != tt.want.ShowResult {
					t.Errorf("ShowResult = %v, want %v", got.ShowResult, tt.want.ShowResult)
				}
				if got.InfoLevel != tt.want.InfoLevel {
					t.Errorf("InfoLevel = %v, want %v", got.InfoLevel, tt.want.InfoLevel)
				}
				if got.Format != tt.want.Format {
					t.Errorf("Format = %v, want %v", got.Format, tt.want.Format)
				}
				if got.UseColor != tt.want.UseColor {
					t.Errorf("UseColor = %v, want %v", got.UseColor, tt.want.UseColor)
				}
				if got.ShowCost != tt.want.ShowCost {
					t.Errorf("ShowCost = %v, want %v", got.ShowCost, tt.want.ShowCost)
				}
			}
		})
	}
}

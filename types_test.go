package main

import (
	"encoding/json"
	"testing"
)

// TestParseMessage_Type はメッセージのtypeフィールドを正しくパースできることを確認する
func TestParseMessage_Type(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "system type",
			input:   `{"type":"system"}`,
			want:    "system",
			wantErr: false,
		},
		{
			name:    "assistant type",
			input:   `{"type":"assistant"}`,
			want:    "assistant",
			wantErr: false,
		},
		{
			name:    "user type",
			input:   `{"type":"user"}`,
			want:    "user",
			wantErr: false,
		},
		{
			name:    "result type",
			input:   `{"type":"result"}`,
			want:    "result",
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg Message
			err := json.Unmarshal([]byte(tt.input), &msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && msg.Type != tt.want {
				t.Errorf("Message.Type = %v, want %v", msg.Type, tt.want)
			}
		})
	}
}

// TestParseSystemMessage はsystemメッセージを正しくパースできることを確認する
func TestParseSystemMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantType string
		wantSubtype string
		wantModel string
		wantErr bool
	}{
		{
			name: "init system message",
			input: `{"type":"system","subtype":"init","cwd":"/test","session_id":"test-123","model":"claude-sonnet-4-5-20250929","claude_code_version":"2.0.21"}`,
			wantType: "system",
			wantSubtype: "init",
			wantModel: "claude-sonnet-4-5-20250929",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg SystemMessage
			err := json.Unmarshal([]byte(tt.input), &msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if msg.Type != tt.wantType {
					t.Errorf("SystemMessage.Type = %v, want %v", msg.Type, tt.wantType)
				}
				if msg.Subtype != tt.wantSubtype {
					t.Errorf("SystemMessage.Subtype = %v, want %v", msg.Subtype, tt.wantSubtype)
				}
				if msg.Model != tt.wantModel {
					t.Errorf("SystemMessage.Model = %v, want %v", msg.Model, tt.wantModel)
				}
			}
		})
	}
}

// TestParseAssistantMessage_Text はassistantメッセージ(textコンテンツ)を正しくパースできることを確認する
func TestParseAssistantMessage_Text(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantText string
		wantErr  bool
	}{
		{
			name:     "simple text content",
			input:    `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, World!"}]}}`,
			wantType: "assistant",
			wantText: "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "japanese text",
			input:    `{"type":"assistant","message":{"content":[{"type":"text","text":"こんにちは"}]}}`,
			wantType: "assistant",
			wantText: "こんにちは",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg AssistantMessage
			err := json.Unmarshal([]byte(tt.input), &msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if msg.Type != tt.wantType {
					t.Errorf("AssistantMessage.Type = %v, want %v", msg.Type, tt.wantType)
				}
				if len(msg.Message.Content) == 0 {
					t.Errorf("AssistantMessage.Message.Content is empty")
					return
				}
				if msg.Message.Content[0].Type != "text" {
					t.Errorf("Content[0].Type = %v, want text", msg.Message.Content[0].Type)
				}
				if msg.Message.Content[0].Text != tt.wantText {
					t.Errorf("Content[0].Text = %v, want %v", msg.Message.Content[0].Text, tt.wantText)
				}
			}
		})
	}
}

// TestParseAssistantMessage_ToolUse はassistantメッセージ(tool_useコンテンツ)を正しくパースできることを確認する
func TestParseAssistantMessage_ToolUse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantName string
		wantErr  bool
	}{
		{
			name:     "Glob tool use",
			input:    `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_xxx","name":"Glob","input":{"pattern":"**/*.go"}}]}}`,
			wantType: "assistant",
			wantName: "Glob",
			wantErr:  false,
		},
		{
			name:     "Bash tool use",
			input:    `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_yyy","name":"Bash","input":{"command":"ls -la"}}]}}`,
			wantType: "assistant",
			wantName: "Bash",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg AssistantMessage
			err := json.Unmarshal([]byte(tt.input), &msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if msg.Type != tt.wantType {
					t.Errorf("AssistantMessage.Type = %v, want %v", msg.Type, tt.wantType)
				}
				if len(msg.Message.Content) == 0 {
					t.Errorf("AssistantMessage.Message.Content is empty")
					return
				}
				if msg.Message.Content[0].Type != "tool_use" {
					t.Errorf("Content[0].Type = %v, want tool_use", msg.Message.Content[0].Type)
				}
				if msg.Message.Content[0].Name != tt.wantName {
					t.Errorf("Content[0].Name = %v, want %v", msg.Message.Content[0].Name, tt.wantName)
				}
			}
		})
	}
}

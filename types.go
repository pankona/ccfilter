package main

import "encoding/json"

// Message はClaude CLIが出力するJSONメッセージの基本構造
type Message struct {
	Type string `json:"type"`
}

// SystemMessage はsystemタイプのメッセージ
type SystemMessage struct {
	Type              string `json:"type"`
	Subtype           string `json:"subtype"`
	Cwd               string `json:"cwd"`
	SessionID         string `json:"session_id"`
	Model             string `json:"model"`
	ClaudeCodeVersion string `json:"claude_code_version"`
}

// AssistantMessage はassistantタイプのメッセージ
type AssistantMessage struct {
	Type    string `json:"type"`
	Message struct {
		Content []Content `json:"content"`
	} `json:"message"`
}

// Content はassistantメッセージのコンテンツ
type Content struct {
	Type string `json:"type"` // "text" or "tool_use"

	// text タイプの場合
	Text string `json:"text,omitempty"`

	// tool_use タイプの場合
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// UserMessage はuserタイプのメッセージ (tool_result)
type UserMessage struct {
	Type    string `json:"type"`
	Message struct {
		Role    string       `json:"role"`
		Content []ToolResult `json:"content"`
	} `json:"message"`
}

// ToolResult はツール実行結果
type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ResultMessage はresultタイプのメッセージ (最終結果とメトリクス)
type ResultMessage struct {
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	IsError      bool    `json:"is_error"`
	Result       string  `json:"result"`
	DurationMs   int     `json:"duration_ms"`
	TotalCostUsd float64 `json:"total_cost_usd"`
	NumTurns     int     `json:"num_turns"`
	SessionID    string  `json:"session_id"`
}

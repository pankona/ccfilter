package main

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
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

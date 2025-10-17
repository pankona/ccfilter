package main

import "encoding/json"

// FilterConfig はフィルタリングの設定を保持する構造体
type FilterConfig struct {
	ShowSystem    bool
	ShowAssistant bool
	ShowTools     bool
	ShowResult    bool
	InfoLevel     string // "minimal", "standard", "verbose"
	ShowCost      bool
	ShowUsage     bool
	ShowTiming    bool
	Format        string // "text", "json", "compact"
	UseColor      bool
}

// NewFilterConfig はデフォルト設定でFilterConfigを作成
func NewFilterConfig() *FilterConfig {
	return &FilterConfig{
		ShowAssistant: true,
		ShowTools:     true,
		ShowResult:    true,
		InfoLevel:     "standard",
		Format:        "text",
		UseColor:      true,
	}
}

// parseMessageType はJSON行からメッセージタイプを判定
func parseMessageType(line string) (string, error) {
	var msg Message
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return "", err
	}
	return msg.Type, nil
}

// shouldDisplay はメッセージを表示すべきかどうかを判定
func shouldDisplay(msgType string, config *FilterConfig) bool {
	switch msgType {
	case "system":
		return config.ShowSystem
	case "assistant":
		return config.ShowAssistant
	case "user":
		return config.ShowTools
	case "result":
		return config.ShowResult
	default:
		return false
	}
}

// shouldDisplayContent は Content を表示すべきかどうかを判定
func shouldDisplayContent(content Content, config *FilterConfig) bool {
	switch content.Type {
	case "text":
		return config.ShowAssistant
	case "tool_use":
		return config.ShowTools
	default:
		return false
	}
}

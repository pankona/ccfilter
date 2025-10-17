package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatMessage はメッセージを人間が読みやすい形式にフォーマット
func formatMessage(msgType string, data []byte, config *FilterConfig) (string, error) {
	switch msgType {
	case "assistant":
		return formatAssistantMessage(data, config)
	case "user":
		return formatUserMessage(data, config)
	case "result":
		return formatResultMessage(data, config)
	default:
		return "", nil
	}
}

// formatAssistantMessage は AssistantMessage をフォーマット
func formatAssistantMessage(data []byte, config *FilterConfig) (string, error) {
	var msg AssistantMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", err
	}

	var output strings.Builder
	for _, content := range msg.Message.Content {
		if !shouldDisplayContent(content, config) {
			continue
		}

		switch content.Type {
		case "text":
			output.WriteString(content.Text)
			output.WriteString("\n")
		case "tool_use":
			formatted := formatToolUse(content, config)
			output.WriteString(formatted)
		}
	}

	return output.String(), nil
}

// formatUserMessage は UserMessage をフォーマット
func formatUserMessage(data []byte, config *FilterConfig) (string, error) {
	var msg UserMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", err
	}

	var output strings.Builder
	for _, result := range msg.Message.Content {
		if result.Type == "tool_result" {
			formatted := formatToolResult(result, config)
			output.WriteString(formatted)
		}
	}

	return output.String(), nil
}

// formatToolUse は tool_use コンテンツをフォーマット
func formatToolUse(content Content, config *FilterConfig) string {
	var output strings.Builder

	arrow := colorize("→", "cyan", config.UseColor)
	output.WriteString(arrow)
	output.WriteString(" ")

	toolName := colorize(content.Name, "blue", config.UseColor)
	output.WriteString(toolName)

	// minimal モードではツール名のみ
	if config.InfoLevel == "minimal" {
		output.WriteString("\n")
		return output.String()
	}

	// standard/verbose モードではパラメータも表示
	if len(content.Input) > 0 {
		params := extractMainParams(content.Name, content.Input)
		if params != "" {
			output.WriteString(": ")
			output.WriteString(params)
		}
	}

	output.WriteString("\n")
	return output.String()
}

// extractMainParams はツールの主要パラメータを抽出
func extractMainParams(toolName string, input []byte) string {
	var params map[string]interface{}
	if err := json.Unmarshal(input, &params); err != nil {
		return ""
	}

	// ツールごとに主要なパラメータを選択
	switch toolName {
	case "Glob":
		if pattern, ok := params["pattern"].(string); ok {
			return fmt.Sprintf("pattern=%q", pattern)
		}
	case "Bash":
		if command, ok := params["command"].(string); ok {
			return fmt.Sprintf("command=%q", command)
		}
	case "Read":
		if path, ok := params["file_path"].(string); ok {
			return fmt.Sprintf("file_path=%q", path)
		}
	case "Write", "Edit":
		if path, ok := params["file_path"].(string); ok {
			return fmt.Sprintf("file_path=%q", path)
		}
	case "Grep":
		if pattern, ok := params["pattern"].(string); ok {
			return fmt.Sprintf("pattern=%q", pattern)
		}
	}

	return ""
}

// formatToolResult は tool_result をフォーマット
func formatToolResult(result ToolResult, config *FilterConfig) string {
	var output strings.Builder

	arrow := colorize("←", "cyan", config.UseColor)
	output.WriteString(arrow)
	output.WriteString(" ")

	if result.IsError {
		errorText := colorize("Error:", "red", config.UseColor)
		output.WriteString(errorText)
		output.WriteString(" ")
	}

	// minimal モードでは1行のみ
	if config.InfoLevel == "minimal" {
		firstLine := strings.Split(result.Content, "\n")[0]
		output.WriteString(firstLine)
		output.WriteString("\n")
		return output.String()
	}

	// standard モードでは適度に省略
	maxLines := 5
	if config.InfoLevel == "verbose" {
		maxLines = -1 // 無制限
	}

	truncated := truncateOutput(result.Content, maxLines)
	output.WriteString(truncated)
	output.WriteString("\n")

	return output.String()
}

// truncateOutput は出力を指定行数で省略
func truncateOutput(s string, maxLines int) string {
	if maxLines < 0 {
		return s // 無制限
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}

	// 最初の maxLines 行を表示
	result := strings.Join(lines[:maxLines], "\n")
	omitted := len(lines) - maxLines
	result += fmt.Sprintf("\n... (%d more lines)", omitted)

	return result
}

// formatResultMessage は ResultMessage をフォーマット
func formatResultMessage(data []byte, config *FilterConfig) (string, error) {
	var msg ResultMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", err
	}

	var output strings.Builder

	// 区切り線
	separator := strings.Repeat("━", 40)
	coloredSeparator := colorize(separator, "gray", config.UseColor)

	output.WriteString("\n")
	output.WriteString(coloredSeparator)
	output.WriteString("\n")

	// 結果
	output.WriteString(msg.Result)
	output.WriteString("\n")

	// メトリクス (standard または verbose の場合)
	if config.InfoLevel != "minimal" {
		output.WriteString("\n")
		metrics := formatMetrics(msg, config)
		coloredMetrics := colorize(metrics, "gray", config.UseColor)
		output.WriteString(coloredMetrics)
		output.WriteString("\n")
	}

	output.WriteString(coloredSeparator)
	output.WriteString("\n")

	return output.String(), nil
}

// formatMetrics はメトリクス情報をフォーマット
func formatMetrics(msg ResultMessage, config *FilterConfig) string {
	var parts []string

	if config.ShowTiming || config.InfoLevel == "standard" || config.InfoLevel == "verbose" {
		parts = append(parts, fmt.Sprintf("Duration: %.1fs", float64(msg.DurationMs)/1000.0))
	}

	if config.ShowCost || config.InfoLevel == "standard" || config.InfoLevel == "verbose" {
		parts = append(parts, fmt.Sprintf("Cost: $%.4f", msg.TotalCostUsd))
	}

	parts = append(parts, fmt.Sprintf("Turns: %d", msg.NumTurns))

	return strings.Join(parts, " | ")
}

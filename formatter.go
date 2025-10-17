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
		if content.Type == "text" && shouldDisplayContent(content, config) {
			output.WriteString(content.Text)
			output.WriteString("\n")
		}
	}

	return output.String(), nil
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
	output.WriteString("\n")
	output.WriteString(separator)
	output.WriteString("\n")

	// 結果
	output.WriteString(msg.Result)
	output.WriteString("\n")

	// メトリクス (standard または verbose の場合)
	if config.InfoLevel != "minimal" {
		output.WriteString("\n")
		metrics := formatMetrics(msg, config)
		output.WriteString(metrics)
	}

	output.WriteString(separator)
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

# Phase 05: ツールフォーマッター

## 目標

tool_use と tool_result のフォーマット機能を TDD で実装する。

## 前提条件

- Phase 04 が完了していること
- text コンテンツのフォーマットが動作していること

## 実装内容

### 1. tool_use のフォーマット

AssistantMessage の tool_use コンテンツをフォーマットする。

#### テストケース (formatter_test.go)

```go
func TestFormatToolUse(t *testing.T)
```

- ツール名の表示
- 主要パラメータの表示 (standard モード)
- パラメータなし (minimal モード)
- 全パラメータ (verbose モード)

#### 実装 (formatter.go)

```go
// formatToolUse は tool_use コンテンツをフォーマット
func formatToolUse(content Content, config *FilterConfig) string {
    var output strings.Builder

    output.WriteString("→ ")
    output.WriteString(content.Name)

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
func extractMainParams(toolName string, input json.RawMessage) string {
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
```

### 2. tool_result のフォーマット

UserMessage の tool_result をフォーマットする。

#### テストケース (formatter_test.go)

```go
func TestFormatUserMessage_ToolResult(t *testing.T)
func TestTruncateOutput(t *testing.T)
```

- 成功結果の表示
- エラー結果の表示
- 長い出力の省略
- 複数行出力の表示

#### 実装 (formatter.go)

```go
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

// formatToolResult は tool_result をフォーマット
func formatToolResult(result ToolResult, config *FilterConfig) string {
    var output strings.Builder

    output.WriteString("← ")

    if result.IsError {
        output.WriteString("Error: ")
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
```

### 3. AssistantMessage の tool_use サポート

formatAssistantMessage に tool_use のサポートを追加。

#### 実装 (formatter.go)

```go
// formatAssistantMessage を更新
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
```

## TDD サイクル

### tool_use フォーマット

1. **Red**: TestFormatToolUse を書く
2. **Green**: formatToolUse と extractMainParams を実装
3. **Test**: テストを実行

### tool_result フォーマット

1. **Red**: TestFormatUserMessage_ToolResult を書く
2. **Green**: formatUserMessage と formatToolResult を実装
3. **Test**: テストを実行

### 出力省略

1. **Red**: TestTruncateOutput を書く
2. **Green**: truncateOutput を実装
3. **Test**: テストを実行

## テストコード例

### TestFormatToolUse

```go
func TestFormatToolUse(t *testing.T) {
    tests := []struct {
        name   string
        content Content
        config FilterConfig
        want   string
    }{
        {
            name: "Glob with standard config",
            content: Content{
                Type:  "tool_use",
                Name:  "Glob",
                Input: json.RawMessage(`{"pattern":"**/*.go"}`),
            },
            config: FilterConfig{InfoLevel: "standard"},
            want:   "→ Glob: pattern=\"**/*.go\"\n",
        },
        {
            name: "Bash with minimal config",
            content: Content{
                Type: "tool_use",
                Name: "Bash",
                Input: json.RawMessage(`{"command":"ls -la"}`),
            },
            config: FilterConfig{InfoLevel: "minimal"},
            want:   "→ Bash\n",
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
```

### TestTruncateOutput

```go
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
```

## 完了条件

- [ ] TestFormatToolUse が実装され、パスする
- [ ] TestFormatUserMessage_ToolResult が実装され、パスする
- [ ] TestTruncateOutput が実装され、パスする
- [ ] formatAssistantMessage が tool_use をサポートする
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add tool use and tool result formatters

TDD approach でツール関連のフォーマッター機能を実装:
- formatToolUse: tool_use コンテンツのフォーマット
  - ツール名とパラメータの表示
  - extractMainParams: 主要パラメータの抽出
- formatUserMessage: tool_result のフォーマット
  - 成功/エラー結果の表示
  - truncateOutput: 長い出力の省略
- formatAssistantMessage を tool_use 対応に更新

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 06: システムメッセージフォーマッター (オプション)
Phase 07: カラー出力

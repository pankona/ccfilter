# Phase 04: 基本フォーマッター (text コンテンツ)

## 目標

AssistantMessage の text コンテンツと ResultMessage の基本的なフォーマット機能を TDD で実装する。

## 前提条件

- Phase 03 が完了していること
- フィルタリングロジックが動作していること

## 実装内容

### 1. formatter.go の基本構造

#### 実装 (formatter.go)

```go
package main

import (
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
```

### 2. AssistantMessage の text フォーマット

#### テストケース (formatter_test.go)

```go
func TestFormatAssistantMessage_Text(t *testing.T)
```

- シンプルなテキストのフォーマット
- 複数行テキストのフォーマット
- 日本語テキストのフォーマット

#### 実装 (formatter.go)

```go
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
```

### 3. ResultMessage の基本フォーマット

#### テストケース (formatter_test.go)

```go
func TestFormatResultMessage(t *testing.T)
```

- 成功結果のフォーマット
- エラー結果のフォーマット
- 区切り線の表示

#### 実装 (formatter.go)

```go
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

    if config.ShowTiming || config.InfoLevel == "verbose" {
        parts = append(parts, fmt.Sprintf("Duration: %.1fs", float64(msg.DurationMs)/1000.0))
    }

    if config.ShowCost || config.InfoLevel == "verbose" {
        parts = append(parts, fmt.Sprintf("Cost: $%.4f", msg.TotalCostUsd))
    }

    parts = append(parts, fmt.Sprintf("Turns: %d", msg.NumTurns))

    return strings.Join(parts, " | ")
}
```

## TDD サイクル

### text フォーマット

1. **Red**: TestFormatAssistantMessage_Text を書く
2. **Green**: formatAssistantMessage を実装
3. **Test**: テストを実行

### Result フォーマット

1. **Red**: TestFormatResultMessage を書く
2. **Green**: formatResultMessage を実装
3. **Test**: テストを実行

## テストコード例

### TestFormatAssistantMessage_Text

```go
func TestFormatAssistantMessage_Text(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        config FilterConfig
        want   string
    }{
        {
            name:  "simple text",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, World!"}]}}`,
            config: FilterConfig{ShowAssistant: true},
            want:  "Hello, World!\n",
        },
        {
            name:  "multiline text",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Line 1\nLine 2\nLine 3"}]}}`,
            config: FilterConfig{ShowAssistant: true},
            want:  "Line 1\nLine 2\nLine 3\n",
        },
        {
            name:  "japanese text",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"こんにちは、世界！"}]}}`,
            config: FilterConfig{ShowAssistant: true},
            want:  "こんにちは、世界！\n",
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
```

### TestFormatResultMessage

```go
func TestFormatResultMessage(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        config FilterConfig
        want   []string // 含まれるべき文字列のリスト
    }{
        {
            name:  "success with standard config",
            input: `{"type":"result","subtype":"success","result":"完了しました","duration_ms":5000,"total_cost_usd":0.0123,"num_turns":3}`,
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
            name:  "minimal config",
            input: `{"type":"result","subtype":"success","result":"完了しました","duration_ms":5000,"total_cost_usd":0.0123,"num_turns":3}`,
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
```

## 完了条件

- [ ] formatter.go が作成されている
- [ ] formatter_test.go が作成されている
- [ ] TestFormatAssistantMessage_Text が実装され、パスする
- [ ] TestFormatResultMessage が実装され、パスする
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add basic formatter for text content and results

TDD approach で基本的なフォーマッター機能を実装:
- formatAssistantMessage: text コンテンツのフォーマット
  - 複数行テキストのサポート
  - 日本語テキストのサポート
- formatResultMessage: 最終結果のフォーマット
  - 区切り線の表示
  - メトリクス情報の表示 (Duration, Cost, Turns)
- formatMetrics: メトリクス情報のフォーマット

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 05: ツールフォーマッター

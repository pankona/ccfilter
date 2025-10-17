# Phase 06: カラー出力

## 目標

ANSIカラーコードによる色付き出力機能を TDD で実装する。

## 前提条件

- Phase 05 が完了していること
- すべてのフォーマッター機能が動作していること

## 実装内容

### 1. color.go の基本構造

#### テストケース (color_test.go)

```go
func TestColorize(t *testing.T)
func TestColorize_Disabled(t *testing.T)
```

- 各色のカラーコード
- カラー無効時の動作

#### 実装 (color.go)

```go
package main

import "fmt"

// ANSIカラーコード
const (
    ColorReset  = "\x1b[0m"
    ColorGreen  = "\x1b[32m"
    ColorYellow = "\x1b[33m"
    ColorBlue   = "\x1b[34m"
    ColorRed    = "\x1b[31m"
    ColorCyan   = "\x1b[36m"
    ColorGray   = "\x1b[90m"
)

// colorize はテキストを指定色で装飾
func colorize(text, color string, enabled bool) string {
    if !enabled {
        return text
    }

    switch color {
    case "green":
        return ColorGreen + text + ColorReset
    case "yellow":
        return ColorYellow + text + ColorReset
    case "blue":
        return ColorBlue + text + ColorReset
    case "red":
        return ColorRed + text + ColorReset
    case "cyan":
        return ColorCyan + text + ColorReset
    case "gray":
        return ColorGray + text + ColorReset
    default:
        return text
    }
}
```

### 2. フォーマッターへのカラー統合

各フォーマッター関数にカラー出力を追加。

#### 実装 (formatter.go)

```go
// formatToolUse を更新
func formatToolUse(content Content, config *FilterConfig) string {
    var output strings.Builder

    arrow := colorize("→", "cyan", config.UseColor)
    output.WriteString(arrow)
    output.WriteString(" ")

    toolName := colorize(content.Name, "blue", config.UseColor)
    output.WriteString(toolName)

    // パラメータ部分は変更なし
    if config.InfoLevel != "minimal" && len(content.Input) > 0 {
        params := extractMainParams(content.Name, content.Input)
        if params != "" {
            output.WriteString(": ")
            output.WriteString(params)
        }
    }

    output.WriteString("\n")
    return output.String()
}

// formatToolResult を更新
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

    // 残りの実装は変更なし
    // ...
}

// formatResultMessage を更新
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

    // メトリクス
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
```

### 3. FilterConfig へのカラー設定追加

NewFilterConfig を更新してカラー設定をサポート。

#### 実装 (filter.go)

```go
// NewFilterConfig を更新
func NewFilterConfig() *FilterConfig {
    return &FilterConfig{
        ShowAssistant: true,
        ShowTools:     true,
        ShowResult:    true,
        InfoLevel:     "standard",
        Format:        "text",
        UseColor:      isTerminal(), // ターミナルの場合のみデフォルトで有効
    }
}

// isTerminal は標準出力がターミナルかどうかを判定
func isTerminal() bool {
    // 簡易実装: 環境変数 TERM が設定されているか
    term := os.Getenv("TERM")
    return term != "" && term != "dumb"
}
```

## TDD サイクル

### カラー関数

1. **Red**: TestColorize を書く
2. **Green**: colorize を実装
3. **Test**: テストを実行

### カラー無効

1. **Red**: TestColorize_Disabled を書く
2. **Green**: enabled パラメータを実装
3. **Test**: テストを実行

### フォーマッター統合

1. **Red**: 既存のフォーマッターテストを更新 (カラー無効でテスト)
2. **Green**: 各フォーマッター関数を更新
3. **Test**: テストを実行

## テストコード例

### TestColorize

```go
func TestColorize(t *testing.T) {
    tests := []struct {
        name    string
        text    string
        color   string
        enabled bool
        want    string
    }{
        {
            name:    "green with color enabled",
            text:    "success",
            color:   "green",
            enabled: true,
            want:    "\x1b[32msuccess\x1b[0m",
        },
        {
            name:    "red with color enabled",
            text:    "error",
            color:   "red",
            enabled: true,
            want:    "\x1b[31merror\x1b[0m",
        },
        {
            name:    "color disabled",
            text:    "text",
            color:   "green",
            enabled: false,
            want:    "text",
        },
        {
            name:    "unknown color",
            text:    "text",
            color:   "unknown",
            enabled: true,
            want:    "text",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := colorize(tt.text, tt.color, tt.enabled)
            if got != tt.want {
                t.Errorf("colorize() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

### フォーマッターテストの更新

既存のフォーマッターテストで `UseColor: false` を設定して、カラーコードが含まれないことを確認。

```go
func TestFormatToolUse(t *testing.T) {
    tests := []struct {
        name    string
        content Content
        config  FilterConfig
        want    string
    }{
        {
            name: "Glob without color",
            content: Content{
                Type:  "tool_use",
                Name:  "Glob",
                Input: json.RawMessage(`{"pattern":"**/*.go"}`),
            },
            config: FilterConfig{
                InfoLevel: "standard",
                UseColor:  false, // カラー無効
            },
            want: "→ Glob: pattern=\"**/*.go\"\n",
        },
    }
    // ... テスト実装
}
```

## 完了条件

- [ ] color.go が作成されている
- [ ] color_test.go が作成されている
- [ ] TestColorize が実装され、パスする
- [ ] formatToolUse がカラー出力をサポートする
- [ ] formatToolResult がカラー出力をサポートする
- [ ] formatResultMessage がカラー出力をサポートする
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add ANSI color output support

TDD approach でカラー出力機能を実装:
- color.go: ANSIカラーコードによる装飾
  - colorize: テキストを指定色で装飾
  - 色の種類: green, yellow, blue, red, cyan, gray
- フォーマッターにカラー統合:
  - ツール名: blue
  - 矢印: cyan
  - エラー: red
  - メトリクス/区切り線: gray
- UseColor 設定でカラー出力の有効/無効を切り替え

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 07: CLI引数パース

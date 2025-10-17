# Phase 07: CLI引数パース

## 目標

コマンドライン引数のパース機能を TDD で実装する。

## 前提条件

- Phase 06 が完了していること
- カラー出力機能が動作していること

## 実装内容

### 1. main.go の基本構造

#### 実装 (main.go)

```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    config, err := parseArgs()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := run(config); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// run はメイン処理を実行 (Phase 08 で実装)
func run(config *FilterConfig) error {
    // TODO: Phase 08 で実装
    return nil
}
```

### 2. 引数パース関数

#### テストケース (main_test.go)

```go
func TestParseArgs(t *testing.T)
func TestParseArgs_Help(t *testing.T)
```

- デフォルト設定
- 各オプションの設定
- ヘルプ表示

#### 実装 (main.go)

```go
// parseArgs はコマンドライン引数をパース
func parseArgs() (*FilterConfig, error) {
    config := NewFilterConfig()

    // フラグ定義
    var (
        showSystem    = flag.Bool("system", false, "Show system messages")
        showAssistant = flag.Bool("assistant", false, "Show only assistant messages")
        showTools     = flag.Bool("tools", false, "Show only tool messages")
        showResult    = flag.Bool("result", false, "Show only result messages")
        showAll       = flag.Bool("all", false, "Show all messages")

        minimal   = flag.Bool("minimal", false, "Show minimal information")
        verbose   = flag.Bool("verbose", false, "Show verbose information")
        verboseV  = flag.Bool("v", false, "Show verbose information (short)")

        showCost    = flag.Bool("show-cost", false, "Always show cost information")
        showUsage   = flag.Bool("show-usage", false, "Always show token usage")
        showTiming  = flag.Bool("show-timing", false, "Always show timing information")
        showSession = flag.Bool("show-session", false, "Always show session ID")

        noColor = flag.Bool("no-color", false, "Disable color output")
        color   = flag.Bool("color", false, "Force enable color output")

        format = flag.String("format", "text", "Output format (text|json|compact)")

        help = flag.Bool("help", false, "Show help message")
        h    = flag.Bool("h", false, "Show help message (short)")
    )

    flag.Parse()

    // ヘルプ表示
    if *help || *h {
        printHelp()
        os.Exit(0)
    }

    // メッセージタイプフィルタ
    if *showAll {
        config.ShowSystem = true
        config.ShowAssistant = true
        config.ShowTools = true
        config.ShowResult = true
    } else {
        if *showSystem {
            config.ShowSystem = true
        }
        if *showAssistant {
            // assistant のみ表示
            config.ShowAssistant = true
            config.ShowTools = false
            config.ShowResult = false
        }
        if *showTools {
            // tools のみ表示
            config.ShowAssistant = false
            config.ShowTools = true
            config.ShowResult = false
        }
        if *showResult {
            // result のみ表示
            config.ShowAssistant = false
            config.ShowTools = false
            config.ShowResult = true
        }
    }

    // 情報レベル
    if *minimal {
        config.InfoLevel = "minimal"
    } else if *verbose || *verboseV {
        config.InfoLevel = "verbose"
    }

    // 個別表示オプション
    if *showCost {
        config.ShowCost = true
    }
    if *showUsage {
        config.ShowUsage = true
    }
    if *showTiming {
        config.ShowTiming = true
    }

    // カラー設定
    if *noColor {
        config.UseColor = false
    } else if *color {
        config.UseColor = true
    }

    // フォーマット
    config.Format = *format
    if config.Format != "text" && config.Format != "json" && config.Format != "compact" {
        return nil, fmt.Errorf("invalid format: %s (must be text, json, or compact)", config.Format)
    }

    return config, nil
}

// printHelp はヘルプメッセージを表示
func printHelp() {
    fmt.Fprintf(os.Stderr, `Usage: ccfilter [options]

ccfilter filters Claude CLI stream-json output for human readability.

Usage:
  claude -p --verbose --output-format=stream-json <prompt> | ccfilter [options]

Message Type Filters:
  --system          Show system messages
  --assistant       Show only assistant messages
  --tools           Show only tool messages (tool_use and tool_result)
  --result          Show only result messages
  --all             Show all messages

Information Level:
  --minimal, -m     Show minimal information
  --verbose, -v     Show verbose information
  (default is standard level)

Additional Information:
  --show-cost       Always show cost information
  --show-usage      Always show token usage
  --show-timing     Always show timing information
  --show-session    Always show session ID

Output Format:
  --format=FORMAT   Output format (text|json|compact) [default: text]
  --color           Force enable color output
  --no-color        Disable color output

Other:
  --help, -h        Show this help message

Examples:
  # Default: show assistant messages, tools, and results
  claude -p --verbose --output-format=stream-json "hello" | ccfilter

  # Show only tool usage
  claude -p --verbose --output-format=stream-json "list files" | ccfilter --tools

  # Minimal output
  claude -p --verbose --output-format=stream-json "hello" | ccfilter --minimal

  # Verbose with cost and timing
  claude -p --verbose --output-format=stream-json "hello" | ccfilter --verbose --show-cost --show-timing
`)
}
```

## TDD サイクル

### 引数パース

1. **Red**: TestParseArgs を書く
2. **Green**: parseArgs を実装
3. **Test**: テストを実行

### ヘルプ表示

1. **Red**: TestParseArgs_Help を書く
2. **Green**: printHelp を実装
3. **Test**: テストを実行

## テストコード例

### TestParseArgs

```go
func TestParseArgs(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        want     FilterConfig
        wantErr  bool
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
                UseColor:      true,
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
                UseColor:      true,
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
                UseColor:      true,
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
                UseColor:      false,
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
            // フラグをリセット
            flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
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
                if got.InfoLevel != tt.want.InfoLevel {
                    t.Errorf("InfoLevel = %v, want %v", got.InfoLevel, tt.want.InfoLevel)
                }
                // ... 他のフィールドも検証
            }
        })
    }
}
```

## 完了条件

- [ ] main.go が作成されている
- [ ] main_test.go が作成されている (引数パーステスト)
- [ ] TestParseArgs が実装され、パスする
- [ ] printHelp が実装されている
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add CLI argument parsing

TDD approach でコマンドライン引数パース機能を実装:
- parseArgs: flag パッケージで引数をパース
  - メッセージタイプフィルタ: --system, --assistant, --tools, --result, --all
  - 情報レベル: --minimal, --verbose
  - 追加情報: --show-cost, --show-usage, --show-timing
  - カラー設定: --color, --no-color
  - フォーマット: --format
- printHelp: ヘルプメッセージの表示
- main: エントリーポイントの実装

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 08: メインロジック統合

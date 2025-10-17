# Phase 08: メインロジック統合

## 目標

標準入力からのJSON読み取り、フィルタリング、フォーマット、出力を統合する。

## 前提条件

- Phase 07 が完了していること
- CLI引数パースが動作していること

## 実装内容

### 1. run 関数の実装

main.go の run 関数を実装し、すべての機能を統合する。

#### テストケース (main_test.go)

```go
func TestRun_Integration(t *testing.T)
```

- 標準入力からの読み取り
- フィルタリングとフォーマットの統合
- エラーハンドリング

#### 実装 (main.go)

```go
import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
)

// run はメイン処理を実行
func run(config *FilterConfig) error {
    return processInput(os.Stdin, os.Stdout, config)
}

// processInput は入力を処理して出力
func processInput(input io.Reader, output io.Writer, config *FilterConfig) error {
    scanner := bufio.NewScanner(input)

    for scanner.Scan() {
        line := scanner.Text()

        // 空行はスキップ
        if line == "" {
            continue
        }

        // メッセージタイプを判定
        msgType, err := parseMessageType(line)
        if err != nil {
            // JSONパースエラーは警告を出してスキップ
            fmt.Fprintf(os.Stderr, "Warning: failed to parse JSON: %v\n", err)
            continue
        }

        // フィルタリング
        if !shouldDisplay(msgType, config) {
            continue
        }

        // フォーマット
        formatted, err := formatMessage(msgType, []byte(line), config)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to format message: %v\n", err)
            continue
        }

        // 出力
        if formatted != "" {
            fmt.Fprint(output, formatted)
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("failed to read input: %w", err)
    }

    return nil
}
```

### 2. formatMessage の統合

formatter.go に formatMessage 関数を実装。

#### 実装 (formatter.go)

```go
// formatMessage はメッセージタイプに応じてフォーマット
func formatMessage(msgType string, data []byte, config *FilterConfig) (string, error) {
    switch msgType {
    case "system":
        return formatSystemMessage(data, config)
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

// formatSystemMessage は SystemMessage をフォーマット (簡易実装)
func formatSystemMessage(data []byte, config *FilterConfig) (string, error) {
    if config.InfoLevel == "minimal" {
        return "", nil // minimal では非表示
    }

    var msg SystemMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return "", err
    }

    var output strings.Builder

    if msg.Subtype == "init" {
        output.WriteString(colorize("Session initialized\n", "gray", config.UseColor))

        if config.InfoLevel == "verbose" {
            output.WriteString(fmt.Sprintf("  Model: %s\n", msg.Model))
            output.WriteString(fmt.Sprintf("  Working Directory: %s\n", msg.Cwd))
            output.WriteString(fmt.Sprintf("  Version: %s\n", msg.ClaudeCodeVersion))
        }
        output.WriteString("\n")
    }

    return output.String(), nil
}
```

## TDD サイクル

### processInput 実装

1. **Red**: TestRun_Integration を書く
2. **Green**: processInput を実装
3. **Test**: テストを実行

### エラーハンドリング

1. **Red**: 不正なJSONのテストケースを追加
2. **Green**: エラーハンドリングを実装
3. **Test**: テストを実行

## テストコード例

### TestRun_Integration

```go
func TestRun_Integration(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        config     FilterConfig
        wantOutput []string // 含まれるべき文字列
        wantErr    bool
    }{
        {
            name: "simple text message",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
            config: FilterConfig{
                ShowAssistant: true,
                ShowResult:    true,
                InfoLevel:     "standard",
                UseColor:      false,
            },
            wantOutput: []string{"Hello", "Done", "Duration"},
            wantErr:    false,
        },
        {
            name: "tool use and result",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Searching..."}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Glob","input":{"pattern":"*.go"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"main.go"}]}}`,
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                InfoLevel:     "standard",
                UseColor:      false,
            },
            wantOutput: []string{"Searching", "→ Glob", "← main.go"},
            wantErr:    false,
        },
        {
            name: "filtering - only result",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
            config: FilterConfig{
                ShowAssistant: false,
                ShowResult:    true,
                InfoLevel:     "standard",
                UseColor:      false,
            },
            wantOutput: []string{"Done"},
            wantErr:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := strings.NewReader(tt.input)
            var output bytes.Buffer

            err := processInput(input, &output, &tt.config)

            if (err != nil) != tt.wantErr {
                t.Errorf("processInput() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            result := output.String()
            for _, want := range tt.wantOutput {
                if !strings.Contains(result, want) {
                    t.Errorf("processInput() output does not contain %q\nGot: %s", want, result)
                }
            }
        })
    }
}
```

### TestProcessInput_ErrorHandling

```go
func TestProcessInput_ErrorHandling(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {
            name:  "invalid json",
            input: `{invalid json}`,
        },
        {
            name:  "empty lines",
            input: "\n\n\n",
        },
        {
            name: "mixed valid and invalid",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"OK"}]}}
{invalid}
{"type":"result","subtype":"success","result":"Done","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := strings.NewReader(tt.input)
            var output bytes.Buffer
            config := NewFilterConfig()
            config.UseColor = false

            // エラーで終了しないことを確認
            err := processInput(input, &output, config)
            if err != nil {
                t.Errorf("processInput() should not return error for invalid JSON, got: %v", err)
            }
        })
    }
}
```

## 完了条件

- [ ] processInput が実装されている
- [ ] formatMessage が実装されている
- [ ] formatSystemMessage が実装されている
- [ ] TestRun_Integration が実装され、パスする
- [ ] TestProcessInput_ErrorHandling が実装され、パスする
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Integrate main processing logic

TDD approach でメインロジックを統合:
- processInput: 標準入力からJSON読み取り、フィルタリング、フォーマット
  - bufio.Scanner で行単位読み取り
  - エラーハンドリング (不正なJSONをスキップ)
- formatMessage: メッセージタイプに応じたフォーマット振り分け
- formatSystemMessage: system メッセージのフォーマット
- run: main 関数から呼び出されるエントリーポイント

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 09: testdata 整備とエンドツーエンドテスト

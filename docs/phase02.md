# Phase 02: UserMessage と ResultMessage のパース

## 目標

UserMessage (tool_result) と ResultMessage の JSON パース機能を TDD で実装する。

## 前提条件

- Phase 01 が完了していること
- すべてのテストがパスしていること

## 実装内容

### 1. UserMessage のパース

UserMessage は tool_result を含むメッセージ。

#### テストケース (types_test.go)

```go
func TestParseUserMessage_ToolResult(t *testing.T)
```

- 成功した tool_result のパース
- エラーを含む tool_result のパース
- is_error フラグの検証

#### 実装 (types.go)

```go
type UserMessage struct {
    Type    string `json:"type"`
    Message struct {
        Role    string       `json:"role"`
        Content []ToolResult `json:"content"`
    } `json:"message"`
}

type ToolResult struct {
    Type      string `json:"type"`
    ToolUseID string `json:"tool_use_id"`
    Content   string `json:"content"`
    IsError   bool   `json:"is_error,omitempty"`
}
```

### 2. ResultMessage のパース

ResultMessage は最終結果とメトリクスを含むメッセージ。

#### テストケース (types_test.go)

```go
func TestParseResultMessage(t *testing.T)
```

- success subtype のパース
- error subtype のパース
- コスト、実行時間、ターン数の検証

#### 実装 (types.go)

```go
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
```

### 3. testdata の準備

実際の Claude 出力をテストデータとして保存。

```bash
# /tmp/claude_output_sample.json を testdata にコピー
cp /tmp/claude_output_sample.json testdata/permission_denied.json
```

## TDD サイクル

### UserMessage

1. **Red**: TestParseUserMessage_ToolResult を書く
2. **Red**: テストを実行して失敗を確認
3. **Green**: UserMessage と ToolResult を実装
4. **Green**: テストを実行して成功を確認
5. **Test**: すべてのテストを実行

### ResultMessage

1. **Red**: TestParseResultMessage を書く
2. **Red**: テストを実行して失敗を確認
3. **Green**: ResultMessage を実装
4. **Green**: テストを実行して成功を確認
5. **Test**: すべてのテストを実行

## テストコード例

### UserMessage のテスト

```go
func TestParseUserMessage_ToolResult(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        wantType      string
        wantContent   string
        wantIsError   bool
        wantErr       bool
    }{
        {
            name:        "successful tool result",
            input:       `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_xxx","content":"No files found","is_error":false}]}}`,
            wantType:    "user",
            wantContent: "No files found",
            wantIsError: false,
            wantErr:     false,
        },
        {
            name:        "error tool result",
            input:       `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_yyy","content":"permission denied","is_error":true}]}}`,
            wantType:    "user",
            wantContent: "permission denied",
            wantIsError: true,
            wantErr:     false,
        },
    }
    // ... テスト実装
}
```

### ResultMessage のテスト

```go
func TestParseResultMessage(t *testing.T) {
    tests := []struct {
        name         string
        input        string
        wantSubtype  string
        wantCost     float64
        wantDuration int
        wantErr      bool
    }{
        {
            name:         "success result",
            input:        `{"type":"result","subtype":"success","is_error":false,"result":"完了しました","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}`,
            wantSubtype:  "success",
            wantCost:     0.01,
            wantDuration: 1000,
            wantErr:      false,
        },
    }
    // ... テスト実装
}
```

## 完了条件

- [ ] TestParseUserMessage_ToolResult が実装され、パスする
- [ ] TestParseResultMessage が実装され、パスする
- [ ] testdata/permission_denied.json が準備されている
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add UserMessage and ResultMessage parsing

TDD approach で UserMessage と ResultMessage のパース機能を実装:
- UserMessage: tool_result コンテンツのパース
  - ToolResult 構造体の追加
  - is_error フラグのサポート
- ResultMessage: 最終結果とメトリクスのパース
  - success/error subtype
  - コスト、実行時間、ターン数
- testdata/permission_denied.json を追加

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 03: フィルタリングロジック

# ccfilter 詳細設計

## 概要

`claude -p --verbose --output-format=stream-json` の出力を人間が読みやすい形にフィルタリングするCLIツール。

### 背景と目的

`claude -p` で実行すると、処理が完了するまで Claude Code の内部動作 (どんな考えでどういう操作をしたのか、どのツールを使ったのか) がわからない。待っている間は進行状況が見えないため、正しい方向に進んでいるのかを確認できない。

このツールは、`stream-json` 出力をリアルタイムでフィルタリングし、通常のインタラクティブモードと同様の進行状況を標準出力に表示することで、この問題を解決する。

## 入力フォーマット

### JSON Stream 構造

Claude CLI は以下の4種類のJSONメッセージを出力する:

1. **system メッセージ** (`type: "system"`)

- `subtype: "init"`: セッション初期化情報
- フィールド: `cwd`, `session_id`, `tools`, `model`, `claude_code_version` など

2. **assistant メッセージ** (`type: "assistant"`)

- Claude の思考、説明、ツール呼び出し
- `content` 配列内に以下のタイプが含まれる:
  - `type: "text"`: Claude の思考や説明文
  - `type: "tool_use"`: ツール呼び出し (name, id, input)
- フィールド: `message` (role, content, usage など)

3. **user メッセージ** (`type: "user"`)

- ツール実行結果
- `content` 配列内に `type: "tool_result"` が含まれる
- フィールド: `tool_use_id`, `content`, `is_error`

4. **result メッセージ** (`type: "result"`)

- `subtype: "success"` または `"error"`
- フィールド: `result`, `duration_ms`, `total_cost_usd`, `usage`, `modelUsage` など

### 実行フロー例

```
1. system (init) → セッション開始
2. assistant (text) → "Globツールで検索します"
3. assistant (tool_use) → Glob実行: {"pattern": "**/*.go"}
4. user (tool_result) → "No files found"
5. assistant (text) → ".goファイルが見つかりませんでした"
6. assistant (tool_use) → Bash実行: {"command": "ls -la"}
7. user (tool_result) → "total 20\ndrwxr-xr-x..."
8. assistant (text) → "現在のディレクトリには..."
9. result (success) → 最終結果とメトリクス
```

### 入力例

```json
{"type":"system","subtype":"init","cwd":"/path","session_id":"xxx",...}
{"type":"assistant","message":{"content":[{"type":"text","text":"検索します"}],...}}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_xxx","name":"Glob","input":{"pattern":"**/*.go"}}],...}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_xxx","content":"No files found"}],...}}
{"type":"result","subtype":"success","result":"最終結果",...}
```

## コマンドライン引数設計

### 基本構文

```bash
claude -p --verbose --output-format=stream-json {prompt} | ccfilter [options]
```

### デフォルト動作

引数なしの場合、インタラクティブモードと同様の表示:

- Assistant の text (Claude の思考・説明)
- Tool use (ツール呼び出し: ツール名と主要パラメータ)
- Tool result (ツール実行結果: 成功/失敗と要約)
- Result (最終結果)

これにより、`claude -p` 実行中の進行状況をリアルタイムで確認できる。

### フィルタオプション

#### 1. メッセージタイプフィルタ

- `--system`, `-s`: system メッセージを表示
- `--assistant`, `-a`: assistant メッセージのみ表示
- `--tools`, `-t`: ツール呼び出しと結果のみ表示
- `--result`, `-r`: result メッセージのみ表示
- `--all`: すべてのメッセージを表示

#### 2. 情報レベルフィルタ

- `--minimal`, `-m`: 最小限の情報のみ表示
  - assistant の text のみ
  - ツール名のみ (パラメータなし)
  - result の最終結果のみ
- `--standard` (デフォルト): 標準的な情報を表示
  - assistant の text
  - ツール呼び出しと主要パラメータ
  - ツール結果の要約 (長い場合は省略)
  - result の最終結果とコスト、実行時間
- `--verbose`, `-v`: 詳細情報を表示
  - すべてのフィールドを整形して表示
  - ツール結果を全文表示

#### 3. 個別フィールド表示

- `--show-cost`: コスト情報を常に表示
- `--show-usage`: トークン使用量を常に表示
- `--show-session`: セッションIDを常に表示
- `--show-model`: モデル情報を常に表示
- `--show-timing`: 実行時間情報を常に表示

#### 4. 出力フォーマット

- `--format=text` (デフォルト): 人間が読みやすいテキスト形式
- `--format=json`: 整形されたJSON
- `--format=compact`: コンパクトなテキスト形式

#### 5. その他

- `--color`, `--no-color`: カラー出力の有効/無効
- `--timestamp`: タイムスタンプを表示
- `--help`, `-h`: ヘルプを表示

### 使用例

```bash
# デフォルト: インタラクティブモード相当の表示
claude -p --verbose --output-format=stream-json "List Go files" | ccfilter

# ツールの使用状況のみを追跡
claude -p --verbose --output-format=stream-json "Implement feature" | ccfilter --tools

# 最小限の情報のみ表示
claude -p --verbose --output-format=stream-json "hello" | ccfilter --minimal

# コストと実行時間を含めて表示
claude -p --verbose --output-format=stream-json "hello" | ccfilter --show-cost --show-timing

# すべての詳細情報を表示 (デバッグ用)
claude -p --verbose --output-format=stream-json "hello" | ccfilter --verbose
```

## フィルタリングロジック

### 処理フロー

1. **入力読み取り**

- 標準入力から行単位でJSONを読み取る
- 各行を個別のJSONオブジェクトとしてパース

2. **メッセージ分類**

- `type` フィールドでメッセージタイプを判定
- `subtype` フィールドでサブタイプを判定

3. **フィルタリング**

- コマンドライン引数に基づいて表示するメッセージを選択
- 各メッセージタイプごとに必要なフィールドを抽出

4. **整形出力**

- 選択された情報を人間が読みやすい形式で出力

### データ構造

```go
package main

type Message struct {
    Type    string          `json:"type"`
    Subtype string          `json:"subtype,omitempty"`
    Raw     json.RawMessage `json:"-"`
}

type SystemMessage struct {
    Type               string   `json:"type"`
    Subtype            string   `json:"subtype"`
    Cwd                string   `json:"cwd"`
    SessionID          string   `json:"session_id"`
    Model              string   `json:"model"`
    ClaudeCodeVersion  string   `json:"claude_code_version"`
    Tools              []string `json:"tools"`
    // ... 他のフィールド
}

type AssistantMessage struct {
    Type    string `json:"type"`
    Message struct {
        Model   string `json:"model"`
        ID      string `json:"id"`
        Content []Content `json:"content"`
        Usage   struct {
            InputTokens              int `json:"input_tokens"`
            CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
            CacheReadInputTokens     int `json:"cache_read_input_tokens"`
            OutputTokens             int `json:"output_tokens"`
        } `json:"usage"`
    } `json:"message"`
}

type Content struct {
    Type string `json:"type"` // "text" or "tool_use"

    // text タイプの場合
    Text string `json:"text,omitempty"`

    // tool_use タイプの場合
    ID    string          `json:"id,omitempty"`
    Name  string          `json:"name,omitempty"`
    Input json.RawMessage `json:"input,omitempty"`
}

type UserMessage struct {
    Type    string `json:"type"`
    Message struct {
        Role    string          `json:"role"`
        Content []ToolResult    `json:"content"`
    } `json:"message"`
}

type ToolResult struct {
    Type       string `json:"type"` // "tool_result"
    ToolUseID  string `json:"tool_use_id"`
    Content    string `json:"content"`
    IsError    bool   `json:"is_error,omitempty"`
}

type ResultMessage struct {
    Type          string  `json:"type"`
    Subtype       string  `json:"subtype"`
    IsError       bool    `json:"is_error"`
    Result        string  `json:"result"`
    DurationMs    int     `json:"duration_ms"`
    TotalCostUsd  float64 `json:"total_cost_usd"`
    SessionID     string  `json:"session_id"`
    // ... 他のフィールド
}

type FilterConfig struct {
    ShowSystem    bool
    ShowAssistant bool
    ShowTools     bool
    ShowResult    bool
    InfoLevel     string // "minimal", "standard", "verbose"
    ShowCost      bool
    ShowUsage     bool
    ShowSession   bool
    ShowModel     bool
    ShowTiming    bool
    Format        string // "text", "json", "compact"
    UseColor      bool
    ShowTimestamp bool
}
```

## 出力フォーマット設計

### テキストフォーマット (デフォルト)

デフォルトモードでは、インタラクティブモードと同様の表示を目指す。

#### System メッセージ (--system 指定時のみ)

```
Session initialized
Model: claude-sonnet-4-5-20250929
Working Directory: /path/to/dir
```

#### Assistant Text メッセージ

```
`**/*.go` パターンでGoファイルを検索します。
```

#### Tool Use メッセージ

```
→ Glob: pattern="**/*.go"
```

#### Tool Result メッセージ

```
← No files found
```

複数行の結果は省略表示:

```
← total 20
  drwxr-xr-x  4 pankona pankona 4096 Oct 17 23:50 .
  drwxr-xr-x 32 pankona pankona 4096 Oct 17 23:44 ..
  ... (3 more lines)
```

#### Result メッセージ

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
現在のディレクトリには `.go` ファイルは存在していません。

Duration: 11.2s | Cost: $0.0322 | Turns: 7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 完全な出力例 (デフォルトモード)

```
`**/*.go` パターンでGoファイルを検索します。

→ Glob: pattern="**/*.go"
← No files found

現在のディレクトリには `.go` ファイルが見つかりませんでした。
念のため、カレントディレクトリの内容を確認してみます。

→ Bash: ls -la
← total 20
  drwxr-xr-x  4 pankona pankona 4096 Oct 17 23:50 .
  drwxr-xr-x 32 pankona pankona 4096 Oct 17 23:44 ..
  ... (2 more lines)

現在のディレクトリには `.go` ファイルは存在していません。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
現在のディレクトリには `.go` ファイルは存在していません。
`go.mod` ファイルと `docs` ディレクトリ、そして `.git` ディレクトリのみがあります。

Duration: 11.2s | Cost: $0.0322 | Turns: 7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Minimal モード

Assistant の text と Result の result のみ表示:

```
応答テキストの内容

実行結果
```

### Verbose モード

すべてのフィールドを構造化して表示:

```
[SYSTEM] init
  type: system
  subtype: init
  cwd: /path/to/dir
  session_id: xxx
  tools: [Task, Bash, Glob, ...]
  model: claude-sonnet-4-5-20250929
  ...
```

### JSON フォーマット

整形されたJSONを出力:

```json
{
  "type": "assistant",
  "message": {
    "content": [{"type": "text", "text": "..."}],
    ...
  }
}
```

## 実装の詳細

### Package 構成

```
main package のみ
├── main.go          # エントリーポイント、CLI引数パース
├── filter.go        # フィルタリングロジック
├── formatter.go     # 出力フォーマット処理
├── types.go         # データ構造定義
└── color.go         # カラー出力ユーティリティ
```

### 主要な関数

```go
// main.go
func main()
func parseArgs() *FilterConfig

// filter.go
func filterMessages(config *FilterConfig) error
func parseMessage(line string) (Message, error)
func shouldDisplay(msg Message, config *FilterConfig) bool

// formatter.go
func formatMessage(msg Message, config *FilterConfig) string
func formatSystem(msg SystemMessage, config *FilterConfig) string
func formatAssistant(msg AssistantMessage, config *FilterConfig) string
func formatUser(msg UserMessage, config *FilterConfig) string
func formatResult(msg ResultMessage, config *FilterConfig) string
func formatToolUse(content Content, config *FilterConfig) string
func formatToolResult(result ToolResult, config *FilterConfig) string
func truncateOutput(s string, maxLines int) string

// color.go
func colorize(text, color string) string
```

### 依存パッケージ

標準ライブラリのみ使用:

- `encoding/json`: JSONパース
- `bufio`: 標準入力読み取り
- `flag`: コマンドライン引数パース
- `fmt`: 出力
- `os`: 標準入出力

## エラーハンドリング

1. **JSON パースエラー**: 不正な行をスキップし、標準エラー出力に警告
2. **不明なメッセージタイプ**: 警告を出力してスキップ
3. **入力なし**: 正常終了 (exit 0)
4. **引数エラー**: ヘルプを表示して終了 (exit 1)

## 将来の拡張性

1. **フィルタ条件の追加**

- 特定のツールのみ表示
- コスト閾値でフィルタ
- セッションIDでフィルタ

2. **出力形式の追加**

- Markdown形式
- CSV形式 (メトリクス抽出用)

3. **統計情報**

- `--stats`: セッション全体の統計を表示
- 合計コスト、合計トークン数など

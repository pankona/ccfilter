# ccfilter 実装プラン

## 概要

このドキュメントは、ccfilter を TDD アプローチで段階的に実装していくための全体プランです。

## 開発方針

### TDD サイクル

各フェーズで以下のサイクルを繰り返します:

1. **Red**: 失敗するテストを書く
2. **Green**: テストを通すための最小限のコードを実装
3. **Refactor**: 必要に応じてリファクタリング
4. **Test**: すべてのテストを実行してパスすることを確認
5. **Commit**: テストがパスしたらコミット

### フェーズの進め方

- 各フェーズは `docs/phaseNN.md` に定義されている
- フェーズの開始時にテストを実行し、現在の状態を確認
- フェーズの終了時にテストを実行し、すべてパスすることを確認
- テストがパスしたら Git コミットを作成

### コミットメッセージフォーマット

```
<概要>

<詳細な説明>

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## アーキテクチャ

### ファイル構成

```
ccfilter/
├── main.go           # エントリーポイント、CLI引数パース
├── types.go          # データ構造定義
├── filter.go         # フィルタリングロジック
├── formatter.go      # 出力フォーマット処理
├── color.go          # カラー出力ユーティリティ
├── types_test.go     # types.go のテスト
├── filter_test.go    # filter.go のテスト
├── formatter_test.go # formatter.go のテスト
├── color_test.go     # color.go のテスト
├── main_test.go      # インテグレーションテスト
└── testdata/         # テストデータ
    ├── simple.json
    ├── with_tools.json
    └── ...
```

### データフロー

```
標準入力
  ↓
JSON行を読み取り (filter.go)
  ↓
メッセージタイプを判定 (types.go)
  ↓
フィルタリング (filter.go)
  ↓
フォーマット (formatter.go)
  ↓
標準出力
```

## データ構造

### 主要な型

```go
// Message: 基本メッセージ構造
type Message struct {
    Type    string `json:"type"`
    Subtype string `json:"subtype,omitempty"`
}

// SystemMessage: system メッセージ
type SystemMessage struct {
    Type              string
    Subtype           string
    Model             string
    ClaudeCodeVersion string
    // ...
}

// AssistantMessage: assistant メッセージ
type AssistantMessage struct {
    Type    string
    Message struct {
        Content []Content
    }
}

// Content: text または tool_use
type Content struct {
    Type string // "text" or "tool_use"
    Text string // text の場合
    Name string // tool_use の場合
    ID   string // tool_use の場合
    Input json.RawMessage // tool_use の場合
}

// UserMessage: user メッセージ (tool_result)
type UserMessage struct {
    Type    string
    Message struct {
        Content []ToolResult
    }
}

// ToolResult: ツール実行結果
type ToolResult struct {
    Type      string
    ToolUseID string
    Content   string
    IsError   bool
}

// ResultMessage: 最終結果
type ResultMessage struct {
    Type         string
    Subtype      string
    Result       string
    DurationMs   int
    TotalCostUsd float64
    NumTurns     int
}

// FilterConfig: フィルタ設定
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
```

## テスト戦略

### ユニットテスト

- 各関数の入出力を検証
- テーブルドリブンテストを使用
- エッジケースを網羅

### インテグレーションテスト

- 実際の Claude 出力を使用
- `testdata/` ディレクトリにサンプルを配置
- エンドツーエンドの動作を検証

### テストデータ

実際の Claude CLI 出力を保存:

```bash
claude -p --verbose --output-format=stream-json "prompt" > testdata/sample.json
```

## フェーズ一覧

### Phase 01: 基本データ構造とパース (完了)

- Message, SystemMessage, AssistantMessage の基本パース
- text と tool_use コンテンツのパース

### Phase 02: UserMessage と ResultMessage のパース

- UserMessage (tool_result) のパース
- ResultMessage のパース
- テストデータの準備

### Phase 03: フィルタリングロジック

- メッセージタイプによるフィルタリング
- FilterConfig の実装
- shouldDisplay 関数の実装

### Phase 04: 基本フォーマッター (text コンテンツ)

- text コンテンツのフォーマット
- AssistantMessage の text 表示

### Phase 05: ツールフォーマッター

- tool_use のフォーマット
- tool_result のフォーマット
- ツール名とパラメータの表示

### Phase 06: Result フォーマッター

- 最終結果の表示
- コスト、実行時間の表示
- 区切り線の表示

### Phase 07: カラー出力

- ANSI カラーコードのサポート
- メッセージタイプごとの色分け
- --color / --no-color オプション

### Phase 08: CLI引数パース

- flag パッケージを使用
- FilterConfig の構築
- ヘルプメッセージ

### Phase 09: メインロジック統合

- 標準入力からの読み取り
- フィルタリングとフォーマットの統合
- エラーハンドリング

### Phase 10: インテグレーションテスト

- 実際の Claude 出力を使ったテスト
- testdata の整備
- エンドツーエンド検証

## 依存関係

### 標準ライブラリのみ使用

- `encoding/json`: JSONパース
- `bufio`: バッファ付き入出力
- `flag`: コマンドライン引数
- `fmt`: フォーマット出力
- `os`: 標準入出力
- `io`: インターフェース
- `strings`: 文字列操作

## コーディング規約

### 命名規則

- エクスポートされる型・関数: PascalCase
- パッケージ内部の関数: camelCase
- テスト関数: `Test<FunctionName>`
- テーブル名: `tests`
- テストケース名: わかりやすい日本語または英語

### コメント

- エクスポートされる型・関数には必ずコメントを付ける
- コメントは型・関数名で始める (godoc スタイル)

### エラーハンドリング

- エラーは呼び出し元に返す
- 標準エラー出力には警告メッセージを出力
- 致命的なエラーは `log.Fatal` ではなく `return err`

## 開発の進め方

1. `docs/phaseNN.md` を読む
2. フェーズ開始時にテストを実行
3. Red-Green-Refactor サイクルを回す
4. フェーズ終了時にテストを実行
5. すべてのテストがパスしたらコミット
6. 次のフェーズへ

## 参考資料

- [design.md](design.md): 詳細設計
- [test_design.md](test_design.md): テスト設計
- [prompt.md](prompt.md): 背景と要件

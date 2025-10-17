# Phase 09: testdata 整備とエンドツーエンドテスト

## 目標

実際のClaude出力を使ったテストデータを整備し、エンドツーエンドテストを実装する。

## 前提条件

- Phase 08 が完了していること
- 基本機能がすべて動作していること

## 実装内容

### 1. testdata ディレクトリの整備

実際のClaude出力をテストデータとして保存。

#### ディレクトリ構造

```
testdata/
├── simple.json              # シンプルなテキスト応答
├── with_tools.json          # ツール使用を含む
├── permission_denied.json   # 権限エラーを含む (既存)
├── multi_turn.json          # 複数ターンの会話
└── golden/
    ├── simple_default.txt
    ├── simple_minimal.txt
    ├── with_tools_default.txt
    └── with_tools_verbose.txt
```

#### テストデータ生成

```bash
# testdata ディレクトリを作成
mkdir -p testdata/golden

# シンプルな応答
claude -p --verbose --output-format=stream-json "Say hello" > testdata/simple.json

# ツール使用
claude -p --verbose --output-format=stream-json "List all Go files" > testdata/with_tools.json

# 既存のファイルをコピー
cp /tmp/claude_output_sample.json testdata/permission_denied.json
```

### 2. ゴールデンファイルテスト

期待される出力をゴールデンファイルとして保存し、実際の出力と比較。

#### テストケース (main_test.go)

```go
func TestEndToEnd_Golden(t *testing.T)
```

- 各設定での出力をゴールデンファイルと比較
- --update フラグでゴールデンファイルを更新

#### 実装 (main_test.go)

```go
var update = flag.Bool("update", false, "update golden files")

func TestEndToEnd_Golden(t *testing.T) {
    tests := []struct {
        name       string
        inputFile  string
        goldenFile string
        config     FilterConfig
    }{
        {
            name:       "simple with default config",
            inputFile:  "testdata/simple.json",
            goldenFile: "testdata/golden/simple_default.txt",
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                ShowResult:    true,
                InfoLevel:     "standard",
                UseColor:      false,
            },
        },
        {
            name:       "simple with minimal config",
            inputFile:  "testdata/simple.json",
            goldenFile: "testdata/golden/simple_minimal.txt",
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                ShowResult:    true,
                InfoLevel:     "minimal",
                UseColor:      false,
            },
        },
        {
            name:       "with tools default config",
            inputFile:  "testdata/with_tools.json",
            goldenFile: "testdata/golden/with_tools_default.txt",
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                ShowResult:    true,
                InfoLevel:     "standard",
                UseColor:      false,
            },
        },
        {
            name:       "with tools verbose config",
            inputFile:  "testdata/with_tools.json",
            goldenFile: "testdata/golden/with_tools_verbose.txt",
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                ShowResult:    true,
                InfoLevel:     "verbose",
                ShowCost:      true,
                ShowTiming:    true,
                UseColor:      false,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 入力ファイルを読み込み
            inputData, err := os.ReadFile(tt.inputFile)
            if err != nil {
                t.Fatalf("failed to read input file: %v", err)
            }

            // 処理を実行
            input := bytes.NewReader(inputData)
            var output bytes.Buffer
            if err := processInput(input, &output, &tt.config); err != nil {
                t.Fatalf("processInput() error = %v", err)
            }

            got := output.String()

            // --update フラグが指定されていればゴールデンファイルを更新
            if *update {
                if err := os.WriteFile(tt.goldenFile, []byte(got), 0644); err != nil {
                    t.Fatalf("failed to update golden file: %v", err)
                }
                t.Logf("Updated golden file: %s", tt.goldenFile)
                return
            }

            // ゴールデンファイルと比較
            want, err := os.ReadFile(tt.goldenFile)
            if err != nil {
                t.Fatalf("failed to read golden file: %v", err)
            }

            if got != string(want) {
                t.Errorf("output differs from golden file")
                t.Errorf("To update golden files, run: go test -update")
                t.Errorf("\nGot:\n%s\n\nWant:\n%s", got, string(want))

                // 差分を表示するために diff を使う (オプション)
                if err := writeDiff(t, got, string(want)); err == nil {
                    t.Logf("Diff saved to /tmp/ccfilter_diff.txt")
                }
            }
        })
    }
}

// writeDiff は差分を一時ファイルに書き出す
func writeDiff(t *testing.T, got, want string) error {
    gotFile := "/tmp/ccfilter_got.txt"
    wantFile := "/tmp/ccfilter_want.txt"

    if err := os.WriteFile(gotFile, []byte(got), 0644); err != nil {
        return err
    }
    if err := os.WriteFile(wantFile, []byte(want), 0644); err != nil {
        return err
    }

    return nil
}
```

### 3. エンドツーエンドテスト

実際の使用シナリオをテスト。

#### テストケース (main_test.go)

```go
func TestEndToEnd_Scenarios(t *testing.T)
```

- デフォルト設定での実行
- ツールのみ表示
- 最小限表示
- エラーケース

#### 実装例

```go
func TestEndToEnd_Scenarios(t *testing.T) {
    tests := []struct {
        name         string
        inputFile    string
        config       FilterConfig
        wantContains []string
        wantNotContains []string
    }{
        {
            name:      "default config shows main flow",
            inputFile: "testdata/with_tools.json",
            config:    *NewFilterConfig(),
            wantContains: []string{
                "→ Glob",
                "← ",
                "Duration:",
            },
            wantNotContains: []string{
                "Session initialized",
            },
        },
        {
            name:      "tools only",
            inputFile: "testdata/with_tools.json",
            config: FilterConfig{
                ShowTools: true,
                InfoLevel: "standard",
                UseColor:  false,
            },
            wantContains: []string{
                "→ ",
                "← ",
            },
            wantNotContains: []string{
                "Duration:",
                "Cost:",
            },
        },
        {
            name:      "minimal mode",
            inputFile: "testdata/with_tools.json",
            config: FilterConfig{
                ShowAssistant: true,
                ShowTools:     true,
                ShowResult:    true,
                InfoLevel:     "minimal",
                UseColor:      false,
            },
            wantNotContains: []string{
                "Duration:",
                "Cost:",
                "pattern=",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inputData, err := os.ReadFile(tt.inputFile)
            if err != nil {
                t.Fatalf("failed to read input file: %v", err)
            }

            input := bytes.NewReader(inputData)
            var output bytes.Buffer
            tt.config.UseColor = false // テスト時はカラー無効

            if err := processInput(input, &output, &tt.config); err != nil {
                t.Fatalf("processInput() error = %v", err)
            }

            result := output.String()

            for _, want := range tt.wantContains {
                if !strings.Contains(result, want) {
                    t.Errorf("output should contain %q\nGot: %s", want, result)
                }
            }

            for _, notWant := range tt.wantNotContains {
                if strings.Contains(result, notWant) {
                    t.Errorf("output should not contain %q\nGot: %s", notWant, result)
                }
            }
        })
    }
}
```

## 実行手順

### testdata の生成

```bash
# testdata ディレクトリを作成
mkdir -p testdata/golden

# テストデータを生成
claude -p --verbose --output-format=stream-json "Say hello" > testdata/simple.json
claude -p --verbose --output-format=stream-json "List all Go files" > testdata/with_tools.json

# permission_denied.json はすでに /tmp に保存されている
cp /tmp/claude_output_sample.json testdata/permission_denied.json
```

### ゴールデンファイルの生成

```bash
# 初回はゴールデンファイルを生成
go test -run TestEndToEnd_Golden -update
```

### テスト実行

```bash
# すべてのテストを実行
go test -v

# エンドツーエンドテストのみ
go test -v -run TestEndToEnd
```

## 完了条件

- [ ] testdata/ ディレクトリが作成されている
- [ ] simple.json が準備されている
- [ ] with_tools.json が準備されている
- [ ] permission_denied.json が準備されている
- [ ] TestEndToEnd_Golden が実装され、パスする
- [ ] TestEndToEnd_Scenarios が実装され、パスする
- [ ] ゴールデンファイルが生成されている
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add testdata and end-to-end tests

実際のClaude出力を使ったエンドツーエンドテストを追加:
- testdata/: 実際のClaude出力サンプル
  - simple.json: シンプルなテキスト応答
  - with_tools.json: ツール使用を含む
  - permission_denied.json: 権限エラーを含む
- ゴールデンファイルテスト:
  - 各設定での期待出力をゴールデンファイルと比較
  - --update フラグでゴールデンファイルを更新
- シナリオテスト:
  - デフォルト設定、ツールのみ、最小限表示などの実使用シナリオ

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 10: ドキュメント整備と最終確認

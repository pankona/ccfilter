# ccfilter テスト設計

## テスト戦略

### 目標

1. **正確性**: 各メッセージタイプを正しくパースし、フォーマットできる
2. **堅牢性**: 不正なJSON、未知のフィールド、エッジケースに対応できる
3. **保守性**: テストコードが読みやすく、変更に強い
4. **信頼性**: テストが決定的で、偶発的な失敗がない

### テストレベル

1. **ユニットテスト**: 各関数の個別動作を検証
2. **インテグレーションテスト**: エンドツーエンドのフィルタリング動作を検証

## テストデータ

### 1. 実際のClaudeデータ (testdata/)

実際のClaude CLI出力を保存:
- `testdata/simple.json`: シンプルな実行 (テキスト応答のみ)
- `testdata/with_tools.json`: ツール使用を含む実行
- `testdata/permission_denied.json`: パーミッション拒否を含む実行
- `testdata/multi_turn.json`: 複数ターンの会話
- `testdata/error.json`: エラーを含む実行

### 2. エッジケーステストデータ

- `testdata/malformed.json`: 不正なJSON行を含む
- `testdata/empty.json`: 空の入力
- `testdata/large_output.json`: 非常に長いツール出力
- `testdata/unicode.json`: 多言語テキストを含む

### テストデータ構造例

```json
// testdata/simple.json
{"type":"system","subtype":"init","cwd":"/test","session_id":"test-123","model":"claude-sonnet-4-5-20250929","claude_code_version":"2.0.21"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, World!"}],"usage":{"input_tokens":3,"output_tokens":10}}}
{"type":"result","subtype":"success","result":"Hello, World!","duration_ms":1000,"total_cost_usd":0.01,"num_turns":1}
```

## ユニットテスト設計

### types_test.go

#### JSONパーステスト

```go
func TestParseSystemMessage(t *testing.T)
func TestParseAssistantMessage_Text(t *testing.T)
func TestParseAssistantMessage_ToolUse(t *testing.T)
func TestParseUserMessage_ToolResult(t *testing.T)
func TestParseResultMessage(t *testing.T)
func TestParseMessage_InvalidJSON(t *testing.T)
func TestParseMessage_UnknownType(t *testing.T)
```

**テストケース例**:
```go
func TestParseAssistantMessage_Text(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    AssistantMessage
        wantErr bool
    }{
        {
            name: "simple text",
            input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`,
            want: AssistantMessage{
                Type: "assistant",
                Message: struct{...}{
                    Content: []Content{{Type: "text", Text: "Hello"}},
                },
            },
            wantErr: false,
        },
        {
            name: "empty content",
            input: `{"type":"assistant","message":{"content":[]}}`,
            want: AssistantMessage{...},
            wantErr: false,
        },
        // ... more cases
    }
    // テスト実行
}
```

### filter_test.go

#### メッセージフィルタリングテスト

```go
func TestShouldDisplay_System(t *testing.T)
func TestShouldDisplay_Assistant(t *testing.T)
func TestShouldDisplay_Tools(t *testing.T)
func TestShouldDisplay_Result(t *testing.T)
func TestFilterMessages(t *testing.T)
```

**テストケース例**:
```go
func TestShouldDisplay_System(t *testing.T) {
    tests := []struct {
        name   string
        msg    Message
        config FilterConfig
        want   bool
    }{
        {
            name: "default config shows no system",
            msg: Message{Type: "system"},
            config: FilterConfig{},
            want: false,
        },
        {
            name: "explicit --system shows system",
            msg: Message{Type: "system"},
            config: FilterConfig{ShowSystem: true},
            want: true,
        },
        {
            name: "--all shows system",
            msg: Message{Type: "system"},
            config: FilterConfig{ShowSystem: true, ShowAssistant: true, ShowTools: true, ShowResult: true},
            want: true,
        },
    }
    // テスト実行
}
```

### formatter_test.go

#### 出力フォーマットテスト

```go
func TestFormatSystem(t *testing.T)
func TestFormatAssistant_Text(t *testing.T)
func TestFormatToolUse(t *testing.T)
func TestFormatToolResult(t *testing.T)
func TestFormatResult(t *testing.T)
func TestTruncateOutput(t *testing.T)
func TestFormatMessage_Minimal(t *testing.T)
func TestFormatMessage_Standard(t *testing.T)
func TestFormatMessage_Verbose(t *testing.T)
```

**テストケース例**:
```go
func TestFormatToolUse(t *testing.T) {
    tests := []struct {
        name   string
        content Content
        config FilterConfig
        want   string
    }{
        {
            name: "Glob tool standard format",
            content: Content{
                Type: "tool_use",
                Name: "Glob",
                Input: json.RawMessage(`{"pattern":"**/*.go"}`),
            },
            config: FilterConfig{InfoLevel: "standard"},
            want: "→ Glob: pattern=\"**/*.go\"",
        },
        {
            name: "Bash tool with long command",
            content: Content{
                Type: "tool_use",
                Name: "Bash",
                Input: json.RawMessage(`{"command":"ls -la /very/long/path"}`),
            },
            config: FilterConfig{InfoLevel: "minimal"},
            want: "→ Bash",
        },
        {
            name: "Write tool with file_path",
            content: Content{
                Type: "tool_use",
                Name: "Write",
                Input: json.RawMessage(`{"file_path":"/tmp/test.go","content":"..."}`),
            },
            config: FilterConfig{InfoLevel: "standard"},
            want: "→ Write: file_path=\"/tmp/test.go\"",
        },
    }
    // テスト実行
}

func TestTruncateOutput(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        maxLines int
        want     string
    }{
        {
            name: "no truncation needed",
            input: "line1\nline2\nline3",
            maxLines: 5,
            want: "line1\nline2\nline3",
        },
        {
            name: "truncate with ellipsis",
            input: "line1\nline2\nline3\nline4\nline5\nline6",
            maxLines: 3,
            want: "line1\nline2\nline3\n... (3 more lines)",
        },
        {
            name: "empty input",
            input: "",
            maxLines: 5,
            want: "",
        },
    }
    // テスト実行
}
```

### color_test.go

#### カラー出力テスト

```go
func TestColorize(t *testing.T)
func TestColorize_NoColor(t *testing.T)
```

**テストケース例**:
```go
func TestColorize(t *testing.T) {
    tests := []struct {
        name  string
        text  string
        color string
        want  string
    }{
        {
            name: "green text",
            text: "success",
            color: "green",
            want: "\x1b[32msuccess\x1b[0m",
        },
        {
            name: "unknown color returns original",
            text: "test",
            color: "unknown",
            want: "test",
        },
    }
    // テスト実行
}
```

## インテグレーションテスト設計

### main_test.go

#### エンドツーエンドテスト

```go
func TestMain_DefaultMode(t *testing.T)
func TestMain_MinimalMode(t *testing.T)
func TestMain_VerboseMode(t *testing.T)
func TestMain_ToolsOnly(t *testing.T)
func TestMain_WithCost(t *testing.T)
func TestMain_MalformedInput(t *testing.T)
func TestMain_EmptyInput(t *testing.T)
```

**テストケース例**:
```go
func TestMain_DefaultMode(t *testing.T) {
    tests := []struct {
        name       string
        inputFile  string
        args       []string
        wantOutput []string // 出力に含まれるべき文字列のリスト
        wantErr    bool
    }{
        {
            name: "simple execution",
            inputFile: "testdata/simple.json",
            args: []string{},
            wantOutput: []string{
                "Hello, World!",
                "Duration:",
                "Cost:",
            },
            wantErr: false,
        },
        {
            name: "with tools",
            inputFile: "testdata/with_tools.json",
            args: []string{},
            wantOutput: []string{
                "→ Glob:",
                "← No files found",
                "→ Bash:",
                "Duration:",
            },
            wantErr: false,
        },
        {
            name: "permission denied",
            inputFile: "testdata/permission_denied.json",
            args: []string{},
            wantOutput: []string{
                "→ Write:",
                "← Error: permission denied",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テストの実行
            // 1. testdata からファイルを読み取る
            // 2. 標準入力をファイルの内容に設定
            // 3. 標準出力をキャプチャ
            // 4. main() を実行
            // 5. 出力に期待する文字列が含まれているか検証
        })
    }
}
```

### テストヘルパー関数

```go
// runFilter は標準入力/出力をリダイレクトしてフィルタを実行する
func runFilter(input io.Reader, args []string) (string, error)

// loadTestData はtestdataファイルを読み込む
func loadTestData(filename string) ([]byte, error)

// captureOutput は関数の標準出力をキャプチャする
func captureOutput(f func()) string

// assertContains は出力に期待する文字列が含まれているか検証
func assertContains(t *testing.T, output, substr string)

// assertNotContains は出力に文字列が含まれていないことを検証
func assertNotContains(t *testing.T, output, substr string)
```

## テーブルドリブンテスト

すべてのテストでテーブルドリブンテストパターンを使用:

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   XXX
        want    YYY
        wantErr bool
    }{
        // テストケース
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## ゴールデンファイルテスト

複雑な出力フォーマットには、ゴールデンファイルテストを使用:

```go
func TestFormatMessage_Golden(t *testing.T) {
    tests := []struct {
        name       string
        inputFile  string
        goldenFile string
        config     FilterConfig
    }{
        {
            name: "default format",
            inputFile: "testdata/with_tools.json",
            goldenFile: "testdata/golden/with_tools_default.txt",
            config: FilterConfig{InfoLevel: "standard"},
        },
        {
            name: "minimal format",
            inputFile: "testdata/with_tools.json",
            goldenFile: "testdata/golden/with_tools_minimal.txt",
            config: FilterConfig{InfoLevel: "minimal"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 入力を読み込み、フォーマット
            got := formatFromFile(tt.inputFile, tt.config)

            // --update フラグが指定されていればゴールデンファイルを更新
            if *update {
                os.WriteFile(tt.goldenFile, []byte(got), 0644)
                return
            }

            // ゴールデンファイルと比較
            want, _ := os.ReadFile(tt.goldenFile)
            if got != string(want) {
                t.Errorf("output differs from golden file")
                t.Errorf("got:\n%s", got)
                t.Errorf("want:\n%s", want)
            }
        })
    }
}
```

## テストカバレッジ目標

- **全体**: 80%以上
- **filter.go**: 90%以上 (コアロジック)
- **formatter.go**: 90%以上 (コアロジック)
- **types.go**: 100% (データ構造)
- **main.go**: 70%以上 (CLI部分)

## CI/CDでのテスト

### テスト実行コマンド

```bash
# すべてのテストを実行
go test ./...

# カバレッジレポート付き
go test -cover ./...

# 詳細なカバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detectorを有効にして実行
go test -race ./...

# ベンチマークテスト
go test -bench=. ./...
```

### GitHub Actions設定例

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi
```

## エッジケースとエラーハンドリング

### テストすべきエッジケース

1. **空の入力**
   - 入力が全くない場合
   - 空行のみの場合

2. **不正なJSON**
   - JSONパースエラー
   - 不完全なJSON行

3. **未知のフィールド**
   - 新しいメッセージタイプ
   - 未知のツール名

4. **極端に長い出力**
   - 数MB のツール出力
   - 非常に長い単一行

5. **特殊文字**
   - Unicode文字
   - 制御文字
   - バイナリデータ

6. **境界値**
   - 0トークン
   - 負のコスト (ありえないがテスト)
   - 空の文字列

## パフォーマンステスト

### ベンチマークテスト

```go
func BenchmarkParseMessage(b *testing.B)
func BenchmarkFormatMessage(b *testing.B)
func BenchmarkFilterMessages(b *testing.B)
func BenchmarkTruncateOutput(b *testing.B)
```

**ベンチマーク例**:
```go
func BenchmarkParseMessage(b *testing.B) {
    input := `{"type":"assistant","message":{"content":[{"type":"text","text":"test"}]}}`
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parseMessage(input)
    }
}

func BenchmarkFilterMessages_LargeInput(b *testing.B) {
    // 1000行のテストデータを準備
    data, _ := loadTestData("testdata/large_output.json")
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        r := bytes.NewReader(data)
        filterMessages(r, &FilterConfig{})
    }
}
```

## モックとスタブ

標準ライブラリのみを使用するため、モックは最小限:

```go
// io.Reader のモック (テスト用)
type mockReader struct {
    data []byte
    pos  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
    if m.pos >= len(m.data) {
        return 0, io.EOF
    }
    n = copy(p, m.data[m.pos:])
    m.pos += n
    return n, nil
}

// io.Writer のモック (出力キャプチャ用)
type mockWriter struct {
    buf bytes.Buffer
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
    return m.buf.Write(p)
}
```

## テスト実行順序

1. **ユニットテスト**: 最初に実行、高速
2. **インテグレーションテスト**: ユニットテスト後、やや時間がかかる
3. **ベンチマークテスト**: オプション、必要に応じて実行

## テストデータメンテナンス

### testdata/ ディレクトリ構造

```
testdata/
├── simple.json              # シンプルな実行
├── with_tools.json          # ツール使用
├── permission_denied.json   # パーミッション拒否
├── multi_turn.json          # 複数ターン
├── error.json               # エラー
├── malformed.json           # 不正なJSON
├── empty.json               # 空の入力
├── large_output.json        # 大きな出力
├── unicode.json             # Unicode文字
└── golden/
    ├── with_tools_default.txt
    ├── with_tools_minimal.txt
    └── with_tools_verbose.txt
```

### テストデータ生成

実際のClaude実行からテストデータを生成:

```bash
# テストデータ生成スクリプト
claude -p --verbose --output-format=stream-json "List Go files" > testdata/with_tools.json
claude -p --verbose --output-format=stream-json "Hello" > testdata/simple.json
```

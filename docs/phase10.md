# Phase 10: ドキュメント整備と最終確認

## 目標

README の作成、ドキュメントの整備、最終的な動作確認を行い、プロジェクトを完成させる。

## 前提条件

- Phase 09 が完了していること
- すべての機能が実装され、テストがパスしていること

## 実装内容

### 1. README.md の作成

プロジェクトのメインドキュメント。

#### 内容

```markdown
# ccfilter

`claude -p --verbose --output-format=stream-json` の出力を人間が読みやすい形にフィルタリングするCLIツール。

## 背景

`claude -p` で実行すると、処理が完了するまで Claude Code の内部動作 (どんな考えでどういう操作をしたのか、どのツールを使ったのか) がわからない。待っている間は進行状況が見えないため、正しい方向に進んでいるのかを確認できない。

このツールは、`stream-json` 出力をリアルタイムでフィルタリングし、通常のインタラクティブモードと同様の進行状況を標準出力に表示する。

## インストール

```bash
go install github.com/pankona/ccfilter@latest
```

または、リポジトリをクローンしてビルド:

```bash
git clone https://github.com/pankona/ccfilter.git
cd ccfilter
go build
```

## 使い方

### 基本的な使用法

```bash
claude -p --verbose --output-format=stream-json "prompt" | ccfilter
```

### オプション

#### メッセージタイプフィルタ

- `--system`: system メッセージを表示
- `--assistant`: assistant メッセージのみ表示
- `--tools`: ツールメッセージのみ表示
- `--result`: result メッセージのみ表示
- `--all`: すべてのメッセージを表示

#### 情報レベル

- `--minimal`: 最小限の情報のみ表示
- `--verbose` / `-v`: 詳細情報を表示
- (デフォルトは standard レベル)

#### 追加情報

- `--show-cost`: コスト情報を常に表示
- `--show-usage`: トークン使用量を常に表示
- `--show-timing`: 実行時間情報を常に表示

#### 出力設定

- `--format=FORMAT`: 出力形式 (text|json|compact) [デフォルト: text]
- `--color`: カラー出力を強制有効化
- `--no-color`: カラー出力を無効化

### 使用例

#### デフォルト: インタラクティブモード相当の表示

```bash
claude -p --verbose --output-format=stream-json "List Go files" | ccfilter
```

出力例:
```
`**/*.go` パターンでGoファイルを検索します。

→ Glob: pattern="**/*.go"
← No files found

現在のディレクトリには `.go` ファイルは存在していません。

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
現在のディレクトリには `.go` ファイルは存在していません。

Duration: 5.0s | Cost: $0.0123 | Turns: 3
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### ツールの使用状況のみを追跡

```bash
claude -p --verbose --output-format=stream-json "Implement feature" | ccfilter --tools
```

#### 最小限の情報のみ表示

```bash
claude -p --verbose --output-format=stream-json "hello" | ccfilter --minimal
```

#### 詳細情報とコスト・実行時間を表示

```bash
claude -p --verbose --output-format=stream-json "hello" | ccfilter --verbose --show-cost --show-timing
```

## 開発

### テストの実行

```bash
# すべてのテストを実行
go test -v

# カバレッジ付き
go test -cover

# ゴールデンファイルを更新
go test -run TestEndToEnd_Golden -update
```

### ビルド

```bash
go build
```

## ドキュメント

- [design.md](docs/design.md): 詳細設計
- [test_design.md](docs/test_design.md): テスト設計
- [plan.md](docs/plan.md): 実装プラン
- [phase*.md](docs/): 各フェーズの実装手順

## ライセンス

MIT License

## 作者

pankona
```

### 2. go.mod の整備

モジュール情報を確認。

```bash
go mod tidy
```

### 3. 最終動作確認

#### 実際のClaude出力で動作確認

```bash
# シンプルなプロンプト
claude -p --verbose --output-format=stream-json "Say hello" | ./ccfilter

# ツールを使うプロンプト
claude -p --verbose --output-format=stream-json "List all Go files" | ./ccfilter

# 各オプションの動作確認
claude -p --verbose --output-format=stream-json "hello" | ./ccfilter --minimal
claude -p --verbose --output-format=stream-json "hello" | ./ccfilter --verbose
claude -p --verbose --output-format=stream-json "hello" | ./ccfilter --tools
claude -p --verbose --output-format=stream-json "hello" | ./ccfilter --no-color
```

### 4. カバレッジ確認

```bash
# カバレッジレポートを生成
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# カバレッジ率を確認 (目標: 80%以上)
go tool cover -func=coverage.out | grep total
```

### 5. .gitignore の整備

#### .gitignore

```
# Binaries
ccfilter
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Coverage
coverage.out
coverage.html

# Temporary files
/tmp/
*.tmp

# IDE
.vscode/
.idea/
*.swp
*.swo
```

### 6. GitHub Actions CI (オプション)

#### .github/workflows/test.yml

```yaml
name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi

      - name: Build
        run: go build -v
```

## チェックリスト

### ドキュメント

- [ ] README.md が作成されている
- [ ] 使い方が明確に説明されている
- [ ] 使用例が複数用意されている
- [ ] 開発者向けドキュメントへのリンクがある

### コード品質

- [ ] go mod tidy が実行されている
- [ ] すべてのテストがパスする
- [ ] カバレッジが 80% 以上
- [ ] go vet でエラーがない
- [ ] gofmt で整形されている

### 動作確認

- [ ] 実際のClaude出力で動作確認済み
- [ ] デフォルト設定での動作確認
- [ ] --minimal オプションの動作確認
- [ ] --verbose オプションの動作確認
- [ ] --tools オプションの動作確認
- [ ] --no-color オプションの動作確認
- [ ] --help の表示確認

### Git

- [ ] .gitignore が整備されている
- [ ] 不要なファイルがコミットされていない
- [ ] コミットメッセージが適切

## 最終確認コマンド

```bash
# コード整形
gofmt -w .

# vet チェック
go vet ./...

# mod tidy
go mod tidy

# すべてのテスト
go test -v

# カバレッジ
go test -cover

# ビルド
go build

# 実行確認
echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}' | ./ccfilter
```

## コミットメッセージ

```
Add README and finalize project

プロジェクトを完成:
- README.md: 使い方、オプション、使用例を詳細に説明
- .gitignore: 不要なファイルを除外
- go mod tidy: 依存関係を整理
- 最終動作確認: 実際のClaude出力で動作確認済み
- ドキュメント整備: 開発ドキュメントへのリンク

カバレッジ: XX% (80%以上達成)
すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 完了

これでプロジェクトは完成です！

次のステップ:
- GitHub にプッシュ
- リリースタグの作成
- 必要に応じて GitHub Actions CI の設定

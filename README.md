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

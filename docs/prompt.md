# ccfilter

- `claude -p --verbose --output-format=stream-json {prompt}` したときに出てくる JSON output を、人間に読みやすい形にフィルターするツール
- `claude -p --verbose --output-format=stream-json {prompt} | ccfilter {コマンドライン引数}` みたいな感じで使う想定
  - コマンドライン引数で、どういう種類のメッセージを表示するかを指定できる

## 設計

- Go 言語で実装する
- package は main package のみで完結させる

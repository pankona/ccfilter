# Phase 03: フィルタリングロジック

## 目標

メッセージのフィルタリングロジックを TDD で実装する。

## 前提条件

- Phase 02 が完了していること
- すべてのメッセージタイプがパースできること

## 実装内容

### 1. FilterConfig 構造体

フィルタリングの設定を保持する構造体。

#### テストケース (filter_test.go)

```go
func TestNewFilterConfig(t *testing.T)
```

- デフォルト設定の検証
- 各オプションの設定

#### 実装 (filter.go)

```go
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

// NewFilterConfig はデフォルト設定でFilterConfigを作成
func NewFilterConfig() *FilterConfig {
    return &FilterConfig{
        ShowAssistant: true,
        ShowTools:     true,
        ShowResult:    true,
        InfoLevel:     "standard",
        Format:        "text",
        UseColor:      true,
    }
}
```

### 2. メッセージタイプの判定

JSON から Message を読み取り、タイプを判定する。

#### テストケース (filter_test.go)

```go
func TestParseMessageType(t *testing.T)
```

- 各メッセージタイプの判定
- 不正な JSON のハンドリング

#### 実装 (filter.go)

```go
// parseMessageType はJSON行からメッセージタイプを判定
func parseMessageType(line string) (string, error) {
    var msg Message
    if err := json.Unmarshal([]byte(line), &msg); err != nil {
        return "", err
    }
    return msg.Type, nil
}
```

### 3. フィルタリング判定

メッセージを表示すべきかどうかを判定する。

#### テストケース (filter_test.go)

```go
func TestShouldDisplay(t *testing.T)
```

- system メッセージのフィルタリング
- assistant メッセージのフィルタリング
- tool メッセージのフィルタリング
- result メッセージのフィルタリング
- デフォルト設定での動作

#### 実装 (filter.go)

```go
// shouldDisplay はメッセージを表示すべきかどうかを判定
func shouldDisplay(msgType string, config *FilterConfig) bool {
    switch msgType {
    case "system":
        return config.ShowSystem
    case "assistant":
        return config.ShowAssistant
    case "user":
        return config.ShowTools
    case "result":
        return config.ShowResult
    default:
        return false
    }
}
```

### 4. Content のフィルタリング

AssistantMessage の Content が tool_use の場合、ShowTools の設定を確認。

#### テストケース (filter_test.go)

```go
func TestShouldDisplayContent(t *testing.T)
```

- text コンテンツは ShowAssistant に従う
- tool_use コンテンツは ShowTools に従う

#### 実装 (filter.go)

```go
// shouldDisplayContent は Content を表示すべきかどうかを判定
func shouldDisplayContent(content Content, config *FilterConfig) bool {
    switch content.Type {
    case "text":
        return config.ShowAssistant
    case "tool_use":
        return config.ShowTools
    default:
        return false
    }
}
```

## TDD サイクル

### FilterConfig

1. **Red**: TestNewFilterConfig を書く
2. **Green**: NewFilterConfig を実装
3. **Test**: テストを実行

### メッセージタイプ判定

1. **Red**: TestParseMessageType を書く
2. **Green**: parseMessageType を実装
3. **Test**: テストを実行

### フィルタリング判定

1. **Red**: TestShouldDisplay を書く
2. **Green**: shouldDisplay を実装
3. **Test**: テストを実行

### Content フィルタリング

1. **Red**: TestShouldDisplayContent を書く
2. **Green**: shouldDisplayContent を実装
3. **Test**: テストを実行

## テストコード例

### TestNewFilterConfig

```go
func TestNewFilterConfig(t *testing.T) {
    config := NewFilterConfig()

    if config.ShowSystem {
        t.Error("ShowSystem should be false by default")
    }
    if !config.ShowAssistant {
        t.Error("ShowAssistant should be true by default")
    }
    if !config.ShowTools {
        t.Error("ShowTools should be true by default")
    }
    if !config.ShowResult {
        t.Error("ShowResult should be true by default")
    }
    if config.InfoLevel != "standard" {
        t.Errorf("InfoLevel = %v, want standard", config.InfoLevel)
    }
}
```

### TestShouldDisplay

```go
func TestShouldDisplay(t *testing.T) {
    tests := []struct {
        name    string
        msgType string
        config  FilterConfig
        want    bool
    }{
        {
            name:    "system with ShowSystem=false",
            msgType: "system",
            config:  FilterConfig{ShowSystem: false},
            want:    false,
        },
        {
            name:    "system with ShowSystem=true",
            msgType: "system",
            config:  FilterConfig{ShowSystem: true},
            want:    true,
        },
        {
            name:    "assistant with default config",
            msgType: "assistant",
            config:  *NewFilterConfig(),
            want:    true,
        },
        {
            name:    "user (tool result) with ShowTools=true",
            msgType: "user",
            config:  FilterConfig{ShowTools: true},
            want:    true,
        },
        {
            name:    "result with ShowResult=true",
            msgType: "result",
            config:  FilterConfig{ShowResult: true},
            want:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := shouldDisplay(tt.msgType, &tt.config)
            if got != tt.want {
                t.Errorf("shouldDisplay() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 完了条件

- [ ] filter.go が作成されている
- [ ] filter_test.go が作成されている
- [ ] TestNewFilterConfig が実装され、パスする
- [ ] TestParseMessageType が実装され、パスする
- [ ] TestShouldDisplay が実装され、パスする
- [ ] TestShouldDisplayContent が実装され、パスする
- [ ] すべてのテスト (go test -v) がパスする
- [ ] Git コミットが作成されている

## コミットメッセージ

```
Add message filtering logic

TDD approach でメッセージフィルタリング機能を実装:
- FilterConfig: フィルタ設定を保持
  - デフォルト設定: assistant, tools, result を表示
- parseMessageType: JSON からメッセージタイプを判定
- shouldDisplay: メッセージを表示すべきかを判定
- shouldDisplayContent: Content を表示すべきかを判定

すべてのテストがパスすることを確認済み

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 次のフェーズ

Phase 04: 基本フォーマッター (text コンテンツ)

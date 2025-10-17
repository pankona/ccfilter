package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	config, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

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

// parseArgs はコマンドライン引数をパース
func parseArgs() (*FilterConfig, error) {
	config := NewFilterConfig()

	// フラグ定義
	var (
		showSystem    = flag.Bool("system", false, "Show system messages")
		showAssistant = flag.Bool("assistant", false, "Show only assistant messages")
		showTools     = flag.Bool("tools", false, "Show only tool messages")
		showResult    = flag.Bool("result", false, "Show only result messages")
		showAll       = flag.Bool("all", false, "Show all messages")

		minimal  = flag.Bool("minimal", false, "Show minimal information")
		verbose  = flag.Bool("verbose", false, "Show verbose information")
		verboseV = flag.Bool("v", false, "Show verbose information (short)")

		showCost   = flag.Bool("show-cost", false, "Always show cost information")
		showUsage  = flag.Bool("show-usage", false, "Always show token usage")
		showTiming = flag.Bool("show-timing", false, "Always show timing information")

		noColor = flag.Bool("no-color", false, "Disable color output")
		color   = flag.Bool("color", false, "Force enable color output")

		format = flag.String("format", "text", "Output format (text|json|compact)")

		help = flag.Bool("help", false, "Show help message")
		h    = flag.Bool("h", false, "Show help message (short)")
	)

	flag.Parse()

	// ヘルプ表示
	if *help || *h {
		printHelp()
		os.Exit(0)
	}

	// メッセージタイプフィルタ
	if *showAll {
		config.ShowSystem = true
		config.ShowAssistant = true
		config.ShowTools = true
		config.ShowResult = true
	} else {
		if *showSystem {
			config.ShowSystem = true
		}
		if *showAssistant {
			// assistant のみ表示
			config.ShowAssistant = true
			config.ShowTools = false
			config.ShowResult = false
		}
		if *showTools {
			// tools のみ表示
			config.ShowAssistant = false
			config.ShowTools = true
			config.ShowResult = false
		}
		if *showResult {
			// result のみ表示
			config.ShowAssistant = false
			config.ShowTools = false
			config.ShowResult = true
		}
	}

	// 情報レベル
	if *minimal {
		config.InfoLevel = "minimal"
	} else if *verbose || *verboseV {
		config.InfoLevel = "verbose"
	}

	// 個別表示オプション
	if *showCost {
		config.ShowCost = true
	}
	if *showUsage {
		config.ShowUsage = true
	}
	if *showTiming {
		config.ShowTiming = true
	}

	// カラー設定
	if *noColor {
		config.UseColor = false
	} else if *color {
		config.UseColor = true
	}

	// フォーマット
	config.Format = *format
	if config.Format != "text" && config.Format != "json" && config.Format != "compact" {
		return nil, fmt.Errorf("invalid format: %s (must be text, json, or compact)", config.Format)
	}

	return config, nil
}

// printHelp はヘルプメッセージを表示
func printHelp() {
	fmt.Fprintf(os.Stderr, `Usage: ccfilter [options]

ccfilter filters Claude CLI stream-json output for human readability.

Usage:
  claude -p --verbose --output-format=stream-json <prompt> | ccfilter [options]

Message Type Filters:
  --system          Show system messages
  --assistant       Show only assistant messages
  --tools           Show only tool messages (tool_use and tool_result)
  --result          Show only result messages
  --all             Show all messages

Information Level:
  --minimal, -m     Show minimal information
  --verbose, -v     Show verbose information
  (default is standard level)

Additional Information:
  --show-cost       Always show cost information
  --show-usage      Always show token usage
  --show-timing     Always show timing information

Output Format:
  --format=FORMAT   Output format (text|json|compact) [default: text]
  --color           Force enable color output
  --no-color        Disable color output

Other:
  --help, -h        Show this help message

Examples:
  # Default: show assistant messages, tools, and results
  claude -p --verbose --output-format=stream-json "hello" | ccfilter

  # Show only tool usage
  claude -p --verbose --output-format=stream-json "list files" | ccfilter --tools

  # Minimal output
  claude -p --verbose --output-format=stream-json "hello" | ccfilter --minimal

  # Verbose with cost and timing
  claude -p --verbose --output-format=stream-json "hello" | ccfilter --verbose --show-cost --show-timing
`)
}

package main

// ANSIカラーコード
const (
	ColorReset  = "\x1b[0m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorBlue   = "\x1b[34m"
	ColorRed    = "\x1b[31m"
	ColorCyan   = "\x1b[36m"
	ColorGray   = "\x1b[90m"
)

// colorize はテキストを指定色で装飾
func colorize(text, color string, enabled bool) string {
	if !enabled {
		return text
	}

	switch color {
	case "green":
		return ColorGreen + text + ColorReset
	case "yellow":
		return ColorYellow + text + ColorReset
	case "blue":
		return ColorBlue + text + ColorReset
	case "red":
		return ColorRed + text + ColorReset
	case "cyan":
		return ColorCyan + text + ColorReset
	case "gray":
		return ColorGray + text + ColorReset
	default:
		return text
	}
}

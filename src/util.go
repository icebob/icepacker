package icepacker

import (
	"fmt"
	"runtime"
	"strings"
)

// ClearLine creates a platform dependent string to clear the current
// line, so it can be overwritten. ANSI sequences are not supported on
// current windows cmd shell.
func ClearLine() string {
	if runtime.GOOS == "windows" {
		return strings.Repeat(" ", 20) + "\r"
	}
	//return "\x1b[2K"
	return "\r"
}

// PrintProgress print the percentage of the progress to the STDOUT
func PrintProgress(message string, pos int, count int) {
	fmt.Printf("%s%s: %s", ClearLine(), message, FormatPercent(uint64(pos), uint64(count)))
}

// FormatBytes humanize the length of the content
func FormatBytes(c int64) string {
	b := float64(c)

	switch {
	case c >= 1<<40:
		return fmt.Sprintf("%.3f TiB", b/(1<<40))
	case c >= 1<<30:
		return fmt.Sprintf("%.3f GiB", b/(1<<30))
	case c >= 1<<20:
		return fmt.Sprintf("%.3f MiB", b/(1<<20))
	case c >= 1<<10:
		return fmt.Sprintf("%.3f KiB", b/(1<<10))
	default:
		return fmt.Sprintf("%d B", c)
	}
}

// FormatSeconds humanize the elapsed seconds
func FormatSeconds(sec uint64) string {
	hours := sec / 3600
	sec -= hours * 3600
	min := sec / 60
	sec -= min * 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, min, sec)
	}

	return fmt.Sprintf("%d:%02d", min, sec)
}

// FormatPercent calculate the percent and format to string
func FormatPercent(numerator uint64, denominator uint64) string {
	if denominator == 0 {
		return ""
	}

	percent := 100.0 * float64(numerator) / float64(denominator)

	if percent > 100 {
		percent = 100
	}

	div := 5
	done := strings.Repeat("=", Round(percent/5))
	rest := strings.Repeat(" ", (100/div)-len(done))
	return fmt.Sprintf("%6.2f%% [%s%s]", percent, done, rest)
}

// Round rounds a float value to int
func Round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

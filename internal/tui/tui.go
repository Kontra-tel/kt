package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Blue   = "\033[34m"
)

func Header(s string) { fmt.Printf("%s%s▸ %s%s\n", Bold, Cyan, s, Reset) }
func OK(s string)     { fmt.Printf("%s✓%s %s\n", Green, Reset, s) }
func Warn(s string)   { fmt.Printf("%s!%s %s\n", Yellow, Reset, s) }
func Err(s string)    { fmt.Printf("%s✗%s %s\n", Red, Reset, s) }
func Info(s string)   { fmt.Printf("%s•%s %s\n", Blue, Reset, s) }

// Select prints a numbered list and returns the 0-based index of the chosen item.
func Select(label string, options []string) int {
	Header(label)
	for i, o := range options {
		fmt.Printf("  %s%d.%s %s\n", Bold, i+1, Reset, o)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s→%s ", Cyan, Reset)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("%s✗%s read error: %v\n", Red, Reset, err)
			}
			return 0
		}
		var n int
		if _, err := fmt.Sscan(strings.TrimSpace(scanner.Text()), &n); err == nil && n >= 1 && n <= len(options) {
			return n - 1
		}
		fmt.Printf("%s✗%s enter a number between 1 and %d\n", Red, Reset, len(options))
	}
}

// Input prints a prompt and returns the entered string, falling back to def on empty input.
func Input(prompt, def string) string {
	if def != "" {
		fmt.Printf("  %s%s%s [%s%s%s]: ", Bold, prompt, Reset, Dim, def, Reset)
	} else {
		fmt.Printf("  %s%s%s: ", Bold, prompt, Reset)
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		if s := strings.TrimSpace(scanner.Text()); s != "" {
			return s
		}
	}
	return def
}

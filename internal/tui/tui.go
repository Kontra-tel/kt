package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	okStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	errStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

func Header(s string) { fmt.Println(headerStyle.Render("▸ " + s)) }
func OK(s string)     { status(os.Stdout, okStyle, "✓", s) }
func Warn(s string)   { status(os.Stdout, warnStyle, "!", s) }
func Err(s string)    { status(os.Stderr, errStyle, "✗", s) }
func Info(s string)   { status(os.Stdout, infoStyle, "•", s) }

func status(w io.Writer, style lipgloss.Style, marker, s string) {
	fmt.Fprintf(w, "%s %s\n", style.Render(marker), s)
}

func Muted(s string) string {
	return dimStyle.Render(s)
}

// Table prints aligned human-readable rows.
func Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, headerStyle.Render(strings.ToUpper(h)))
	}
	fmt.Fprintln(tw)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}
	_ = tw.Flush()
}

// Select prompts for a choice and returns the 0-based index of the chosen item.
func Select(label string, options []string) int {
	if interactiveTerminal() {
		var choice int
		huhOptions := make([]huh.Option[int], len(options))
		for i, o := range options {
			huhOptions[i] = huh.NewOption(stripANSI(o), i)
		}
		if err := huh.NewSelect[int]().
			Title(label).
			Options(huhOptions...).
			Value(&choice).
			WithTheme(huh.ThemeCharm()).
			Run(); err == nil {
			return choice
		}
	}

	Header(label)
	for i, o := range options {
		fmt.Printf("  %s %s\n", headerStyle.Render(fmt.Sprintf("%d.", i+1)), o)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s ", headerStyle.Render("→"))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				Err(fmt.Sprintf("read error: %v", err))
			}
			return 0
		}
		var n int
		if _, err := fmt.Sscan(strings.TrimSpace(scanner.Text()), &n); err == nil && n >= 1 && n <= len(options) {
			return n - 1
		}
		Err(fmt.Sprintf("enter a number between 1 and %d", len(options)))
	}
}

// Input prompts for text and returns def on empty input.
func Input(prompt, def string) string {
	if interactiveTerminal() {
		value := def
		field := huh.NewInput().
			Title(prompt).
			Value(&value)
		if def != "" {
			field.Placeholder(def)
		}
		if err := field.WithTheme(huh.ThemeCharm()).Run(); err == nil {
			if s := strings.TrimSpace(value); s != "" {
				return s
			}
			return def
		}
	}

	if def != "" {
		fmt.Printf("  %s [%s]: ", headerStyle.Render(prompt), dimStyle.Render(def))
	} else {
		fmt.Printf("  %s: ", headerStyle.Render(prompt))
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		if s := strings.TrimSpace(scanner.Text()); s != "" {
			return s
		}
	}
	return def
}

func interactiveTerminal() bool {
	stdin, err := os.Stdin.Stat()
	if err != nil || stdin.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	stdout, err := os.Stdout.Stat()
	return err == nil && stdout.Mode()&os.ModeCharDevice != 0
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\x1b' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		if s[i] == '[' {
			for i < len(s) && (s[i] < '@' || s[i] > '~' || s[i] == '[') {
				i++
			}
			continue
		}
		for i < len(s) && (s[i] < '@' || s[i] > '~') {
			i++
		}
	}
	return b.String()
}

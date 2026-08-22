package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

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
	quiet       bool
	plain       bool
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	okStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	errStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

func SetQuiet(v bool) { quiet = v }

func SetColor(enabled bool) {
	plain = !enabled
	if enabled {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		okStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
		warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
		errStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
		infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
		dimStyle = lipgloss.NewStyle().Faint(true)
		return
	}
	titleStyle = lipgloss.NewStyle().Bold(true)
	headerStyle = lipgloss.NewStyle().Bold(true)
	okStyle = lipgloss.NewStyle().Bold(true)
	warnStyle = lipgloss.NewStyle().Bold(true)
	errStyle = lipgloss.NewStyle().Bold(true)
	infoStyle = lipgloss.NewStyle()
	dimStyle = lipgloss.NewStyle()
}

func Title(name, subtitle string) {
	if quiet {
		return
	}
	fmt.Println(titleStyle.Render(name))
	if subtitle != "" {
		fmt.Println(dimStyle.Render(subtitle))
	}
}

func Header(s string) {
	if quiet {
		return
	}
	fmt.Println()
	fmt.Println(headerStyle.Render(s))
	fmt.Println(dimStyle.Render(strings.Repeat("─", lipgloss.Width(s))))
}

func OK(s string)   { status(os.Stdout, okStyle, "ok", s, true) }
func Warn(s string) { status(os.Stdout, warnStyle, "warn", s, true) }
func Err(s string)  { status(os.Stderr, errStyle, "error", s, false) }
func Info(s string) { status(os.Stdout, infoStyle, "next", s, true) }

func status(w io.Writer, style lipgloss.Style, label, s string, suppress bool) {
	if quiet && suppress {
		return
	}
	fmt.Fprintf(w, "%s %s\n", style.Render(label+":"), s)
}

func Muted(s string) string {
	return dimStyle.Render(s)
}

// Table prints aligned human-readable rows without letting ANSI escape codes
// distort column widths.
func Table(headers []string, rows [][]string) {
	if quiet {
		return
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = lipgloss.Width(stripANSI(strings.ToUpper(h)))
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && lipgloss.Width(stripANSI(cell)) > widths[i] {
				widths[i] = lipgloss.Width(stripANSI(cell))
			}
		}
	}
	printRow := func(cells []string, style lipgloss.Style) {
		fmt.Print("  ")
		for i := range headers {
			if i > 0 {
				fmt.Print("  ")
			}
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			fmt.Print(style.Render(padRight(cell, widths[i])))
		}
		fmt.Println()
	}
	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}
	printRow(upper, headerStyle)
	for _, row := range rows {
		printRow(row, lipgloss.NewStyle())
	}
}

func padRight(s string, width int) string {
	padding := width - lipgloss.Width(stripANSI(s))
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
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
		value := ""
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
	if plain {
		return false
	}
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

package tui

import "fmt"

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

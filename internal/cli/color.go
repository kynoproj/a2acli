package cli

import (
	"io"
	"os"

	"golang.org/x/term"
)

// ANSI SGR escape sequences.
const (
	ansiReset = "\x1b[0m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
	ansiCyan  = "\x1b[36m"
)

// colorMode controls whether ANSI colors are emitted.
type colorMode int

const (
	colorAuto colorMode = iota // emit colors only when the writer is a TTY and NO_COLOR is unset
	colorOff                   // never emit colors
)

// colorEnabled reports whether ANSI colors should be applied when writing to w.
// Honors the NO_COLOR convention (https://no-color.org) and requires the writer
// to be an *os.File pointing at a terminal.
func colorEnabled(w io.Writer, mode colorMode) bool {
	if mode == colorOff {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// colorize wraps s in the given ANSI code when enabled is true; otherwise it
// returns s unchanged.
func colorize(enabled bool, code, s string) string {
	if !enabled || code == "" {
		return s
	}
	return code + s + ansiReset
}

// ErrorLabel returns the "Error:" prefix used at top-level, colorized in red
// when w is a TTY and NO_COLOR is unset.
func ErrorLabel(w io.Writer) string {
	return colorize(colorEnabled(w, colorAuto), ansiRed, "Error:")
}

// Package cli holds the conventions shared by every command in this
// repository: exit statuses, the common Run signature, and the placeholder
// used by unimplemented tools. It must never contain command logic.
package cli

import (
	"fmt"
	"io"
)

// Process exit statuses used by every command.
const (
	// ExitOK means the command completed successfully.
	ExitOK = 0
	// ExitFailure means a runtime error occurred; output may be partial.
	ExitFailure = 1
	// ExitUsage means the arguments could not be parsed; nothing was executed.
	ExitUsage = 2
)

// RunFunc is the entry point implemented by every tool package. It receives
// the arguments after the program name, reads from stdin, writes results to
// stdout/stderr, and returns the process exit status.
type RunFunc func(args []string, stdin io.Reader, stdout, stderr io.Writer) int

// NotImplemented is the placeholder used by the Run stubs. It reports a
// failure so the acceptance tests stay red until the tool is implemented.
func NotImplemented(prog string, stderr io.Writer) int {
	_, _ = fmt.Fprintf(stderr, "%s: not implemented\n", prog)
	return ExitFailure
}

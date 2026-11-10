// Package echo implements the echo utility described in docs/specs/echo.md.
package echo

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes echo with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/echo.md; the acceptance tests
// live in echo_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("echo", stderr)
}

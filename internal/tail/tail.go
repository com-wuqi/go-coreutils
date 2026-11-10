// Package tail implements the tail utility described in docs/specs/tail.md.
package tail

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes tail with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/tail.md; the acceptance tests
// live in tail_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("tail", stderr)
}

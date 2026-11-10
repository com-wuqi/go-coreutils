// Package head implements the head utility described in docs/specs/head.md.
package head

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes head with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/head.md; the acceptance tests
// live in head_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("head", stderr)
}

// Package tee implements the tee utility described in docs/specs/tee.md.
package tee

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes tee with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/tee.md; the acceptance tests
// live in tee_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("tee", stderr)
}

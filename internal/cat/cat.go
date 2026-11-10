// Package cat implements the cat utility described in docs/specs/cat.md.
package cat

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes cat with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/cat.md; the acceptance tests
// live in cat_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("cat", stderr)
}

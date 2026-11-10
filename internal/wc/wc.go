// Package wc implements the wc utility described in docs/specs/wc.md.
package wc

import (
	"io"

	"go-coreutils/internal/cli"
)

// Run executes wc with args (the program name is not included), reading
// stdin and writing results to stdout/stderr. It returns the process exit
// status.
//
// TODO: implement it according to docs/specs/wc.md; the acceptance tests
// live in wc_test.go.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cli.NotImplemented("wc", stderr)
}

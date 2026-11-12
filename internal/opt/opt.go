// Package opt implements the shared short-option scanner used by the command
// implementations. It only handles command-line syntax; number validation,
// last-wins policies, error text, and file I/O belong to the individual
// tools.
//
// See docs/specs/opt.md for the contract and internal/opt/opt_test.go for the
// acceptance tests.
package opt

import "errors"

// Option declares one short option. When TakesValue is true, the option
// requires a value.
type Option struct {
	Short      byte
	TakesValue bool
}

// Occurrence is one option as it appeared on the command line.
type Occurrence struct {
	Short byte
	Value string // empty for options that take no value
}

// Result is the output of Parse.
type Result struct {
	Operands []string     // in command-line order
	Options  []Occurrence // in command-line order, for last-wins policies
}

// UsageError reports an argument syntax error. Error returns
// "<prog>: <msg>".
type UsageError struct {
	Prog string
	Msg  string
}

// Error implements the error interface.
func (e *UsageError) Error() string {
	if e.Prog == "" {
		return e.Msg
	}
	return e.Prog + ": " + e.Msg
}

// Parse scans args according to options. It handles short-option clusters,
// attached and separate option values, "--", a lone "-", and interleaved
// options and operands. It does not validate values or touch I/O.
//
// TODO: implement Parse according to docs/specs/opt.md. The stub returns an
// error so the acceptance tests stay red.
func Parse(prog string, options []Option, args []string) (Result, error) {
	return Result{}, errors.New("opt: not implemented")
}

package cat_test

import (
	"path/filepath"
	"testing"

	"go-coreutils/internal/cat"
	"go-coreutils/internal/testutil"
)

// TestRun covers stdin-based behavior: copying, numbering, squeezing, and
// the display options.
func TestRun(t *testing.T) {
	tests := []testutil.Case{
		{Name: "stdin passthrough", Stdin: "hello\n", WantStdout: "hello\n"},
		{Name: "empty stdin", Stdin: "", WantStdout: ""},
		{Name: "number all lines", Stdin: "a\n\nb", Args: []string{"-n"}, WantStdout: "     1\ta\n     2\t\n     3\tb"},
		{Name: "number non-blank lines", Stdin: "a\n\nb\n", Args: []string{"-b"}, WantStdout: "     1\ta\n\n     2\tb\n"},
		{Name: "number non-blank lines overrides -n", Stdin: "a\n\nb\n", Args: []string{"-n", "-b"}, WantStdout: "     1\ta\n\n     2\tb\n"},
		{Name: "number non-blank lines overrides -n regardless of order", Stdin: "a\n\nb\n", Args: []string{"-b", "-n"}, WantStdout: "     1\ta\n\n     2\tb\n"},
		{Name: "squeeze blank lines", Stdin: "a\n\n\n\nb\n", Args: []string{"-s"}, WantStdout: "a\n\nb\n"},
		{Name: "squeeze leading blank lines", Stdin: "\n\n\na\n", Args: []string{"-s"}, WantStdout: "\na\n"},
		{Name: "show ends", Stdin: "a\nb\n", Args: []string{"-E"}, WantStdout: "a$\nb$\n"},
		{Name: "show ends without trailing newline", Stdin: "a", Args: []string{"-E"}, WantStdout: "a"},
		{Name: "show tabs", Stdin: "a\tb\n", Args: []string{"-T"}, WantStdout: "a^Ib\n"},
		{Name: "combined -nE", Stdin: "a\nb\n", Args: []string{"-nE"}, WantStdout: "     1\ta$\n     2\tb$\n"},
		{Name: "CRLF is preserved", Stdin: "a\r\n", WantStdout: "a\r\n"},
		{Name: "numbered CRLF is preserved", Stdin: "a\r\n", Args: []string{"-n"}, WantStdout: "     1\ta\r\n"},
		{Name: "single dash reads stdin", Stdin: "x\n", Args: []string{"-"}, WantStdout: "x\n"},
		{Name: "double dash keeps reading stdin", Stdin: "a\n", Args: []string{"-n", "--"}, WantStdout: "     1\ta\n"},
	}

	testutil.RunCases(t, cat.Run, tests)
}

// TestRunFiles covers operands, multiple files, numbering continuity, and
// option/operand interleaving.
func TestRunFiles(t *testing.T) {
	one := filepath.Join("testdata", "one.txt")
	two := filepath.Join("testdata", "two.txt")
	empty := filepath.Join("testdata", "empty.txt")
	noNL := filepath.Join("testdata", "no_trailing_newline.txt")
	blanks := filepath.Join("testdata", "blanks.txt")

	tests := []testutil.Case{
		{Name: "single file", Args: []string{one}, WantStdout: "one\ntwo\n"},
		{Name: "multiple files concatenated", Args: []string{one, two}, WantStdout: "one\ntwo\nthree\nfour\n"},
		{Name: "empty file", Args: []string{empty}, WantStdout: ""},
		{Name: "no trailing newline preserved", Args: []string{noNL}, WantStdout: "first\nsecond"},
		{Name: "stdin operand among files", Stdin: "middle\n", Args: []string{one, "-", two}, WantStdout: "one\ntwo\nmiddle\nthree\nfour\n"},
		{Name: "numbering continues across files", Args: []string{"-n", one, two}, WantStdout: "     1\tone\n     2\ttwo\n     3\tthree\n     4\tfour\n"},
		{Name: "squeeze state is global across files", Args: []string{"-s", blanks, blanks}, WantStdout: "a\n\nb\na\n\nb\n"},
		{Name: "empty file between operands", Args: []string{empty, one}, WantStdout: "one\ntwo\n"},
		{Name: "options after operands", Args: []string{one, "-n"}, WantStdout: "     1\tone\n     2\ttwo\n"},
		{Name: "double dash ends options", Args: []string{"--", one}, WantStdout: "one\ntwo\n"},
	}

	testutil.RunCases(t, cat.Run, tests)
}

// TestRunErrors covers usage errors and unreadable operands.
func TestRunErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	one := filepath.Join("testdata", "one.txt")

	tests := []testutil.Case{
		{Name: "unknown option", Args: []string{"-z"}, WantExit: 2, WantStderr: "cat: unrecognized option: -z\n"},
		{Name: "missing file", Args: []string{missing}, WantExit: 1, WantStderrSub: "cat: " + missing},
		{Name: "missing file before valid file", Args: []string{missing, one}, WantExit: 1, WantStdout: "one\ntwo\n", WantStderrSub: "cat: " + missing},
		{Name: "missing file after valid file", Args: []string{one, missing}, WantExit: 1, WantStdout: "one\ntwo\n", WantStderrSub: "cat: " + missing},
	}

	testutil.RunCases(t, cat.Run, tests)
}

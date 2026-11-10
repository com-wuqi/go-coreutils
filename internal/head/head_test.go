package head_test

import (
	"path/filepath"
	"testing"

	"go-coreutils/internal/head"
	"go-coreutils/internal/testutil"
)

// TestRun covers stdin-based line and byte selection and option precedence.
func TestRun(t *testing.T) {
	const twelve = "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n"

	tests := []testutil.Case{
		{Name: "default is ten lines", Stdin: twelve, WantStdout: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"},
		{Name: "line count as separate argument", Stdin: "a\nb\nc\n", Args: []string{"-n", "2"}, WantStdout: "a\nb\n"},
		{Name: "line count attached", Stdin: "a\nb\nc\n", Args: []string{"-n1"}, WantStdout: "a\n"},
		{Name: "zero lines", Stdin: "a\n", Args: []string{"-n", "0"}, WantStdout: ""},
		{Name: "byte count", Stdin: "a\nb\n", Args: []string{"-c", "2"}, WantStdout: "a\n"},
		{Name: "byte count attached", Stdin: "a\nb\n", Args: []string{"-c2"}, WantStdout: "a\n"},
		{Name: "zero bytes", Stdin: "a\n", Args: []string{"-c", "0"}, WantStdout: ""},
		{Name: "count larger than input", Stdin: "a\n", Args: []string{"-n", "5"}, WantStdout: "a\n"},
		{Name: "no trailing newline preserved", Stdin: "a\nb", Args: []string{"-n", "2"}, WantStdout: "a\nb"},
		{Name: "line without trailing newline", Stdin: "a\nb", Args: []string{"-n", "1"}, WantStdout: "a\n"},
		{Name: "last of line and byte count wins", Stdin: "abc\ndef\n", Args: []string{"-n", "1", "-c", "3"}, WantStdout: "abc"},
		{Name: "last of byte and line count wins", Stdin: "abc\ndef\n", Args: []string{"-c", "3", "-n", "1"}, WantStdout: "abc\n"},
		{Name: "double dash ends options", Stdin: "a\n", Args: []string{"-n", "1", "--"}, WantStdout: "a\n"},
		{Name: "dash operand reads stdin", Stdin: "a\nb\n", Args: []string{"-n", "1", "-"}, WantStdout: "a\n"},
		{Name: "verbose shows stdin header", Stdin: "a\n", Args: []string{"-v"}, WantStdout: "==> standard input <==\na\n"},
	}

	testutil.RunCases(t, head.Run, tests)
}

// TestRunFiles covers headers, multiple operands, and option/operand order.
func TestRunFiles(t *testing.T) {
	a := filepath.Join("testdata", "a.txt")
	b := filepath.Join("testdata", "b.txt")
	empty := filepath.Join("testdata", "empty.txt")

	tests := []testutil.Case{
		{Name: "multiple files with headers", Args: []string{"-n", "2", a, b}, WantStdout: "==> " + a + " <==\n1\n2\n\n==> " + b + " <==\nx\ny\n"},
		{Name: "quiet suppresses headers", Args: []string{"-q", "-n", "2", a, b}, WantStdout: "1\n2\nx\ny\n"},
		{Name: "verbose adds header for one file", Args: []string{"-v", "-n", "1", a}, WantStdout: "==> " + a + " <==\n1\n"},
		{Name: "later quiet beats verbose", Args: []string{"-v", "-q", "-n", "1", a, b}, WantStdout: "1\nx\n"},
		{Name: "later verbose beats quiet", Args: []string{"-q", "-v", "-n", "1", a}, WantStdout: "==> " + a + " <==\n1\n"},
		{Name: "verbose empty file still shows header", Args: []string{"-v", empty}, WantStdout: "==> " + empty + " <==\n"},
		{Name: "stdin operand among files", Stdin: "s\n", Args: []string{"-n", "1", a, "-"}, WantStdout: "==> " + a + " <==\n1\n\n==> standard input <==\ns\n"},
		{Name: "options after operands", Args: []string{a, "-n", "1"}, WantStdout: "1\n"},
		{Name: "combined quiet and count", Args: []string{"-qn2", a, b}, WantStdout: "1\n2\nx\ny\n"},
	}

	testutil.RunCases(t, head.Run, tests)
}

// TestRunErrors covers usage errors and unreadable operands.
func TestRunErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	a := filepath.Join("testdata", "a.txt")
	b := filepath.Join("testdata", "b.txt")

	tests := []testutil.Case{
		{Name: "invalid number", Args: []string{"-n", "abc"}, WantExit: 2, WantStderr: "head: invalid number: abc\n"},
		{Name: "negative number is unsupported", Args: []string{"-n", "-1"}, WantExit: 2, WantStderr: "head: invalid number: -1\n"},
		{Name: "missing option value", Args: []string{"-n"}, WantExit: 2, WantStderr: "head: option requires an argument: -n\n"},
		{Name: "unknown option", Args: []string{"-z"}, WantExit: 2, WantStderr: "head: unrecognized option: -z\n"},
		{Name: "bad argument stops before reading files", Args: []string{a, "-n", "abc"}, WantExit: 2, WantStderr: "head: invalid number: abc\n"},
		{Name: "missing file is reported and skipped", Args: []string{missing, b, "-n", "1"}, WantExit: 1, WantStdout: "==> " + b + " <==\nx\n", WantStderrSub: "head: " + missing},
		{Name: "single missing file", Args: []string{missing}, WantExit: 1, WantStderrSub: "head: " + missing},
	}

	testutil.RunCases(t, head.Run, tests)
}

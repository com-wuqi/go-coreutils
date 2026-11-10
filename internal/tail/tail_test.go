package tail_test

import (
	"path/filepath"
	"testing"

	"go-coreutils/internal/tail"
	"go-coreutils/internal/testutil"
)

// TestRun covers stdin-based line/byte selection and the +N start form.
func TestRun(t *testing.T) {
	const twelve = "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n"

	tests := []testutil.Case{
		{Name: "default is ten lines", Stdin: twelve, WantStdout: "3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n"},
		{Name: "last N lines", Stdin: "a\nb\nc\n", Args: []string{"-n", "2"}, WantStdout: "b\nc\n"},
		{Name: "line count attached", Stdin: "a\nb\nc\n", Args: []string{"-n2"}, WantStdout: "b\nc\n"},
		{Name: "zero lines", Stdin: "a\n", Args: []string{"-n", "0"}, WantStdout: ""},
		{Name: "from line N", Stdin: "a\nb\nc\n", Args: []string{"-n", "+2"}, WantStdout: "b\nc\n"},
		{Name: "from line one", Stdin: "a\nb\nc\n", Args: []string{"-n", "+1"}, WantStdout: "a\nb\nc\n"},
		{Name: "byte count", Stdin: "a\nb\n", Args: []string{"-c", "2"}, WantStdout: "b\n"},
		{Name: "byte count without trailing newline", Stdin: "a\nb", Args: []string{"-c", "2"}, WantStdout: "\nb"},
		{Name: "count larger than input", Stdin: "a\n", Args: []string{"-n", "99"}, WantStdout: "a\n"},
		{Name: "no trailing newline preserved", Stdin: "a\nb", Args: []string{"-n", "1"}, WantStdout: "b"},
		{Name: "from line N without trailing newline", Stdin: "a\nb", Args: []string{"-n", "+2"}, WantStdout: "b"},
		{Name: "last of line and byte count wins", Stdin: "abc\ndef\n", Args: []string{"-n", "2", "-c", "3"}, WantStdout: "ef\n"},
		{Name: "last of byte and line count wins", Stdin: "abc\ndef\n", Args: []string{"-c", "3", "-n", "2"}, WantStdout: "abc\ndef\n"},
		{Name: "double dash ends options", Stdin: "a\n", Args: []string{"-n", "1", "--"}, WantStdout: "a\n"},
		{Name: "dash operand reads stdin", Stdin: "a\nb\n", Args: []string{"-n", "1", "-"}, WantStdout: "b\n"},
		{Name: "verbose shows stdin header", Stdin: "a\n", Args: []string{"-v"}, WantStdout: "==> standard input <==\na\n"},
		{Name: "CRLF line endings are preserved", Stdin: "a\r\nb\r\n", Args: []string{"-n", "1"}, WantStdout: "b\r\n"},
	}

	testutil.RunCases(t, tail.Run, tests)
}

// TestRunFiles covers headers, multiple operands, and option/operand order.
func TestRunFiles(t *testing.T) {
	a := filepath.Join("testdata", "a.txt")
	b := filepath.Join("testdata", "b.txt")
	empty := filepath.Join("testdata", "empty.txt")
	noNL := filepath.Join("testdata", "no_trailing_newline.txt")

	tests := []testutil.Case{
		{Name: "multiple files with headers", Args: []string{"-n", "1", a, b}, WantStdout: "==> " + a + " <==\n5\n\n==> " + b + " <==\ny\n"},
		{Name: "quiet suppresses headers", Args: []string{"-q", "-n", "1", a, b}, WantStdout: "5\ny\n"},
		{Name: "verbose adds header for one file", Args: []string{"-v", "-n", "1", a}, WantStdout: "==> " + a + " <==\n5\n"},
		{Name: "later quiet beats verbose", Args: []string{"-v", "-q", "-n", "1", a, b}, WantStdout: "5\ny\n"},
		{Name: "later verbose beats quiet", Args: []string{"-q", "-v", "-n", "1", a}, WantStdout: "==> " + a + " <==\n5\n"},
		{Name: "verbose empty file still shows header", Args: []string{"-v", empty}, WantStdout: "==> " + empty + " <==\n"},
		{Name: "stdin operand among files", Stdin: "s\n", Args: []string{"-n", "1", a, "-"}, WantStdout: "==> " + a + " <==\n5\n\n==> standard input <==\ns\n"},
		{Name: "options after operands", Args: []string{a, "-n", "1"}, WantStdout: "5\n"},
		{Name: "no trailing newline file", Args: []string{noNL, "-n", "1"}, WantStdout: "b"},
	}

	testutil.RunCases(t, tail.Run, tests)
}

// TestRunErrors covers usage errors and unreadable operands.
func TestRunErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	a := filepath.Join("testdata", "a.txt")
	b := filepath.Join("testdata", "b.txt")

	tests := []testutil.Case{
		{Name: "invalid number", Args: []string{"-n", "abc"}, WantExit: 2, WantStderr: "tail: invalid number: abc\n"},
		{Name: "plus zero is invalid", Args: []string{"-n", "+0"}, WantExit: 2, WantStderr: "tail: invalid number: +0\n"},
		{Name: "negative number is invalid", Args: []string{"-n", "-1"}, WantExit: 2, WantStderr: "tail: invalid number: -1\n"},
		{Name: "missing option value", Args: []string{"-n"}, WantExit: 2, WantStderr: "tail: option requires an argument: -n\n"},
		{Name: "unknown option", Args: []string{"-z"}, WantExit: 2, WantStderr: "tail: unrecognized option: -z\n"},
		{Name: "bad argument stops before reading files", Args: []string{a, "-n", "abc"}, WantExit: 2, WantStderr: "tail: invalid number: abc\n"},
		{Name: "missing file is reported and skipped", Args: []string{missing, b, "-n", "1"}, WantExit: 1, WantStdout: "==> " + b + " <==\ny\n", WantStderrSub: "tail: " + missing},
		{Name: "single missing file", Args: []string{missing}, WantExit: 1, WantStderrSub: "tail: " + missing},
	}

	testutil.RunCases(t, tail.Run, tests)
}

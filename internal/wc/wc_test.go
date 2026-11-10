package wc_test

import (
	"path/filepath"
	"testing"

	"go-coreutils/internal/testutil"
	"go-coreutils/internal/wc"
)

const (
	sampleContent  = "hello world\nfoo bar baz\n"
	unicodeContent = "你好 世界\n"
)

// TestRun covers counting, column selection, ordering, and Unicode handling.
func TestRun(t *testing.T) {
	tests := []testutil.Case{
		{Name: "default counts", Stdin: sampleContent, WantStdout: "2 5 24\n"},
		{Name: "empty stdin", Stdin: "", WantStdout: "0 0 0\n"},
		{Name: "no trailing newline", Stdin: "one two\nthree", WantStdout: "1 3 13\n"},
		{Name: "lines only", Stdin: sampleContent, Args: []string{"-l"}, WantStdout: "2\n"},
		{Name: "words only", Stdin: sampleContent, Args: []string{"-w"}, WantStdout: "5\n"},
		{Name: "bytes only", Stdin: sampleContent, Args: []string{"-c"}, WantStdout: "24\n"},
		{Name: "chars only", Stdin: sampleContent, Args: []string{"-m"}, WantStdout: "24\n"},
		{Name: "longest line only", Stdin: sampleContent, Args: []string{"-L"}, WantStdout: "11\n"},
		{Name: "columns keep fixed order", Stdin: sampleContent, Args: []string{"-cw"}, WantStdout: "5 24\n"},
		{Name: "all columns keep fixed order", Stdin: sampleContent, Args: []string{"-L", "-c", "-m", "-w", "-l"}, WantStdout: "2 5 24 24 11\n"},
		{Name: "unicode counts", Stdin: unicodeContent, Args: []string{"-lwmcL"}, WantStdout: "1 2 6 14 5\n"},
		{Name: "whitespace splits words", Stdin: "  a\tb  \n", WantStdout: "1 2 8\n"},
		{Name: "CRLF bytes count literally", Stdin: "a\r\n", WantStdout: "1 1 3\n"},
		{Name: "double dash ends options", Stdin: "a\n", Args: []string{"--"}, WantStdout: "1 1 2\n"},
		{Name: "dash operand is named", Stdin: "a\nb\n", Args: []string{"-l", "-"}, WantStdout: "2 -\n"},
		{Name: "duplicate flags are harmless", Stdin: sampleContent, Args: []string{"-l", "-l"}, WantStdout: "2\n"},
	}

	testutil.RunCases(t, wc.Run, tests)
}

// TestRunFiles covers per-file output, the total line, and stdin operands.
func TestRunFiles(t *testing.T) {
	sample := filepath.Join("testdata", "sample.txt")
	unicode := filepath.Join("testdata", "unicode.txt")
	empty := filepath.Join("testdata", "empty.txt")
	noNL := filepath.Join("testdata", "no_trailing_newline.txt")

	tests := []testutil.Case{
		{Name: "single file", Args: []string{sample}, WantStdout: "2 5 24 " + sample + "\n"},
		{Name: "selected columns", Args: []string{"-l", sample}, WantStdout: "2 " + sample + "\n"},
		{Name: "two files and total", Args: []string{sample, unicode}, WantStdout: "2 5 24 " + sample + "\n1 2 14 " + unicode + "\n3 7 38 total\n"},
		{Name: "empty file", Args: []string{empty}, WantStdout: "0 0 0 " + empty + "\n"},
		{Name: "no trailing newline file", Args: []string{noNL}, WantStdout: "1 3 13 " + noNL + "\n"},
		{Name: "stdin operand among files", Stdin: "a\n", Args: []string{sample, "-"}, WantStdout: "2 5 24 " + sample + "\n1 1 2 -\n3 6 26 total\n"},
		{Name: "options after operands", Args: []string{sample, "-l"}, WantStdout: "2 " + sample + "\n"},
		{Name: "double dash ends options", Args: []string{"--", sample}, WantStdout: "2 5 24 " + sample + "\n"},
	}

	testutil.RunCases(t, wc.Run, tests)
}

// TestRunErrors covers usage errors, missing files, and totals over the
// readable operands only.
func TestRunErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	sample := filepath.Join("testdata", "sample.txt")

	tests := []testutil.Case{
		{Name: "unknown option", Args: []string{"-x"}, WantExit: 2, WantStderr: "wc: unrecognized option: -x\n"},
		{Name: "missing file before valid file", Args: []string{missing, sample}, WantExit: 1, WantStdout: "2 5 24 " + sample + "\n2 5 24 total\n", WantStderrSub: "wc: " + missing},
		{Name: "single missing file", Args: []string{missing}, WantExit: 1, WantStderrSub: "wc: " + missing},
	}

	testutil.RunCases(t, wc.Run, tests)
}

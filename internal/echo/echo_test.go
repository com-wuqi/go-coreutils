package echo_test

import (
	"testing"

	"go-coreutils/internal/echo"
	"go-coreutils/internal/testutil"
)

// TestRun covers argument joining, option parsing, and the -e escape table.
func TestRun(t *testing.T) {
	tests := []testutil.Case{
		{Name: "no args prints newline", WantStdout: "\n"},
		{Name: "single arg", Args: []string{"hello"}, WantStdout: "hello\n"},
		{Name: "multiple args joined by space", Args: []string{"a", "b", "c"}, WantStdout: "a b c\n"},
		{Name: "empty argument preserved", Args: []string{""}, WantStdout: "\n"},
		{Name: "empty argument between words", Args: []string{"a", "", "b"}, WantStdout: "a  b\n"},
		{Name: "-n omits trailing newline", Args: []string{"-n", "hello"}, WantStdout: "hello"},
		{Name: "-n only", Args: []string{"-n"}, WantStdout: ""},

		{Name: "-e interprets newline", Args: []string{"-e", `a\nb`}, WantStdout: "a\nb\n"},
		{Name: "-e interprets tab", Args: []string{"-e", `a\tb`}, WantStdout: "a\tb\n"},
		{Name: "-e interprets backslash", Args: []string{"-e", `a\\b`}, WantStdout: "a\\b\n"},
		{Name: "-e interprets alert backspace formfeed", Args: []string{"-e", `a\a\b\f`}, WantStdout: "a\a\b\f\n"},
		{Name: "-e interprets carriage return and vertical tab", Args: []string{"-e", `a\r\v`}, WantStdout: "a\r\v\n"},
		{Name: "-e interprets escape", Args: []string{"-e", `\e`}, WantStdout: "\x1b\n"},
		{Name: "-e interprets hex escape", Args: []string{"-e", `\x41`}, WantStdout: "A\n"},
		{Name: "-e interprets octal escape", Args: []string{"-e", `\0101`}, WantStdout: "A\n"},
		{Name: "-e interprets NUL", Args: []string{"-e", `\0`}, WantStdout: "\x00\n"},
		{Name: "-e leaves unknown escape unchanged", Args: []string{"-e", `a\qb`}, WantStdout: "a\\qb\n"},
		{Name: "-e leaves trailing backslash unchanged", Args: []string{"-e", `a\`}, WantStdout: "a\\\n"},
		{Name: "-e with \\c stops output", Args: []string{"-e", `a\cb`}, WantStdout: "a"},
		{Name: "\\c suppresses newline", Args: []string{"-e", `\c`}, WantStdout: ""},

		{Name: "combined -ne", Args: []string{"-ne", `a\tb`}, WantStdout: "a\tb"},
		{Name: "-E disables escapes", Args: []string{"-E", `a\nb`}, WantStdout: "a\\nb\n"},
		{Name: "-e followed by -E disables escapes", Args: []string{"-e", "-E", `a\nb`}, WantStdout: "a\\nb\n"},
		{Name: "-E followed by -e enables escapes", Args: []string{"-E", "-e", `a\nb`}, WantStdout: "a\nb\n"},

		{Name: "option parsing stops at first operand", Args: []string{"-n", "x", "-e"}, WantStdout: "x -e"},
		{Name: "unknown leading option is literal", Args: []string{"-z"}, WantStdout: "-z\n"},
		{Name: "double dash is literal", Args: []string{"--"}, WantStdout: "--\n"},
		{Name: "mixed flags are literal", Args: []string{"-nx"}, WantStdout: "-nx\n"},
	}

	testutil.RunCases(t, echo.Run, tests)
}

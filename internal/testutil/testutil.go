// Package testutil contains helpers shared by the command acceptance tests.
// It is imported only from _test.go files and must not be used by commands.
package testutil

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-coreutils/internal/cli"
)

// Result holds the captured output and exit status of one Run invocation.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run invokes fn with a fresh stdin reader and captures stdout/stderr.
func Run(t *testing.T, fn cli.RunFunc, stdin string, args ...string) Result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := fn(args, strings.NewReader(stdin), &stdout, &stderr)
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

// Case describes one table-driven acceptance test.
type Case struct {
	Name string
	// Stdin is fed to the command as standard input.
	Stdin string
	// Args are passed after the program name.
	Args []string
	// WantStdout must match stdout exactly.
	WantStdout string
	// WantStderr, when non-empty, must match stderr exactly. Use it for
	// usage errors whose text is defined by this project.
	WantStderr string
	// WantStderrSub, when non-empty, must be contained in stderr. Use it for
	// operating-system error text, which is not stable across platforms.
	WantStderrSub string
	// WantExit is the expected exit status; the zero value means success.
	WantExit int
}

// RunCases runs every case and asserts stdout, stderr, and the exit status.
// Cases without a stderr expectation must produce empty stderr.
func RunCases(t *testing.T, fn cli.RunFunc, cases []Case) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got := Run(t, fn, tc.Stdin, tc.Args...)
			AssertEqual(t, "exit code", got.ExitCode, tc.WantExit)
			AssertEqual(t, "stdout", got.Stdout, tc.WantStdout)
			switch {
			case tc.WantStderr != "":
				AssertEqual(t, "stderr", got.Stderr, tc.WantStderr)
			case tc.WantStderrSub != "":
				AssertContains(t, "stderr", got.Stderr, tc.WantStderrSub)
			default:
				AssertEqual(t, "stderr", got.Stderr, "")
			}
		})
	}
}

// AssertEqual fails when got is not equal to want.
func AssertEqual[T comparable](t *testing.T, label string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s mismatch:\n got: %#v\nwant: %#v", label, got, want)
	}
}

// AssertContains fails when got does not contain sub. It is used for error
// text that includes operating-system-specific details.
func AssertContains(t *testing.T, label, got, sub string) {
	t.Helper()
	if !strings.Contains(got, sub) {
		t.Fatalf("%s must contain %q:\n got: %q", label, sub, got)
	}
}

// WriteFile writes content to dir/name and returns the resulting path. It
// fails the test if the file cannot be written.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// WriteFiles creates a temporary directory containing files (name to
// content) and returns the directory path.
func WriteFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		WriteFile(t, dir, name, content)
	}
	return dir
}

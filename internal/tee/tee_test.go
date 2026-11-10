package tee_test

import (
	"os"
	"path/filepath"
	"testing"

	"go-coreutils/internal/tee"
	"go-coreutils/internal/testutil"
)

// TestRun covers behavior with no file operands.
func TestRun(t *testing.T) {
	tests := []testutil.Case{
		{Name: "no files passes stdin through", Stdin: "hello\n", WantStdout: "hello\n"},
		{Name: "empty stdin", Stdin: "", WantStdout: ""},
		{Name: "binary passthrough", Stdin: "\x00\xff\x01", WantStdout: "\x00\xff\x01"},
		{Name: "append flag without files", Stdin: "x", Args: []string{"-a"}, WantStdout: "x"},
		{Name: "double dash with no files", Stdin: "x", Args: []string{"--"}, WantStdout: "x"},
	}

	testutil.RunCases(t, tee.Run, tests)
}

// TestRunFiles covers writing, appending, truncating, and partial failures.
func TestRunFiles(t *testing.T) {
	t.Run("writes stdin to one file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")
		got := testutil.Run(t, tee.Run, "hello\n", path)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "hello\n")
		testutil.AssertEqual(t, "stderr", got.Stderr, "")
		testutil.AssertEqual(t, "file content", readFile(t, path), "hello\n")
	})

	t.Run("writes stdin to multiple files", func(t *testing.T) {
		dir := t.TempDir()
		first := filepath.Join(dir, "first.txt")
		second := filepath.Join(dir, "second.txt")
		got := testutil.Run(t, tee.Run, "shared\n", first, second)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "shared\n")
		testutil.AssertEqual(t, "first file", readFile(t, first), "shared\n")
		testutil.AssertEqual(t, "second file", readFile(t, second), "shared\n")
	})

	t.Run("truncates an existing file by default", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")
		writeFile(t, path, "old\n")

		got := testutil.Run(t, tee.Run, "new\n", path)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "new\n")
		testutil.AssertEqual(t, "file content", readFile(t, path), "new\n")
	})

	t.Run("append creates a missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")

		got := testutil.Run(t, tee.Run, "new\n", "-a", path)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "new\n")
		testutil.AssertEqual(t, "file content", readFile(t, path), "new\n")
	})

	t.Run("append keeps existing content", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")
		writeFile(t, path, "old\n")

		got := testutil.Run(t, tee.Run, "new\n", "-a", path)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "new\n")
		testutil.AssertEqual(t, "file content", readFile(t, path), "old\nnew\n")
	})

	t.Run("empty stdin creates an empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")

		got := testutil.Run(t, tee.Run, "", path)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "")
		testutil.AssertEqual(t, "file content", readFile(t, path), "")
	})

	t.Run("options after operands", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")
		writeFile(t, path, "old\n")

		got := testutil.Run(t, tee.Run, "new\n", path, "-a")

		testutil.AssertEqual(t, "exit code", got.ExitCode, 0)
		testutil.AssertEqual(t, "stdout", got.Stdout, "new\n")
		testutil.AssertEqual(t, "file content", readFile(t, path), "old\nnew\n")
	})

	t.Run("failed target does not stop other outputs", func(t *testing.T) {
		dir := t.TempDir()
		bad := filepath.Join(dir, "subdir")
		if err := os.Mkdir(bad, 0o755); err != nil {
			t.Fatalf("create directory %s: %v", bad, err)
		}
		good := filepath.Join(dir, "good.txt")

		got := testutil.Run(t, tee.Run, "x\n", bad, good)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 1)
		testutil.AssertEqual(t, "stdout", got.Stdout, "x\n")
		testutil.AssertContains(t, "stderr", got.Stderr, "tee: "+bad)
		testutil.AssertEqual(t, "good file", readFile(t, good), "x\n")
	})

	t.Run("failing only target still copies to stdout", func(t *testing.T) {
		dir := t.TempDir()
		bad := filepath.Join(dir, "subdir")
		if err := os.Mkdir(bad, 0o755); err != nil {
			t.Fatalf("create directory %s: %v", bad, err)
		}

		got := testutil.Run(t, tee.Run, "x\n", bad)

		testutil.AssertEqual(t, "exit code", got.ExitCode, 1)
		testutil.AssertEqual(t, "stdout", got.Stdout, "x\n")
		testutil.AssertContains(t, "stderr", got.Stderr, "tee: "+bad)
	})
}

// TestRunErrors covers usage errors.
func TestRunErrors(t *testing.T) {
	tests := []testutil.Case{
		{Name: "unknown option", Args: []string{"-z"}, WantExit: 2, WantStderr: "tee: unrecognized option: -z\n"},
	}

	testutil.RunCases(t, tee.Run, tests)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

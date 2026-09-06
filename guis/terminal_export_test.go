package guis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTerminalOutputPreservesCompleteRenderedTextPrivately(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "session.txt")
	want := "oldest line\nUnicode 中文 · Привет\nnewest line\n"

	if err := writeTerminalOutput(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("exported output = %q, want exact rendered text %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0o600 {
		t.Fatalf("export mode = %o, want 600", gotMode)
	}
}

func TestWriteTerminalOutputRejectsEmptyPathAndOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.txt")
	for name, candidate := range map[string]struct {
		path   string
		output string
	}{
		"empty path":   {path: "", output: "terminal output"},
		"empty output": {path: path, output: ""},
	} {
		t.Run(name, func(t *testing.T) {
			if err := writeTerminalOutput(candidate.path, candidate.output); err == nil {
				t.Fatal("expected export validation error")
			}
		})
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("rejected export created a file: %v", err)
	}
}

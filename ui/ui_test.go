package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewModelLoadsDocumentBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("# Hello\n\nSome text here.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := newModel(Config{Path: path}, "").(model)
	if m.state != stateShowDocument {
		t.Fatalf("expected stateShowDocument, got %s", m.state)
	}
	if !strings.Contains(m.pager.currentDocument.Body, "Some text") {
		t.Errorf("expected document body to be loaded, got %q", m.pager.currentDocument.Body)
	}
}

func TestTTYOptionsKeepsUsedTTYOpen(t *testing.T) {
	newTTY := func(t *testing.T) *os.File {
		t.Helper()
		f, err := os.CreateTemp(t.TempDir(), "tty")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })
		return f
	}
	write := func(f *os.File) error {
		_, err := f.WriteString("x")
		return err
	}

	for _, tc := range []struct {
		name        string
		stdinIsTTY  bool
		stdoutIsTTY bool
	}{
		{"stdin terminal, stdout piped", true, false},
		{"stdin piped, stdout terminal", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("shared handle", func(t *testing.T) {
				// tea.OpenTTY returns a single /dev/tty file for both sides on Unix.
				tty := newTTY(t)
				ttyOptions(tc.stdinIsTTY, tc.stdoutIsTTY, tty, tty)
				if err := write(tty); err != nil {
					t.Fatalf("shared tty closed while still in use: %v", err)
				}
			})

			t.Run("separate handles", func(t *testing.T) {
				// tea.OpenTTY returns CONIN$ and CONOUT$ separately on Windows.
				ttyIn, ttyOut := newTTY(t), newTTY(t)
				ttyOptions(tc.stdinIsTTY, tc.stdoutIsTTY, ttyIn, ttyOut)
				used, unused := ttyIn, ttyOut
				if tc.stdinIsTTY {
					used, unused = ttyOut, ttyIn
				}
				if err := write(used); err != nil {
					t.Fatalf("tty in use was closed: %v", err)
				}
				if err := write(unused); err == nil {
					t.Fatal("unused tty was left open")
				}
			})
		})
	}
}

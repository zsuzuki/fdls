package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListFilesDepthAndRelativePaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "root.txt"), "root")
	writeFile(t, filepath.Join(dir, "a", "child.txt"), "child")
	writeFile(t, filepath.Join(dir, "a", "b", "deep.txt"), "deep")

	var out bytes.Buffer
	if err := run([]string{"-depth", "1", dir}, &out); err != nil {
		t.Fatal(err)
	}

	got := strings.Split(strings.TrimSpace(out.String()), "\n")
	want := []string{filepath.Join("a", "child.txt"), "root.txt"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

func TestListFilesAllowsSpacesInRootAndFileNames(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Program Files")
	path := filepath.Join(dir, "app data.txt")
	writeFile(t, path, "hello")

	var out bytes.Buffer
	args := append([]string{"-size"}, strings.Split(dir, " ")...)
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}

	fields := strings.Split(strings.TrimSpace(out.String()), "\t")
	if len(fields) != 2 {
		t.Fatalf("expected tab-delimited 2 columns, got %d: %q", len(fields), out.String())
	}
	if fields[0] != "app data.txt" {
		t.Fatalf("expected filename with spaces, got %q", fields[0])
	}
	if fields[1] != "5" {
		t.Fatalf("expected size 5, got %q", fields[1])
	}
}

func TestListFilesAbsolutePathSizeDateAndSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	writeFile(t, path, "hello")

	var out bytes.Buffer
	if err := run([]string{"-path", "abs", "-size", "-date", "-sha256", "-depth", "0", dir}, &out); err != nil {
		t.Fatal(err)
	}

	fields := strings.Split(strings.TrimSpace(out.String()), "\t")
	if len(fields) != 4 {
		t.Fatalf("expected 4 columns, got %d: %q", len(fields), out.String())
	}
	if fields[0] != path {
		t.Fatalf("expected absolute path %q, got %q", path, fields[0])
	}
	if fields[1] != "5" {
		t.Fatalf("expected size 5, got %q", fields[1])
	}

	sum := sha256.Sum256([]byte("hello"))
	if fields[3] != hex.EncodeToString(sum[:]) {
		t.Fatalf("unexpected sha256: %q", fields[3])
	}
}

func TestInvalidPathMode(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"-path", "full", "."}, &out)
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

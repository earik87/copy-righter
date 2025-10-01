package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const copyright = "Copyright (c) 2025 Example Corp. All rights reserved."

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return file
}

func runCLI(t *testing.T, files ...string) string {
	t.Helper()
	args := append([]string{"run", "main.go", "--copyright=" + copyright}, files...)
	cmd := exec.Command("go", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, string(out))
	}
	return string(out)
}

func readFile(t *testing.T, file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(b)
}

func TestAddCopyrightToEmptyFile(t *testing.T) {
	file := writeTempFile(t, "")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not added to empty file: %q", content)
	}
}

func TestAddCopyrightToFileWithCode(t *testing.T) {
	file := writeTempFile(t, "package main\nfunc main() {}\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not added to file with code: %q", content)
	}
}

func TestIdempotency(t *testing.T) {
	file := writeTempFile(t, "package main\nfunc main() {}\n")
	runCLI(t, file)
	first := readFile(t, file)
	runCLI(t, file)
	second := readFile(t, file)
	if first != second {
		t.Errorf("file changed after second run; not idempotent")
	}
}

func TestUpdateOutdatedCopyright(t *testing.T) {
	file := writeTempFile(t, "// Old copyright\npackage main\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not updated: %q", content)
	}
}

func TestReplaceNonCopyrightComment(t *testing.T) {
	file := writeTempFile(t, "// Just a comment\npackage main\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("non-copyright comment not replaced: %q", content)
	}
}

func TestWhitespaceOnlyFile(t *testing.T) {
	file := writeTempFile(t, "   \n\n\t\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not added to whitespace file: %q", content)
	}
}

func TestShebangFile(t *testing.T) {
	file := writeTempFile(t, "#!/usr/bin/env bash\necho hi\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not added after shebang: %q", content)
	}
}

func TestReadOnlyFile(t *testing.T) {
	file := writeTempFile(t, "package main\n")
	if err := os.Chmod(file, 0400); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	cmd := exec.Command("go", "run", "main.go", "--copyright="+copyright, file)
	_ = cmd.Run()
	// Should not panic or crash; error is expected
	_ = os.Chmod(file, 0600) // restore for cleanup
}

func TestNonExistentFile(t *testing.T) {
	cmd := exec.Command("go", "run", "main.go", "--copyright="+copyright, "no_such_file.go")
	_ = cmd.Run() // Should not panic or crash
}

func TestUnicodeFile(t *testing.T) {
	file := writeTempFile(t, "package main // π\nfunc main() {} // 你好\n")
	runCLI(t, file)
	content := readFile(t, file)
	if !strings.HasPrefix(content, "// "+copyright) {
		t.Errorf("copyright not added to unicode file: %q", content)
	}
}

func TestRecursiveDirectory(t *testing.T) {
	dir := t.TempDir()
	file1 := filepath.Join(dir, "a.go")
	file2 := filepath.Join(dir, "b.go")
	os.WriteFile(file1, []byte("package main\n"), 0644)
	os.WriteFile(file2, []byte("package main\n"), 0644)
	runCLI(t, dir)
	for _, file := range []string{file1, file2} {
		content := readFile(t, file)
		if !strings.HasPrefix(content, "// "+copyright) {
			t.Errorf("copyright not added to %s: %q", file, content)
		}
	}
}

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
	expected := "// " + copyright
	if !strings.HasPrefix(content, expected) {
		t.Errorf("copyright header not added to empty file: %q", content)
	}
	if !strings.HasSuffix(strings.TrimSpace(content), expected) {
		t.Errorf("copyright footer not added to empty file: %q", content)
	}
}

func TestAddCopyrightToFileWithCode(t *testing.T) {
	file := writeTempFile(t, "package main\nfunc main() {}\n")
	runCLI(t, file)
	content := readFile(t, file)
	expected := "// " + copyright
	if !strings.HasPrefix(content, expected) {
		t.Errorf("copyright header not added to file with code: %q", content)
	}
	if !strings.HasSuffix(strings.TrimSpace(content), expected) {
		t.Errorf("copyright footer not added to file with code: %q", content)
	}
	lines := strings.Split(content, "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines (header, blank, code, footer), got %d", len(lines))
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
	if err := os.WriteFile(file1, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	runCLI(t, dir)
	for _, file := range []string{file1, file2} {
		content := readFile(t, file)
		if !strings.HasPrefix(content, "// "+copyright) {
			t.Errorf("copyright not added to %s: %q", file, content)
		}
	}
}

// New comprehensive tests for header and footer logic

func TestFooterAddedToSimpleFile(t *testing.T) {
	file := writeTempFile(t, "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Check header
	if !strings.HasPrefix(lines[0], "// "+copyright) {
		t.Errorf("header not added correctly: %q", lines[0])
	}

	// Check footer (last non-empty line)
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine == "" && len(lines) > 1 {
		lastLine = strings.TrimSpace(lines[len(lines)-2])
	}
	if lastLine != "// "+copyright {
		t.Errorf("footer not added correctly: %q", lastLine)
	}
}

func TestHeaderAndFooterIdempotency(t *testing.T) {
	file := writeTempFile(t, "package main\n\nfunc main() {}\n")

	// First run - adds header and footer
	runCLI(t, file)
	first := readFile(t, file)

	// Second run - should not change anything
	runCLI(t, file)
	second := readFile(t, file)

	if first != second {
		t.Errorf("file changed after second run:\nFirst:\n%s\n\nSecond:\n%s", first, second)
	}

	// Third run - still no change
	runCLI(t, file)
	third := readFile(t, file)

	if first != third {
		t.Errorf("file changed after third run")
	}
}

func TestUpdateOutdatedHeader(t *testing.T) {
	initial := "// Old copyright header\n\npackage main\n\nfunc main() {}\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	if lines[0] != "// "+copyright {
		t.Errorf("outdated header not updated: %q", lines[0])
	}
}

func TestUpdateOutdatedFooter(t *testing.T) {
	initial := "package main\n\nfunc main() {}\n\n// Old copyright footer\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	lastLine := lines[len(lines)-1]
	if lastLine != "// "+copyright {
		t.Errorf("outdated footer not updated: %q", lastLine)
	}
}

func TestUpdateBothOutdatedHeaderAndFooter(t *testing.T) {
	initial := "// Old header\n\npackage main\n\nfunc main() {}\n\n// Old footer\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Check header
	if lines[0] != "// "+copyright {
		t.Errorf("header not updated: %q", lines[0])
	}

	// Check footer
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine == "" && len(lines) > 1 {
		lastLine = strings.TrimSpace(lines[len(lines)-2])
	}
	if lastLine != "// "+copyright {
		t.Errorf("footer not updated: %q", lastLine)
	}
}

func TestFileWithOnlyHeader(t *testing.T) {
	initial := "// " + copyright + "\n\npackage main\n\nfunc main() {}\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Header should remain
	if lines[0] != "// "+copyright {
		t.Errorf("header changed unexpectedly: %q", lines[0])
	}

	// Footer should be added
	lastLine := lines[len(lines)-1]
	if lastLine != "// "+copyright {
		t.Errorf("footer not added when only header existed: %q", lastLine)
	}
}

func TestFileWithOnlyFooter(t *testing.T) {
	initial := "package main\n\nfunc main() {}\n\n// " + copyright + "\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Header should be added
	if lines[0] != "// "+copyright {
		t.Errorf("header not added when only footer existed: %q", lines[0])
	}

	// Footer should remain
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine == "" && len(lines) > 1 {
		lastLine = strings.TrimSpace(lines[len(lines)-2])
	}
	if lastLine != "// "+copyright {
		t.Errorf("footer changed unexpectedly: %q", lastLine)
	}
}

func TestFileWithMultipleComments(t *testing.T) {
	initial := "// Some comment\n// Another comment\npackage main\n\nfunc main() {}\n// End comment\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// First comment line should be replaced with copyright
	if lines[0] != "// "+copyright {
		t.Errorf("first comment not replaced: %q", lines[0])
	}

	// Last comment line should be replaced with copyright
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine == "" && len(lines) > 1 {
		lastLine = strings.TrimSpace(lines[len(lines)-2])
	}
	if lastLine != "// "+copyright {
		t.Errorf("last comment not replaced: %q", lastLine)
	}
}

func TestBlankLinesPreservation(t *testing.T) {
	initial := "package main\n\nfunc main() {}\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Should have: copyright, blank, package, blank, func, blank, copyright
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines with proper spacing, got %d", len(lines))
	}

	// Line after header should be blank
	if lines[1] != "" {
		t.Errorf("expected blank line after header, got: %q", lines[1])
	}
}

func TestFileWithTrailingNewline(t *testing.T) {
	initial := "package main\n\nfunc main() {}\n\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)

	expected := "// " + copyright
	if !strings.Contains(content, expected) {
		t.Errorf("copyright not found in file with trailing newlines")
	}

	// Count occurrences - should have exactly 2 (header and footer)
	count := strings.Count(content, expected)
	if count != 2 {
		t.Errorf("expected 2 copyright lines, got %d", count)
	}
}

func TestFileWithNoTrailingNewline(t *testing.T) {
	initial := "package main\n\nfunc main() {}"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)

	expected := "// " + copyright
	count := strings.Count(content, expected)
	if count != 2 {
		t.Errorf("expected 2 copyright lines, got %d", count)
	}
}

func TestLargeFileWithManyLines(t *testing.T) {
	var builder strings.Builder
	builder.WriteString("package main\n\n")
	for i := 0; i < 1000; i++ {
		builder.WriteString("func test")
		builder.WriteString(string(rune(i)))
		builder.WriteString("() {}\n")
	}

	file := writeTempFile(t, builder.String())
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Check header
	if lines[0] != "// "+copyright {
		t.Errorf("header not added to large file")
	}

	// Check footer
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine == "" && len(lines) > 1 {
		lastLine = strings.TrimSpace(lines[len(lines)-2])
	}
	if lastLine != "// "+copyright {
		t.Errorf("footer not added to large file")
	}
}

func TestSingleLineFile(t *testing.T) {
	initial := "package main"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// Should have header, blank, original line, blank, footer
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines, got %d: %v", len(lines), lines)
	}

	if lines[0] != "// "+copyright {
		t.Errorf("header not correct: %q", lines[0])
	}

	lastNonEmpty := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastNonEmpty = lines[i]
			break
		}
	}
	if lastNonEmpty != "// "+copyright {
		t.Errorf("footer not correct: %q", lastNonEmpty)
	}
}

func TestFileWithOnlyComments(t *testing.T) {
	initial := "// This is a comment\n// Another comment\n// Yet another\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)
	lines := strings.Split(content, "\n")

	// First line should be copyright
	if lines[0] != "// "+copyright {
		t.Errorf("first line not replaced with copyright: %q", lines[0])
	}

	// Last non-empty line should be copyright
	lastLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastLine = lines[i]
			break
		}
	}
	if lastLine != "// "+copyright {
		t.Errorf("last line not replaced with copyright: %q", lastLine)
	}
}

func TestMultipleFilesHeaderAndFooter(t *testing.T) {
	file1 := writeTempFile(t, "package main\nfunc main() {}\n")
	dir := filepath.Dir(file1)
	file2 := filepath.Join(dir, "test2.go")
	if err := os.WriteFile(file2, []byte("package utils\nfunc helper() {}\n"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	runCLI(t, file1, file2)

	for _, file := range []string{file1, file2} {
		content := readFile(t, file)
		expected := "// " + copyright

		if !strings.HasPrefix(content, expected) {
			t.Errorf("header not added to %s", file)
		}

		if !strings.Contains(content[len(expected):], expected) {
			t.Errorf("footer not added to %s", file)
		}

		count := strings.Count(content, expected)
		if count != 2 {
			t.Errorf("expected 2 copyrights in %s, got %d", file, count)
		}
	}
}

// Tests for trailing newline behavior

func TestNewFooterAddsTrailingNewline(t *testing.T) {
	// Test that when adding a NEW footer, a trailing newline is added (Go idiomatic)
	initial := "package main\n\nfunc main() {}\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)

	// File should end with a newline after the footer
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("expected trailing newline after adding new footer, but file doesn't end with newline")
	}

	// Should end with: "// <copyright>\n"
	expectedEnding := "// " + copyright + "\n"
	if !strings.HasSuffix(content, expectedEnding) {
		t.Errorf("expected file to end with %q, got %q", expectedEnding, content[len(content)-len(expectedEnding):])
	}
}

func TestUpdatingFooterWithoutTrailingNewlinePreservesFormat(t *testing.T) {
	// Test that when UPDATING an existing footer that has NO trailing newline,
	// we preserve that format (developer's responsibility)
	initial := "package main\n\nfunc main() {}\n\n// Old copyright footer"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)

	// The updated footer should NOT have a trailing newline added
	expectedEnding := "// " + copyright
	actualEnding := strings.TrimSuffix(content, "\n")
	if !strings.HasSuffix(actualEnding, expectedEnding) {
		t.Errorf("expected file to end with %q (no trailing newline), got %q", expectedEnding, actualEnding)
	}

	// Verify it doesn't end with newline after the copyright
	if strings.HasSuffix(content, "// "+copyright+"\n") {
		t.Errorf("should NOT have added trailing newline when updating existing footer without one")
	}
}

func TestUpdatingFooterWithTrailingNewlinePreservesIt(t *testing.T) {
	// Test that when UPDATING an existing footer that HAS a trailing newline,
	// we preserve that format
	initial := "package main\n\nfunc main() {}\n\n// Old copyright footer\n"
	file := writeTempFile(t, initial)
	runCLI(t, file)
	content := readFile(t, file)

	// The updated footer should preserve the trailing newline
	expectedEnding := "// " + copyright + "\n"
	if !strings.HasSuffix(content, expectedEnding) {
		t.Errorf("expected file to end with %q (with trailing newline), got %q", expectedEnding, content[len(content)-100:])
	}
}

func TestIdempotencyPreservesTrailingNewline(t *testing.T) {
	// Test that running multiple times preserves the trailing newline from first run
	initial := "package main\n\nfunc main() {}\n"
	file := writeTempFile(t, initial)

	// First run - adds footer with trailing newline
	runCLI(t, file)
	firstRun := readFile(t, file)

	if !strings.HasSuffix(firstRun, "\n") {
		t.Errorf("first run should add trailing newline")
	}

	// Second run - should preserve exact format
	runCLI(t, file)
	secondRun := readFile(t, file)

	if firstRun != secondRun {
		t.Errorf("second run changed file content:\nFirst:\n%q\n\nSecond:\n%q", firstRun, secondRun)
	}

	// Third run - still identical
	runCLI(t, file)
	thirdRun := readFile(t, file)

	if firstRun != thirdRun {
		t.Errorf("third run changed file content")
	}
}

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func formatCopyrightLine(copyrightText string) string {
	trimmed := strings.TrimSpace(copyrightText)
	if strings.HasPrefix(trimmed, "//") {
		return trimmed
	}
	return "// " + trimmed
}

func processFile(filePath, copyrightText string) error {
	copyrightLine := formatCopyrightLine(copyrightText)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "Error closing file %s: %v\n", filePath, cerr)
		}
	}()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	if len(lines) == 0 {
		// Empty file, just add copyright
		return os.WriteFile(filePath, []byte(copyrightLine+"\n\n"), 0644)
	}

	firstLine := lines[0]
	currentHash := hashString(firstLine)
	if currentHash == hashString(copyrightLine) {
		fmt.Printf("Copyright already up to date in: %s\n", filePath)
		return nil
	}

	if strings.HasPrefix(firstLine, "//") {
		fmt.Printf("Updating copyright in: %s (hash mismatch)\n", filePath)
		// Replace first line
		newContent := append([]string{copyrightLine, ""}, lines[1:]...)
		return os.WriteFile(filePath, []byte(strings.Join(newContent, "\n")), 0644)
	}

	// No copyright found, add at top
	fmt.Printf("Adding copyright to: %s\n", filePath)
	newContent := append([]string{copyrightLine, ""}, lines...)
	return os.WriteFile(filePath, []byte(strings.Join(newContent, "\n")), 0644)
}

func runCopyright(cmd *cobra.Command, args []string) {
	copyrightText, _ := cmd.Flags().GetString("copyright")
	if copyrightText == "" || len(args) == 0 {
		fmt.Println("Usage: copy-righter --copyright='Your copyright' file1 [file2 ...]")
		os.Exit(1)
	}
	for _, file := range args {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		if info.IsDir() {
			err := filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return fmt.Errorf("error accessing path %s: %w", path, err)
				}
				if info.IsDir() {
					return nil
				}
				if err := processFile(path, copyrightText); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", path, err)
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", file, err)
			}
		} else {
			if err := processFile(file, copyrightText); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", file, err)
			}
		}
	}
}

func main() {
	var copyrightText string

	rootCmd := &cobra.Command{
		Use:   "copy-righter [flags] file1 [file2 ...]",
		Short: "A CLI tool to check and add copyright headers to files.",
		Args:  cobra.MinimumNArgs(1),
		Run:   runCopyright,
	}
	rootCmd.Flags().StringVar(&copyrightText, "copyright", "", "Copyright text to add (required)")
	if err := rootCmd.MarkFlagRequired("copyright"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag required: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

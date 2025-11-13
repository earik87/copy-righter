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

func processFile(filePath, copyrightText string) (modified bool, err error) {
	copyrightLine := formatCopyrightLine(copyrightText)
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
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
		return false, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	if len(lines) == 0 {
		// Empty file, just add copyright header and footer
		err := os.WriteFile(filePath, []byte(copyrightLine+"\n\n"+copyrightLine+"\n"), 0644)
		return true, err
	}

	headerUpdated := false
	footerUpdated := false

	// Check and update header
	firstLine := lines[0]
	currentHash := hashString(firstLine)
	if currentHash == hashString(copyrightLine) {
		fmt.Printf("Copyright header already up to date in: %s\n", filePath)
		headerUpdated = true
	} else if strings.HasPrefix(firstLine, "//") {
		fmt.Printf("Updating copyright header in: %s (hash mismatch)\n", filePath)
		lines[0] = copyrightLine
		if len(lines) > 1 && lines[1] == "" {
			// Keep blank line after header
		} else {
			lines = append([]string{copyrightLine, ""}, lines[1:]...)
		}
		headerUpdated = true
	} else {
		// No copyright found, add at top
		fmt.Printf("Adding copyright header to: %s\n", filePath)
		lines = append([]string{copyrightLine, ""}, lines...)
		headerUpdated = true
	}

	// Check and update footer
	lastLine := lines[len(lines)-1]
	lastLineHash := hashString(lastLine)
	if lastLineHash == hashString(copyrightLine) {
		fmt.Printf("Copyright footer already up to date in: %s\n", filePath)
		footerUpdated = true
	} else if strings.HasPrefix(lastLine, "//") {
		fmt.Printf("Updating copyright footer in: %s (hash mismatch)\n", filePath)
		// Check if there's a blank line before the footer comment
		if len(lines) > 1 && lines[len(lines)-2] == "" {
			lines[len(lines)-1] = copyrightLine
		} else {
			lines[len(lines)-1] = copyrightLine
			lines = append(lines[:len(lines)-1], "", copyrightLine)
		}
		footerUpdated = true
	} else {
		// No copyright footer found, add at bottom
		fmt.Printf("Adding copyright footer to: %s\n", filePath)
		lines = append(lines, "", copyrightLine)
		footerUpdated = true
	}

	if !headerUpdated && !footerUpdated {
		fmt.Printf("Copyright already up to date in: %s\n", filePath)
		return false, nil
	}

	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	return true, err
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
					fmt.Fprintf(os.Stderr, "Error accessing path %s: %v\n", path, err)
					return nil // Continue walking
				}

				if info.IsDir() {
					fmt.Printf("Skipping directory: %s\n", path)
					return nil
				}

				if !isSupportedFile(path) {
					fmt.Printf("Skipping unsupported file: %s\n", path)
					return nil
				}

				fmt.Printf("Processing file: %s\n", path)
				if _, err := processFile(path, copyrightText); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", path, err)
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error walking directory %s: %v\n", file, err)
			}
		} else {
			if _, err := processFile(file, copyrightText); err != nil {
				fmt.Fprintf(os.Stderr, "Error processing file %s: %v\n", file, err)
			}
		}
	}
}

func isSupportedFile(filePath string) bool {
	supportedExtensions := []string{".go"}
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
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

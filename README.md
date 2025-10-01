# copy-righter

`copy-righter` is a CLI tool that automatically adds or updates copyright headers in source code files. It supports multiple programming languages and can process individual files or entire directories recursively.

## Features
- Automatically detects and adds missing copyright headers.
- Updates outdated copyright headers.
- Supports `.go` files.
- Processes individual files or entire directories.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/earik87/copy-righter.git
   ```

2. Navigate to the project directory:
   ```bash
   cd copy-righter
   ```

3. Build the CLI tool:
   ```bash
   go install
   ```

## Usage

### Basic Command
```bash
copy-righter --copyright="© 2025 Example Corp. All rights reserved." <file_or_directory>
```

### Examples

1. Add copyright to a single file:
   ```bash
   copy-righter --copyright="© 2025 Example Corp. All rights reserved." main.go
   ```

2. Add copyright to all files in a directory:
   ```bash
   copy-righter --copyright="© 2025 Example Corp. All rights reserved." ./src
   ```

## Supported File Types
- `.go`

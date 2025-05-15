#!/bin/bash

copyright_text="// © 2024 - 2025, CompanyName. All rights reserved"

# Function to add copyright to a file if it doesn't already exist
add_copyright() {
    local file="$1"
    
    # Skip if file already has the copyright text
    if grep -q "© 2024 - 2025, CompanyName." "$file"; then
        echo "Copyright already exists in: $file"
        return
    fi    # Added missing 'fi' here

    # Create temp file
    temp_file=$(mktemp)
    
    # Add copyright as first line, then original content
    echo "$copyright_text" > "$temp_file"
    echo "" >> "$temp_file" # Add blank line after copyright
    cat "$file" >> "$temp_file"
    
    # Replace original with updated content
    mv "$temp_file" "$file"
    echo "Added copyright to: $file"
}

# Check if directory argument provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <directory>"
    exit 1
fi

directory="$1"

# Recurse through directory and process each file
find "$directory" -type f \( -name "*.go" -o -name "*.js" -o -name "*.py" -o -name "*.sh" -o -name "*.java" -o -name "*.cpp" -o -name "*.h" \) -print0 | while IFS= read -r -d '' file; do
    add_copyright "$file"
done

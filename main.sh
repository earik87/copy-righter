#!/bin/bash

# Copyright constants - Make these configurable
COPYRIGHT_START_YEAR="2013"
COPYRIGHT_END_YEAR=$(date +%Y)
COPYRIGHT_COMPANY="CompanyName. All rights reserved"

# Extract company identifiers for detection (make this generic)
# This will extract key words from the company name for copyright detection
COPYRIGHT_IDENTIFIERS=$(echo "$COPYRIGHT_COMPANY" | grep -oE '\b[A-Z][a-zA-Z]{3,}\b' | tr '\n' '|' | sed 's/|$//')

# Construct the complete copyright text
copyright_text="// © ${COPYRIGHT_START_YEAR} - ${COPYRIGHT_END_YEAR}, ${COPYRIGHT_COMPANY}"

# Calculate hash of the expected copyright text for fast comparison
expected_copyright_hash=$(echo "$copyright_text" | shasum -a 256 | cut -d' ' -f1)

# Function to add or update copyright in a file
add_copyright() {
    local file="$1"
    
    # Check if file already has any copyright from this company (generic detection)
    if grep -qE "$COPYRIGHT_IDENTIFIERS" "$file"; then
        # Extract the first line (assumed to be copyright) and hash it
        first_line=$(head -n1 "$file")
        if echo "$first_line" | grep -qE "$COPYRIGHT_IDENTIFIERS"; then
            current_copyright_hash=$(echo "$first_line" | shasum -a 256 | cut -d' ' -f1)

            # Fast hash comparison instead of string comparison
            if [ "$current_copyright_hash" = "$expected_copyright_hash" ]; then
                echo "Copyright already up to date in: $file"
                return
            else
                echo "Updating copyright in: $file (hash mismatch)"
                # Remove first line (old copyright) and add new one
                temp_file=$(mktemp)
                tail -n +2 "$file" > "$temp_file"

                # Add new copyright as first line
                {
                    echo "$copyright_text"
                    echo ""
                    cat "$temp_file"
                } > "${temp_file}.new"

                mv "${temp_file}.new" "$file"
                rm -f "$temp_file"
                echo "Updated copyright in: $file"
            fi
        else
            # Copyright exists but not on first line, use the old method
            echo "Updating copyright in: $file (not on first line)"
            temp_file=$(mktemp)
            grep -vE "© .*($COPYRIGHT_IDENTIFIERS)" "$file" > "$temp_file"

            {
                echo "$copyright_text"
                echo ""
                cat "$temp_file"
            } > "${temp_file}.new"

            mv "${temp_file}.new" "$file"
            rm -f "$temp_file"
            echo "Updated copyright in: $file"
        fi
    else
        # No existing copyright, add it
        temp_file=$(mktemp)

        # Add copyright as first line, then original content
        echo "$copyright_text" > "$temp_file"
        echo "" >> "$temp_file" # Add blank line after copyright
        cat "$file" >> "$temp_file"

        # Replace original with updated content
        mv "$temp_file" "$file"
        echo "Added copyright to: $file"
    fi
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

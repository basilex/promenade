#!/usr/bin/env python3
"""
Add testing.Short() skip check to all integration test functions.
Usage: python3 scripts/add_integration_test_skip.py
"""

import os
import re
from pathlib import Path

def add_short_skip(file_path):
    """Add short mode skip to all test functions in a file."""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Pattern: func TestXxx(t *testing.T) {
    pattern = r'(func Test\w+\(t \*testing\.T\) \{\n)'
    
    # Skip check to add
    skip_check = r'\1\tif testing.Short() {\n\t\tt.Skip("Skipping integration test in short mode")\n\t}\n'
    
    # Replace only if not already present
    if 'testing.Short()' not in content:
        new_content = re.sub(pattern, skip_check, content)
        
        if new_content != content:
            with open(file_path, 'w') as f:
                f.write(new_content)
            return True
    return False

def main():
    """Process all integration test files."""
    test_dir = Path('test/integration/contexts')
    
    if not test_dir.exists():
        print(f"Error: {test_dir} does not exist")
        return
    
    modified_count = 0
    for test_file in test_dir.rglob('*_test.go'):
        if add_short_skip(test_file):
            print(f"✓ Modified: {test_file}")
            modified_count += 1
        else:
            print(f"○ Skipped: {test_file} (already has skip check)")
    
    print(f"\nDone! Modified {modified_count} files.")

if __name__ == '__main__':
    main()

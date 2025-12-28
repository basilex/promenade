#!/usr/bin/env python3
"""
Clean documentation files from:
1. Box-drawing characters (ASCII tables)
2. Unwanted emoji (except language flag emoji in specific contexts)
"""

import os
import re
import sys
from pathlib import Path

# Box-drawing characters to remove
BOX_CHARS = [
    '─', '│', '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼',
    '═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬',
    '╭', '╮', '╯', '╰', '╱', '╲', '╳',
]

# Emoji ranges to check (excluding flags which are combining characters)
EMOJI_PATTERN = re.compile(
    r'[\U0001F300-\U0001F5FF'  # symbols & pictographs
    r'\U0001F600-\U0001F64F'   # emoticons
    r'\U0001F680-\U0001F6FF'   # transport & map
    r'\U0001F700-\U0001F77F'   # alchemical
    r'\U0001F780-\U0001F7FF'   # Geometric Shapes Extended
    r'\U0001F800-\U0001F8FF'   # Supplemental Arrows-C
    r'\U0001F900-\U0001F9FF'   # Supplemental Symbols and Pictographs
    r'\U0001FA00-\U0001FA6F'   # Chess Symbols
    r'\U0001FA70-\U0001FAFF'   # Symbols and Pictographs Extended-A
    r'\U00002702-\U000027B0'   # Dingbats
    r'\U000024C2-\U0001F251]+', 
    flags=re.UNICODE
)

def should_process_file(filepath):
    """Check if file should be processed"""
    # Skip certain directories
    skip_dirs = {'node_modules', 'vendor', 'public', 'tmp'}
    
    parts = Path(filepath).parts
    for skip in skip_dirs:
        if skip in parts:
            return False
    
    return filepath.endswith('.md')

def clean_box_drawing(text):
    """Remove box-drawing characters and clean up the formatting"""
    for char in BOX_CHARS:
        text = text.replace(char, '')
    return text

def clean_emoji(text, filepath):
    """Remove ALL emoji from documentation"""
    # Remove all emoji without exceptions
    return EMOJI_PATTERN.sub('', text)

def clean_file(filepath):
    """Clean a single markdown file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Clean box-drawing characters
        content = clean_box_drawing(content)
        
        # Clean emoji
        content = clean_emoji(content, filepath)
        
        # Remove excessive blank lines (more than 2 in a row)
        content = re.sub(r'\n{4,}', '\n\n\n', content)
        
        # Only write if changed
        if content != original:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True
        
        return False
        
    except Exception as e:
        print(f"Error processing {filepath}: {e}", file=sys.stderr)
        return False

def main():
    root_dir = Path(__file__).parent.parent
    
    print(f"Scanning {root_dir} for markdown files...")
    
    processed = 0
    modified = 0
    
    for root, dirs, files in os.walk(root_dir):
        # Skip hidden directories except .github
        dirs[:] = [d for d in dirs if not d.startswith('.') or d == '.github']
        
        for file in files:
            filepath = os.path.join(root, file)
            
            if should_process_file(filepath):
                processed += 1
                if clean_file(filepath):
                    modified += 1
                    print(f"✓ Cleaned: {os.path.relpath(filepath, root_dir)}")
    
    print(f"\n✅ Processed {processed} files, modified {modified} files")

if __name__ == '__main__':
    main()

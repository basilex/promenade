#!/usr/bin/env python3
"""
Professional emoji and special character cleaner for Promenade Platform.

Removes ALL emoji from documentation and code files according to the official
Documentation Style Guide (docs/guides/documentation-style-guide.md).

Features:
- Removes all emoji characters (no exceptions)
- Removes box-drawing characters
- Processes .md, .go, .yaml, .yml, .json files
- Dry-run mode for safety
- Detailed reporting
- Enforces official no-emoji policy (effective January 5, 2026)

Usage:
    python scripts/clean-docs.py              # Dry-run mode (show what would change)
    python scripts/clean-docs.py --apply      # Actually modify files
    python scripts/clean-docs.py --check      # Exit 1 if emoji found (CI mode)

See: docs/guides/documentation-style-guide.md for complete policy
"""

import os
import re
import sys
import argparse
from pathlib import Path
from typing import List, Tuple

# Version and policy info
VERSION = "2.0.0"
POLICY_DATE = "January 5, 2026"

# Box-drawing characters to remove (legacy tables)
BOX_CHARS = [
    '─', '│', '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼',
    '═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬',
    '╭', '╮', '╯', '╰', '╱', '╲', '╳',
]

# Comprehensive emoji pattern (ALL emoji ranges)
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
    r'\U00002600-\U000027BF'   # Miscellaneous Symbols (includes common emoji)
    r'\U000024C2-\U0001F251]+', 
    flags=re.UNICODE
)

# Common emoji to text replacements (for manual review)
EMOJI_REPLACEMENTS = {
    '✅': 'DONE',
    '❌': 'FAILED',
    '⚠️': 'WARNING',
    '📋': 'TODO',
    '🎯': 'TARGET',
    '🏗️': 'IN_PROGRESS',
    '📚': 'DOCS',
    '🚀': 'DEPLOY',
    '🎉': 'COMPLETE',
    '🔥': 'CRITICAL',
    '⚡': 'FAST',
    '💡': 'IDEA',
    '🔒': 'SECURE',
    '🛡️': 'PROTECTED',
}


def should_process_file(filepath: str) -> bool:
    """Check if file should be processed based on extension and directory"""
    # Skip directories
    skip_dirs = {
        'node_modules', 'vendor', 'public', 'tmp', 'dist', 'build',
        '.git', 'bin', 'data', '__pycache__', '.pytest_cache'
    }
    
    parts = Path(filepath).parts
    for skip in skip_dirs:
        if skip in parts:
            return False
    
    # Process these file extensions
    allowed_extensions = {'.md', '.go', '.yaml', '.yml', '.json', '.toml'}
    return Path(filepath).suffix in allowed_extensions


def clean_box_drawing(text: str) -> str:
    """Remove box-drawing characters from old ASCII tables"""
    for char in BOX_CHARS:
        text = text.replace(char, '')
    return text


def find_emoji_occurrences(text: str) -> List[Tuple[int, str, str]]:
    """Find all emoji occurrences with line numbers and context"""
    occurrences = []
    lines = text.split('\n')
    
    for line_num, line in enumerate(lines, 1):
        matches = EMOJI_PATTERN.finditer(line)
        for match in matches:
            emoji = match.group()
            occurrences.append((line_num, emoji, line.strip()[:80]))
    
    return occurrences


def clean_emoji(text: str) -> str:
    """Remove ALL emoji - no exceptions per official policy"""
    return EMOJI_PATTERN.sub('', text)


def clean_file(filepath: str, apply: bool = False) -> Tuple[bool, int]:
    """
    Clean a single file from emoji and box-drawing characters
    
    Args:
        filepath: Path to file to clean
        apply: If True, write changes. If False, dry-run only.
    
    Returns:
        (changed, emoji_count): Whether file would change and how many emoji found
    """
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Find emoji for reporting
        emoji_occurrences = find_emoji_occurrences(content)
        
        # Clean box-drawing characters
        content = clean_box_drawing(content)
        
        # Clean emoji
        content = clean_emoji(content)
        
        # Remove excessive blank lines (more than 2 in a row)
        content = re.sub(r'\n{4,}', '\n\n\n', content)
        
        # Write if changed and apply=True
        changed = content != original
        if changed and apply:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
        
        return changed, len(emoji_occurrences)
        
    except Exception as e:
        print(f"ERROR processing {filepath}: {e}", file=sys.stderr)
        return False, 0


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description=f'Emoji cleaner v{VERSION} - Enforce no-emoji policy',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=f"""
Examples:
  {sys.argv[0]}              # Dry-run (show what would change)
  {sys.argv[0]} --apply      # Actually modify files
  {sys.argv[0]} --check      # CI mode (exit 1 if emoji found)

Policy: Official no-emoji policy effective {POLICY_DATE}
See: docs/guides/documentation-style-guide.md
        """
    )
    
    parser.add_argument(
        '--apply',
        action='store_true',
        help='Actually modify files (default: dry-run only)'
    )
    
    parser.add_argument(
        '--check',
        action='store_true',
        help='CI mode: exit 1 if any emoji found (for pre-commit hooks)'
    )
    
    parser.add_argument(
        '--version',
        action='version',
        version=f'%(prog)s {VERSION}'
    )
    
    args = parser.parse_args()
    
    root_dir = Path(__file__).parent.parent
    
    # Print header
    print("=" * 70)
    print(f"Promenade Emoji Cleaner v{VERSION}")
    print(f"Official Policy: No emoji (effective {POLICY_DATE})")
    print("=" * 70)
    
    if args.apply:
        print("MODE: Applying changes (files will be modified)")
    elif args.check:
        print("MODE: Check only (CI mode - will exit 1 if emoji found)")
    else:
        print("MODE: Dry-run (use --apply to actually modify files)")
    
    print(f"\nScanning {root_dir} for files...")
    print()
    
    processed = 0
    modified = 0
    total_emoji = 0
    files_with_emoji = []
    
    for root, dirs, files in os.walk(root_dir):
        # Skip hidden directories except .github
        dirs[:] = [d for d in dirs if not d.startswith('.') or d == '.github']
        
        for file in files:
            filepath = os.path.join(root, file)
            
            if should_process_file(filepath):
                processed += 1
                changed, emoji_count = clean_file(filepath, apply=args.apply)
                
                if changed:
                    modified += 1
                    total_emoji += emoji_count
                    rel_path = os.path.relpath(filepath, root_dir)
                    files_with_emoji.append((rel_path, emoji_count))
                    
                    status = "CLEANED" if args.apply else "WOULD CLEAN"
                    print(f"[{status}] {rel_path} ({emoji_count} emoji)")
    
    # Print summary
    print()
    print("=" * 70)
    print("SUMMARY")
    print("=" * 70)
    print(f"Processed files: {processed}")
    print(f"Files with emoji: {modified}")
    print(f"Total emoji found: {total_emoji}")
    
    if modified > 0:
        print()
        print("Files requiring cleanup:")
        for filepath, count in files_with_emoji:
            print(f"  - {filepath} ({count} emoji)")
    
    print()
    
    if args.check and modified > 0:
        print("POLICY VIOLATION: Emoji found in codebase!")
        print("Run 'make clean-emoji' or 'python scripts/clean-docs.py --apply' to fix")
        print()
        sys.exit(1)
    elif not args.apply and modified > 0:
        print("To apply changes, run:")
        print(f"  python {sys.argv[0]} --apply")
        print("Or use:")
        print("  make clean-emoji")
    elif args.apply and modified > 0:
        print(f"SUCCESS: Cleaned {modified} files, removed {total_emoji} emoji")
    else:
        print("SUCCESS: No emoji found - codebase is clean!")
    
    print()


if __name__ == '__main__':
    main()

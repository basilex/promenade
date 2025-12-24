#!/usr/bin/env python3
"""
Clean documentation files:
1. Remove pseudographics (box drawing characters)
2. Remove emoji (except country flags)
"""
import os
import re
import sys

def clean_file(filepath):
    """Clean pseudographics and emoji from a markdown file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        return False, f"Read error: {e}"
    
    original = content
    
    # Step 1: Remove box drawing table characters
    # Keep tree structure chars (├ └ │) for directory listings
    box_chars = r'[╔═╗║╚╝┌┐┘━┏┓┗┛┬┴┼╟╢╤╧╪╞╡╥╨╫]'
    content = re.sub(box_chars, '', content)
    
    # Step 2: Remove emoji (except country flags)
    # Country flags: 🇺🇦 🇬🇧 🇩🇪 🇵🇹 🇪🇸 etc (U+1F1E6-U+1F1FF pairs)
    # Remove: ✅ ❌ 🎉 📊 💡 etc
    
    # Common emoji to remove
    emoji_pattern = r'[✅✓❌✗⚠️💡📊🎉🔧🚀📝💻🌟⭐🔥💪👍👎😀😁🤔🏆✨🎯📈📉💼🏢🌍🌎🌏🗺️📍📌🔔🔕💬💭🗨️🗯️💾💿📀🖥️🖨️⌨️🖱️]'
    content = re.sub(emoji_pattern, '', content)
    
    # Remove emoji ranges (but NOT regional indicators U+1F1E6-U+1F1FF for flags)
    # Miscellaneous Symbols
    content = re.sub(r'[\u2600-\u26FF]', '', content)
    # Dingbats
    content = re.sub(r'[\u2700-\u27BF]', '', content)
    
    # Clean up excessive blank lines
    content = re.sub(r'\n{4,}', '\n\n\n', content)
    
    # Trim trailing whitespace from lines
    lines = content.split('\n')
    lines = [line.rstrip() for line in lines]
    content = '\n'.join(lines)
    
    if content != original:
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True, "Updated"
        except Exception as e:
            return False, f"Write error: {e}"
    
    return False, "No changes"

def main():
    # Find all markdown files
    files_to_process = []
    
    for root, dirs, files in os.walk('.'):
        # Skip hidden directories
        dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ['node_modules', 'vendor']]
        
        for file in files:
            if file.endswith('.md'):
                filepath = os.path.join(root, file)
                files_to_process.append(filepath)
    
    print(f"Found {len(files_to_process)} markdown files to process\n")
    
    updated = []
    skipped = []
    errors = []
    
    for filepath in sorted(files_to_process):
        changed, message = clean_file(filepath)
        if changed:
            updated.append(filepath)
            print(f"✓ {filepath}")
        elif "error" in message.lower():
            errors.append((filepath, message))
            print(f"✗ {filepath}: {message}")
        else:
            skipped.append(filepath)
    
    # Summary
    print(f"\n{'='*60}")
    print(f"SUMMARY")
    print(f"{'='*60}")
    print(f"✓ Updated: {len(updated)}")
    print(f"○ Skipped (no changes): {len(skipped)}")
    print(f"✗ Errors: {len(errors)}")
    print(f"Total processed: {len(files_to_process)}")
    
    if errors:
        print(f"\nErrors:")
        for filepath, message in errors:
            print(f"  {filepath}: {message}")
    
    return 0 if not errors else 1

if __name__ == '__main__':
    sys.exit(main())

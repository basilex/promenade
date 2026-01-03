#!/usr/bin/env python3
"""
Check all markdown links in the repository.
Verifies that all relative file links point to existing files.
"""

import os
import re
import sys
from pathlib import Path
from urllib.parse import urlparse

def is_external_link(link):
    """Check if link is external (http/https/mailto) or VitePress route."""
    # VitePress routes start with / and are internal to the site
    if link.startswith('/'):
        return True
    return link.startswith(('http://', 'https://', 'mailto:', '#'))

def resolve_path(base_file, link):
    """Resolve relative path from base file to target."""
    base_dir = os.path.dirname(base_file)
    # Remove anchor
    link_path = link.split('#')[0]
    if not link_path:
        return None
    
    # Resolve relative path
    target = os.path.normpath(os.path.join(base_dir, link_path))
    return target

def check_markdown_file(filepath, project_root):
    """Check all links in a markdown file."""
    broken_links = []
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find all markdown links [text](url)
    link_pattern = r'\[([^\]]+)\]\(([^)]+)\)'
    matches = re.findall(link_pattern, content)
    
    for text, link in matches:
        # Skip external links
        if is_external_link(link):
            continue
        
        # Resolve path
        target = resolve_path(filepath, link)
        if target is None:
            continue
        
        # Check if file/directory exists
        if not os.path.exists(target):
            broken_links.append({
                'file': filepath,
                'link': link,
                'text': text,
                'resolved': target
            })
    
    return broken_links

def main():
    project_root = Path('/Users/basilex/Workspace/src/promenade')
    
    print("🔍 Checking all markdown links in the repository...\n")
    
    # Find all markdown files
    md_files = []
    for pattern in ['**/*.md']:
        md_files.extend(project_root.glob(pattern))
    
    # Filter out node_modules, .vitepress, vendor
    md_files = [f for f in md_files 
                if 'node_modules' not in str(f) 
                and '.vitepress' not in str(f)
                and 'vendor' not in str(f)]
    
    all_broken = []
    total_files = len(md_files)
    
    for i, md_file in enumerate(md_files, 1):
        broken = check_markdown_file(str(md_file), str(project_root))
        if broken:
            all_broken.extend(broken)
        
        # Progress indicator
        if i % 10 == 0:
            print(f"Progress: {i}/{total_files} files checked...")
    
    print(f"\n📊 Summary:")
    print(f"   Total files checked: {total_files}")
    print(f"   Files with broken links: {len(set(b['file'] for b in all_broken))}")
    print(f"   Total broken links: {len(all_broken)}\n")
    
    if all_broken:
        print("❌ Broken links found:\n")
        for item in all_broken:
            rel_file = os.path.relpath(item['file'], project_root)
            print(f"File: {rel_file}")
            print(f"  Link: {item['link']}")
            print(f"  Text: [{item['text']}]")
            print(f"  Expected: {item['resolved']}")
            print()
        
        return 1
    else:
        print("✅ All links are valid!")
        return 0

if __name__ == '__main__':
    sys.exit(main())

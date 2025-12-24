#!/usr/bin/env python3
"""Clean pseudographics from LICENSE_ARCHITECTURE diagrams"""
import re

files = [
    'docs/LICENSE_ARCHITECTURE.md',
    'docs/uk/LICENSE_ARCHITECTURE.uk.md',
    'docs/de/LICENSE_ARCHITECTURE.de.md',
    'docs/pt/LICENSE_ARCHITECTURE.pt.md',
    'docs/es/LICENSE_ARCHITECTURE.es.md',
]

for filepath in files:
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Remove box drawing characters from flowcharts
    lines = content.split('\n')
    new_lines = []
    in_code_block = False
    
    for line in lines:
        # Track code blocks
        if line.strip().startswith('```'):
            in_code_block = not in_code_block
            new_lines.append(line)
            continue
        
        # If in code block, clean box drawing
        if in_code_block:
            # Skip lines that are only box drawing characters
            if re.match(r'^[\s─━│┌└┐┘├┤┬┴┼]+$', line):
                continue
            # Remove box characters but keep content
            line = re.sub(r'[─━│┌└┐┘├┤┬┴┼]', '', line)
            # Clean up excessive spaces
            line = re.sub(r'  +', ' ', line)
        
        new_lines.append(line)
    
    content = '\n'.join(new_lines)
    
    # Write back
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"✓ {filepath}")

print("\n✓ All LICENSE_ARCHITECTURE files cleaned!")

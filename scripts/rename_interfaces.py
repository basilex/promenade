#!/usr/bin/env python3
"""
Rename all Go interfaces to use I prefix (e.g., UserRepository → IUserRepository)
This improves code readability by making it immediately clear what is an interface.
"""
import os
import re
from pathlib import Path

# Interfaces to rename (interface_name -> IInterfaceName)
INTERFACE_RENAMES = {
    # Core UseCases
    'AuthUseCase': 'IAuthUseCase',
    'CountryUseCase': 'ICountryUseCase',
    'CurrencyUseCase': 'ICurrencyUseCase',
    'LanguageUseCase': 'ILanguageUseCase',
    'TimezoneUseCase': 'ITimezoneUseCase',
    'RegionUseCase': 'IRegionUseCase',
    'CityUseCase': 'ICityUseCase',
    'RoleUseCase': 'IRoleUseCase',
    'PermissionUseCase': 'IPermissionUseCase',
    'PurgeUseCase': 'IPurgeUseCase',
    
    # Core Repositories
    'UserRepository': 'IUserRepository',
    'SessionRepository': 'ISessionRepository',
    'RoleRepository': 'IRoleRepository',
    'PermissionRepository': 'IPermissionRepository',
    'CountryRepository': 'ICountryRepository',
    'CurrencyRepository': 'ICurrencyRepository',
    'LanguageRepository': 'ILanguageRepository',
    'TimezoneRepository': 'ITimezoneRepository',
    'RegionRepository': 'IRegionRepository',
    'CityRepository': 'ICityRepository',
    
    # Module: Posts
    'UserPostUseCase': 'IUserPostUseCase',
    'CommentUseCase': 'ICommentUseCase',
    'UserPostRepository': 'IUserPostRepository',
    'CommentRepository': 'ICommentRepository',
    'PostCommentRepository': 'IPostCommentRepository',
    
    # Module: Profiles
    'UserProfileUseCase': 'IUserProfileUseCase',
    'UserContactUseCase': 'IUserContactUseCase',
    'UserProfileRepository': 'IUserProfileRepository',
    'UserContactRepository': 'IUserContactRepository',
    
    # Module: Analytics
    'AnalyticsUseCase': 'IAnalyticsUseCase',
    'MetricRepository': 'IMetricRepository',
    
    # Module: Audit
    'AuditEventUseCase': 'IAuditEventUseCase',
    'AuditEventRepository': 'IAuditEventRepository',
    
    # Pkg interfaces
    'Module': 'IModule',
    'Bus': 'IBus',
    'EventHandler': 'IEventHandler',
}

def find_go_files(root_dir):
    """Find all Go files in the project"""
    go_files = []
    for root, dirs, files in os.walk(root_dir):
        # Skip vendor and hidden directories
        dirs[:] = [d for d in dirs if not d.startswith('.') and d != 'vendor']
        for file in files:
            if file.endswith('.go'):
                go_files.append(os.path.join(root, file))
    return go_files

def find_markdown_files(root_dir):
    """Find all markdown files"""
    md_files = []
    for root, dirs, files in os.walk(root_dir):
        dirs[:] = [d for d in dirs if not d.startswith('.')]
        for file in files:
            if file.endswith('.md'):
                md_files.append(os.path.join(root, file))
    return md_files

def rename_in_file(filepath, is_markdown=False):
    """Rename interfaces in a file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        changes = []
        
        # Apply renames
        for old_name, new_name in INTERFACE_RENAMES.items():
            # For Go files: match type definitions, variable declarations, function params, etc.
            if not is_markdown:
                # Interface definition: type OldName interface
                pattern1 = rf'\btype {old_name} interface\b'
                if re.search(pattern1, content):
                    content = re.sub(pattern1, f'type {new_name} interface', content)
                    changes.append(f"Interface definition: {old_name} -> {new_name}")
                
                # Variable/field declarations: var x OldName or x OldName
                pattern2 = rf'\b{old_name}\b'
                matches = len(re.findall(pattern2, content))
                if matches > 0:
                    content = re.sub(pattern2, new_name, content)
                    changes.append(f"References: {old_name} -> {new_name} ({matches} occurrences)")
            else:
                # For markdown: only in code blocks
                pattern = rf'\b{old_name}\b'
                matches = len(re.findall(pattern, content))
                if matches > 0:
                    content = re.sub(pattern, new_name, content)
                    changes.append(f"{old_name} -> {new_name} ({matches} occurrences)")
        
        # Write if changed
        if content != original:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True, changes
        
        return False, []
    
    except Exception as e:
        return False, [f"Error: {e}"]

def main():
    print("=" * 70)
    print("ПЕРЕЙМЕНУВАННЯ ІНТЕРФЕЙСІВ: Додавання префіксу I")
    print("=" * 70)
    print()
    
    # Process Go files
    print("Обробка Go файлів...")
    go_files = find_go_files('.')
    go_changed = 0
    go_changes_detail = {}
    
    for filepath in go_files:
        changed, changes = rename_in_file(filepath, is_markdown=False)
        if changed:
            go_changed += 1
            go_changes_detail[filepath] = changes
            print(f"  ✓ {filepath}")
    
    print(f"\nGo файлів змінено: {go_changed}/{len(go_files)}")
    
    # Process markdown files
    print("\nОбробка Markdown файлів...")
    md_files = find_markdown_files('.')
    md_changed = 0
    md_changes_detail = {}
    
    for filepath in md_files:
        changed, changes = rename_in_file(filepath, is_markdown=True)
        if changed:
            md_changed += 1
            md_changes_detail[filepath] = changes
            print(f"  ✓ {filepath}")
    
    print(f"\nMarkdown файлів змінено: {md_changed}/{len(md_files)}")
    
    # Summary
    print("\n" + "=" * 70)
    print("ПІДСУМОК")
    print("=" * 70)
    print(f"Усього інтерфейсів перейменовано: {len(INTERFACE_RENAMES)}")
    print(f"Go файлів оновлено: {go_changed}")
    print(f"Markdown файлів оновлено: {md_changed}")
    print(f"Загалом файлів змінено: {go_changed + md_changed}")
    
    # Show detailed changes for first few files
    if go_changes_detail:
        print("\nПриклади змін (перші 3 файли):")
        for i, (filepath, changes) in enumerate(list(go_changes_detail.items())[:3]):
            print(f"\n{filepath}:")
            for change in changes[:5]:  # Show first 5 changes per file
                print(f"  - {change}")
    
    print("\n✓ Перейменування завершено!")
    print("\nНаступні кроки:")
    print("1. Запустіть тести: make test")
    print("2. Перевірте що код компілюється: make build")
    print("3. Перегляньте зміни: git diff")
    print("4. Закомітьте: git add -A && git commit")

if __name__ == '__main__':
    main()

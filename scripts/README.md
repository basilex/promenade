**Navigation**: [Home](../README.md) > Development Scripts

---

# Scripts - Development Utilities

**Helper scripts** for development, deployment, and maintenance.

---

## Directory Structure

```
scripts/
 create-migration.sh               # Create new migration files
 clean-emojies.py                  # Remove emojis from documentation
 check-links.py                    # Validate markdown links (Python)
 check-links.sh                    # Validate markdown links (Bash)
 generate-accounting-postman.sh    # Generate Postman collection for accounting
```

---

## Scripts

### create-migration.sh

**Creates new migration files** with proper naming and namespace.

**Usage**:

```bash
# Via Makefile (recommended)
make migrate-new CONTEXT=identity NAME=add_email_verification

# Direct usage
./scripts/create-migration.sh identity add_email_verification
```

**What it does**:

1. Gets next migration number for namespace
2. Creates `migrations/{namespace}/{number}_{name}.up.sql`
3. Creates `migrations/{namespace}/{number}_{name}.down.sql`
4. Adds boilerplate SQL comments

**Example**:

```bash
$ make migrate-new CONTEXT=identity NAME=add_email_verification

Created:
- migrations/identity/000006_add_email_verification.up.sql
- migrations/identity/000006_add_email_verification.down.sql
```

---

### clean-emojies.py

**Removes emojis** from markdown documentation files.

**Usage**:

```bash
python scripts/clean-emojies.py
```

**What it does**:

- Scans all .md files in docs/
- Removes emoji characters
- Preserves markdown formatting
- Creates backups before modifying

---

### check-links.py / check-links.sh

**Validates markdown links** in documentation.

**Usage**:

```bash
# Python version (more features)
python scripts/check-links.py

# Bash version (faster)
bash scripts/check-links.sh
```

**What it does**:

- Checks all markdown links
- Reports broken internal links
- Validates file references
- Checks heading anchors

---

### generate-accounting-postman.sh

**Generates Postman collection fragment** for Accounting context.

**Usage**:

```bash
bash scripts/generate-accounting-postman.sh
```

**What it does**:

- Creates Postman requests for all accounting endpoints
- Generates /tmp/accounting-postman.json
- Includes all CRUD operations
- Ready for import into main collection

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [Documentation Index](../docs/INDEX.md) - Complete documentation
- [Migrations](../migrations/README.md) - Migration system

---

**Status**: Development utilities  
**Requirements**: Python 3.8+ (for clean-docs.py)  
**Maintainer**: Promenade Team

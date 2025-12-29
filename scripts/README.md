# Scripts - Development Utilities

**Helper scripts** for development, deployment, and maintenance.

---

## Directory Structure

```
scripts/
 create-migration.sh   # Create new migration file
 clean-docs.py         # Documentation cleanup utility
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

### clean-docs.py

**Python utility** for documentation maintenance.

**Usage**:

```bash
python scripts/clean-docs.py
```

**What it does**:
- Removes temporary files
- Validates markdown links
- Checks documentation structure

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [Documentation Index](../docs/INDEX.md) - Complete documentation
- [Migrations](../migrations/README.md) - Migration system

---

**Status**: Development utilities  
**Requirements**: Python 3.8+ (for clean-docs.py)  
**Maintainer**: Promenade Team

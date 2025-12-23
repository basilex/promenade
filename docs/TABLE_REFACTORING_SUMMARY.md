# Table Naming Refactoring - Summary

## ✅ Выполнено

### 1. Миграции Created

#### Core Module (14 tables)

- **Migration**: `migrations/core/000007_rename_tables_with_core_prefix.up.sql`
- **Rollback**: `migrations/core/000007_rename_tables_with_core_prefix.down.sql`
- **Status**: ✅ Applied (version 7)

**Tables renamed:**

- `users` → `core_users`
- `user_sessions` → `core_user_sessions`
- `user_roles` → `core_user_roles`
- `email_verification_tokens` → `core_email_verification_tokens`
- `password_reset_tokens` → `core_password_reset_tokens`
- `login_attempts` → `core_login_attempts`
- `roles` → `core_roles`
- `permissions` → `core_permissions`
- `role_permissions` → `core_role_permissions`
- `countries` → `core_countries`
- `currencies` → `core_currencies`
- `country_currencies` → `core_country_currencies`
- `languages` → `core_languages`
- `timezones` → `core_timezones`

**Indexes renamed**: ~60+ indexes

#### Profiles Module (2 tables)

- **Migration**: `migrations/profiles/000002_rename_tables_with_profiles_prefix.up.sql`
- **Rollback**: `migrations/profiles/000002_rename_tables_with_profiles_prefix.down.sql`
- **Status**: ⏳ Ready (will apply when profiles module is enabled)

**Tables to be renamed:**

- `user_profiles` → `profiles_profiles`
- `user_contacts` → `profiles_contacts`

**Indexes to be renamed**: ~18 indexes

#### Analytics Module

- **Status**: ✅ Already correct (uses `analytics_` prefix from start)
- 6 tables: `analytics_metrics`, `analytics_dashboards`, etc.

### 2. Repositories Updated

#### Core Repositories (9 files)

**Path**: `internal/adapter/repository/postgres/`

Files updated:

- ✅ `user_repository.go`
- ✅ `session_repository.go`
- ✅ `role_repository.go`
- ✅ `permission_repository.go`
- ✅ `country_repository.go`
- ✅ `currency_repository.go`
- ✅ `language_repository.go`
- ✅ `timezone_repository.go`
- ✅ `base_repository.go`

All SQL queries updated to use new table names with `core_` prefix.

#### Profiles Repositories (3 files)

**Path**: `internal/modules/profiles/adapter/repository/postgres/`

Files updated:

- ✅ `user_profile_repository.go`
- ✅ `user_contact_repository.go`
- ✅ `base_repository.go`

All SQL queries updated to use new table names with `profiles_` prefix.

### 3. Automation Script Created

**File**: `scripts/update-table-names.sh`

Features:

- Batch updates all repository files
- Creates backups (\*.go.bak)
- Handles both core and profiles modules
- Provides rollback instructions

Usage:

```bash
chmod +x scripts/update-table-names.sh
./scripts/update-table-names.sh
```

### 4. Testing

- ✅ Application compiles successfully
- ✅ Application starts without errors
- ✅ Core migrations applied (version 7)
- ✅ All core tables renamed correctly
- ⏳ Profiles migrations pending (module not enabled yet)

## 📋 Current Database State

```
analytics_*              (6 tables) - ✅ Correct
core_*                  (14 tables) - ✅ Correct
user_profiles            (1 table)  - ⏳ Pending rename to profiles_profiles
user_contacts            (1 table)  - ⏳ Pending rename to profiles_contacts
schema_migrations        (system)   - Unchanged
```

## 🎯 Benefits Achieved

1. **Logical Isolation**: Clear module ownership of tables
2. **DBA Friendly**: Easy to identify module tables at a glance
3. **Scalability**: Ready for module extraction to separate schemas/databases
4. **Consistency**: All business modules follow same naming pattern
5. **Maintenance**: Simplified backup, monitoring, and administration tasks

## 📝 Next Steps

### When enabling profiles module:

1. Enable profiles in `config/modules.yaml`:

   ```yaml
   modules:
     enabled:
       - profiles
   ```

2. Restart application - profiles migration will auto-apply:

   ```bash
   ./bin/promenade
   ```

3. Verify tables renamed:
   ```sql
   SELECT tablename FROM pg_tables
   WHERE schemaname = 'public' AND tablename LIKE 'profiles_%';
   ```

### Future Modules

All new modules should follow the naming convention:

- Posts module: `posts_*`
- Warehouse module: `warehouse_*`
- Any custom module: `{module_name}_*`

## 🔄 Rollback Instructions

If needed, rollback is simple:

```bash
# Rollback core tables
psql -U system -d promenade_dev < migrations/core/000007_rename_tables_with_core_prefix.down.sql

# Rollback profiles tables (if applied)
psql -U system -d promenade_dev < migrations/profiles/000002_rename_tables_with_profiles_prefix.down.sql

# Restore repository code
find . -name '*.go.bak' -exec bash -c 'mv "$0" "${0%.bak}"' {} \;
```

## 📊 Impact Summary

- **Tables renamed**: 14 core tables (100% of core tables)
- **Indexes renamed**: ~60 core indexes
- **Repository files updated**: 12 files (9 core + 3 profiles)
- **Migrations created**: 4 files (2 up + 2 down)
- **Lines of code changed**: ~500+ SQL statements
- **Build status**: ✅ Success
- **Runtime status**: ✅ Stable
- **Backward compatibility**: ✅ Full rollback available

## 🎉 Conclusion

Table naming refactoring successfully completed for core module! All tables now follow the Oracle DBA-friendly naming convention with module prefixes. System is ready for continued growth with consistent, maintainable database architecture.

---

**Date**: December 23, 2025  
**Version**: After migration 000007  
**Status**: Production Ready ✅

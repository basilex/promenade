# Quick Reference: Table Names

## Core Module (`core_*`)

### Authentication & Users

- `core_users` - User accounts
- `core_user_sessions` - Active user sessions
- `core_user_roles` - User-role assignments
- `core_email_verification_tokens` - Email verification tokens
- `core_password_reset_tokens` - Password reset tokens
- `core_login_attempts` - Failed login tracking

### RBAC (Access Control)

- `core_roles` - Role definitions
- `core_permissions` - Permission definitions
- `core_role_permissions` - Role-permission assignments

### Reference Data

- `core_countries` - Country directory
- `core_currencies` - Currency directory
- `core_country_currencies` - Country-currency relationships
- `core_languages` - Language directory
- `core_timezones` - Timezone directory

## Analytics Module (`analytics_*`)

- `analytics_metrics` - Metric events
- `analytics_metric_aggregates` - Pre-computed aggregates
- `analytics_reports` - Report definitions
- `analytics_report_schedules` - Scheduled reports
- `analytics_dashboards` - Dashboard definitions
- `analytics_dashboard_widgets` - Dashboard widgets

## Profiles Module (`profiles_*`)

- `profiles_profiles` - User profile data (was: user_profiles)
- `profiles_contacts` - User contact information (was: user_contacts)

## Posts Module (`posts_*`)

> _Not yet created - will use `posts_` prefix when implemented_

- `posts_posts` - User posts
- `posts_comments` - Post comments
- `posts_comment_likes` - Comment likes

## System Tables

- `schema_migrations` - Migration history (no prefix)

---

## SQL Query Examples

### Find all tables for a module

```sql
-- Core tables
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND tablename LIKE 'core_%';

-- Analytics tables
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND tablename LIKE 'analytics_%';

-- Profiles tables
SELECT tablename FROM pg_tables
WHERE schemaname = 'public' AND tablename LIKE 'profiles_%';
```

### Table sizes by module

```sql
SELECT
    CASE
        WHEN tablename LIKE 'core_%' THEN 'core'
        WHEN tablename LIKE 'analytics_%' THEN 'analytics'
        WHEN tablename LIKE 'profiles_%' THEN 'profiles'
        WHEN tablename LIKE 'posts_%' THEN 'posts'
        ELSE 'system'
    END AS module,
    COUNT(*) as table_count,
    pg_size_pretty(SUM(pg_total_relation_size(schemaname||'.'||tablename))) as total_size
FROM pg_tables
WHERE schemaname = 'public'
GROUP BY module
ORDER BY module;
```

### Backup by module

```bash
# Core module only
pg_dump -U system -d promenade_dev -t 'core_*' > backup_core.sql

# Analytics module only
pg_dump -U system -d promenade_dev -t 'analytics_*' > backup_analytics.sql

# Profiles module only
pg_dump -U system -d promenade_dev -t 'profiles_*' > backup_profiles.sql
```

### Grant permissions by module

```sql
-- Read-only access to analytics module
GRANT SELECT ON ALL TABLES IN SCHEMA public TO analytics_reader
WHERE tablename LIKE 'analytics_%';

-- Full access to profiles module
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO profiles_admin
WHERE tablename LIKE 'profiles_%';
```

---

**Last Updated**: December 23, 2025  
**Total Tables**: 23 (14 core + 6 analytics + 2 profiles + 1 system)

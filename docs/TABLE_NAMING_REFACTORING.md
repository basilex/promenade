# Table Naming Refactoring Plan

## Текущая ситуация

### ✅ Правильные имена (УЖЕ с префиксами):

- `analytics_*` (6 таблиц) - модуль analytics
  - analytics_metrics
  - analytics_metric_aggregates
  - analytics_reports
  - analytics_report_schedules
  - analytics_dashboards
  - analytics_dashboard_widgets

### ❌ Требуют рефакторинга:

#### Core модуль (должен быть префикс `core_`):

1. `users` → `core_users`
2. `user_sessions` → `core_user_sessions`
3. `user_roles` → `core_user_roles`
4. `email_verification_tokens` → `core_email_verification_tokens`
5. `password_reset_tokens` → `core_password_reset_tokens`
6. `login_attempts` → `core_login_attempts`
7. `roles` → `core_roles`
8. `permissions` → `core_permissions`
9. `role_permissions` → `core_role_permissions`
10. `countries` → `core_countries`
11. `country_currencies` → `core_country_currencies`
12. `currencies` → `core_currencies`
13. `languages` → `core_languages`
14. `timezones` → `core_timezones`

#### Profiles модуль (должен быть префикс `profiles_`):

1. `user_profiles` → `profiles_user_profiles` ИЛИ `profiles_profiles`
2. `user_contacts` → `profiles_contacts`

#### Posts модуль (должен быть префикс `posts_`):

Таблицы еще не созданы, но будут:

1. `user_posts` → `posts_posts`
2. `post_comments` → `posts_comments`
3. `comment_likes` → `posts_comment_likes`

## Стратегия рефакторинга

### Фаза 1: Core таблицы (14 таблиц)

- Создать миграцию `migrations/core/000007_rename_tables_with_core_prefix.up.sql`
- Переименовать все таблицы с добавлением префикса `core_`
- Обновить все индексы, constraints, foreign keys
- Обновить репозитории в `internal/adapter/repository/postgres/`

### Фаза 2: Profiles таблицы (2 таблицы)

- Создать миграцию `migrations/profiles/000002_rename_tables_with_profiles_prefix.up.sql`
- Переименовать таблицы с префиксом `profiles_`
- Обновить репозитории в `internal/modules/profiles/adapter/repository/postgres/`

### Фаза 3: Posts таблицы (будущие)

- При создании таблиц сразу использовать префикс `posts_`

## Преимущества

1. **Логическая изоляция**: Четко видно к какому модулю относится таблица
2. **Администрирование**: Упрощается работа DBA с бэкапами, мониторингом, партиционированием
3. **Масштабирование**: В будущем легко перенести модуль в отдельную схему или БД
4. **Безопасность**: Легче настроить права доступа на уровне модулей
5. **Документация**: Самодокументирующиеся имена таблиц

## Порядок выполнения

1. ✅ Создать план рефакторинга (этот документ)
2. ⏳ Создать миграции для core
3. ⏳ Обновить core репозитории
4. ⏳ Создать миграции для profiles
5. ⏳ Обновить profiles репозитории
6. ⏳ Применить миграции на dev окружении
7. ⏳ Запустить тесты
8. ⏳ Обновить документацию

## Важные замечания

- Schema migrations таблица остается без префикса (системная таблица)
- Foreign keys нужно пересоздать с новыми именами таблиц
- Индексы автоматически переименуются, но лучше явно задать имена
- Down миграции должны вернуть старые имена для rollback

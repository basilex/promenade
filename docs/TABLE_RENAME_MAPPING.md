# Table Renaming Mapping

## Core Module Tables

| Old Name                  | New Name                       | File Pattern                              |
| ------------------------- | ------------------------------ | ----------------------------------------- |
| users                     | core_users                     | user_repository.go, session_repository.go |
| user_sessions             | core_user_sessions             | session_repository.go                     |
| user_roles                | core_user_roles                | role_repository.go                        |
| email_verification_tokens | core_email_verification_tokens | email_verification_repository.go          |
| password_reset_tokens     | core_password_reset_tokens     | password_reset_repository.go              |
| login_attempts            | core_login_attempts            | login_attempt_repository.go               |
| roles                     | core_roles                     | role_repository.go                        |
| permissions               | core_permissions               | permission_repository.go                  |
| role_permissions          | core_role_permissions          | role_repository.go                        |
| countries                 | core_countries                 | country_repository.go                     |
| country_currencies        | core_country_currencies        | country_repository.go                     |
| currencies                | core_currencies                | currency_repository.go                    |
| languages                 | core_languages                 | language_repository.go                    |
| timezones                 | core_timezones                 | timezone_repository.go                    |

## Profiles Module Tables

| Old Name      | New Name          | File Pattern          |
| ------------- | ----------------- | --------------------- |
| user_profiles | profiles_profiles | profile_repository.go |
| user_contacts | profiles_contacts | contact_repository.go |

## Search & Replace Patterns

### Core Repositories

```bash
# Users
FROM users → FROM core_users
INTO users → INTO core_users
UPDATE users → UPDATE core_users
DELETE FROM users → DELETE FROM core_users
JOIN users → JOIN core_users

# Sessions
FROM user_sessions → FROM core_user_sessions
INTO user_sessions → INTO core_user_sessions
UPDATE user_sessions → UPDATE core_user_sessions
DELETE FROM user_sessions → DELETE FROM core_user_sessions

# And so on for each table...
```

### Profiles Repositories

```bash
# Profiles
FROM user_profiles → FROM profiles_profiles
INTO user_profiles → INTO profiles_profiles
UPDATE user_profiles → UPDATE profiles_profiles

# Contacts
FROM user_contacts → FROM profiles_contacts
INTO user_contacts → INTO profiles_contacts
UPDATE user_contacts → UPDATE profiles_contacts
```

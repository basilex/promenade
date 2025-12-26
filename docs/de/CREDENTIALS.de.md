[ English](../CREDENTIALS.md) | [ Українська](../uk/CREDENTIALS.uk.md) |  **Deutsch** | [ Português](CREDENTIALS.pt.md) | [ Español](CREDENTIALS.es.md)

---

# Entwicklungs-Zugangsdaten

Schnellreferenz für Standard-Benutzer und Zugangsdaten mit RBAC-Rollen.

## Standard-Benutzer

### 1. Systemadministrator (Hauptadmin - Oracle-Stil)

```
Email:    system@promenade.com
Password: passw0rd
Role:     admin
Access:   *:* (vollständiger Systemzugriff)
```

**Verwenden für:**

- Systeminitialisierung und Bootstrap
- RBAC-Verwaltung (Rollen, Berechtigungen)
- Benutzerverwaltung (Sperren, Aussetzen, Rollen zuweisen)
- Alle administrativen Operationen

### 2. Administrator

```
Email:    admin@promenade.com
Password: passw0rd
Role:     admin
Access:   users:*, posts:*, comments:*, profiles:*, roles:read|list|assign
```

**Verwenden für:**

- Benutzerverwaltung (Erstellen, Aktualisieren, Löschen, Sperren)
- Inhaltsverwaltung (Beiträge, Kommentare)
- Rollenzuweisung an Benutzer
- Testen von Admin-Berechtigungen

### 4. Moderator

```
Email:    moderator@promenade.com
Password: passw0rd
Role:     moderator
Access:   posts:read|update|delete|list, comments:*, profiles:read|list
```

**Verwenden für:**

- Inhaltsmoderation (Beiträge, Kommentare)
- Kommentarverwaltung (Genehmigen, Löschen)
- Testen von Moderations-Workflows
- Eingeschränkte Benutzersichtbarkeit (nur Lesen)

### 5. Regulärer Benutzer

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     user
Access:   posts:create|read, comments:create|read, profiles:create|read
```

**Verwenden für:**

- Testen regulärer Benutzer-Workflows
- Eigene Inhaltserstellung (Beiträge, Kommentare)
- Profilverwaltung
- Grundlegende Benutzeroperationen

### Gastzugriff (nicht authentifiziert)

Kein Login erforderlich:

```
Role:     guest (implizit)
Access:   posts:read, comments:read, profiles:read
```

**Verwenden für:**

- Durchsuchen öffentlicher Inhalte
- Testen nicht authentifizierten Zugriffs
- Nur-Lese-Operationen

## Schnelle Login-Beispiele

```bash
# Login als Systemadministrator (admin mit vollem Zugriff)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Login als Administrator (admin)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@promenade.com","password":"passw0rd"}' | jq

# Login als Moderator (moderator)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@promenade.com","password":"passw0rd"}' | jq

# Login als regulärer Benutzer (user)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Token zur Wiederverwendung speichern (Systemadmin)
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Token in Anfragen verwenden
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Testen von RBAC-Berechtigungen

```bash
# Test Systemadmin-Zugriff (sollte funktionieren - voller Zugriff)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq

# Test Admin-Zugriff (sollte funktionieren - hat users:*)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Test Moderator-Zugriff (sollte fehlschlagen - keine users:list Berechtigung)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $MODERATOR_TOKEN" | jq

# Test Benutzer-Zugriff (sollte fehlschlagen - keine Admin-Berechtigungen)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Datenbankzugriff

```bash
# Verbindung zur Entwicklungsdatenbank
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# Alle Benutzer mit ihren Rollen anzeigen
SELECT
    u.email,
    u.name,
    r.name as role,
    r.description
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.email;

# Rollen-Berechtigungen anzeigen
SELECT
    r.name as role,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.resource, p.action;
```

## [!] Sicherheitswarnung

**Diese Zugangsdaten sind NUR FÜR DIE ENTWICKLUNG!**

Vor der Bereitstellung in der Produktion:

1. Alle Standard-Passwörter ändern
2. Standard-Admin-Konten entfernen oder deaktivieren
3. Umgebungsspezifische Zugangsdaten verwenden
4. Ordnungsgemäßes Secret-Management aktivieren
5. Ordnungsgemäße Authentifizierungsanbieter einrichten

## Benötigen Sie Hilfe?

- [Autorisierungsleitfaden](docs/AUTHORIZATION.md) - RBAC-Systemdokumentation
- [Testleitfaden](docs/TESTING_GUIDE.md) - Testen mit Authentifizierung
- [API-Dokumentation](README.md#api-examples) - Vollständige API-Referenz

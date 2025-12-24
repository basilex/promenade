# Promenade Dokumentationsindex

[🇬🇧 English](../INDEX.md) | [🇺🇦 Українська](../uk/INDEX.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/INDEX.pt.md) | [🇪🇸 Español](../es/INDEX.es.md)

Dieses Verzeichnis enthält umfassende Dokumentation zur Promenade-Anwendungsarchitektur, Entwicklungs-Workflows und Best Practices.

> 🌍 **Neu!** Die Dokumentation ist jetzt in mehreren Sprachen verfügbar. Siehe [TRANSLATIONS.md](../TRANSLATIONS.md) für Übersetzungsstatus und Beitragsrichtlinien.

---

## Hier Beginnen

### Neu bei Promenade?

1. **[ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md)** - Visuelle Architekturdiagramme und Komponentenübersicht
2. **[ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md)** - Schnellreferenz für Entwickler
3. **[README.md](../../README.md)** - Haupt-README des Projekts

### Architektur-Review

- **[ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md)** - Vollständiges Architektur-Compliance-Audit (Core vs Module)

---

## Kernkonzepte

### Architektur & Design

| Dokument                                                | Beschreibung                                     | Wann Lesen                          |
| ------------------------------------------------------- | ------------------------------------------------ | ----------------------------------- |
| [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) | Vollständige visuelle Architektur mit Diagrammen | Systemstruktur verstehen            |
| [ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md)       | Architektur-Compliance-Bericht                   | Design-Prinzipien verifizieren      |
| [ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md) | Schnellreferenz für gängige Muster               | Tägliche Entwicklung                |
| [../../internal/CORE.md](../../internal/CORE.md)        | Dokumentation der Core-Komponenten               | Core-Verantwortlichkeiten verstehen |

### Module

| Dokument                                                             | Beschreibung                                       | Wann Lesen                       |
| -------------------------------------------------------------------- | -------------------------------------------------- | -------------------------------- |
| [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md)                    | Vollständiger Leitfaden zur Erstellung von Modulen | Neue Module erstellen            |
| [MODULE_INDEPENDENCE.md](../MODULE_INDEPENDENCE.md)                  | Prinzipien der Modul-Unabhängigkeit                | Modulg renzen verstehen          |
| [MODULE_CONFIG_ARCHITECTURE.md](../MODULE_CONFIG_ARCHITECTURE.md)    | Konfigurationsmanagement für Module                | Modul-Konfigurationen einrichten |
| [../../internal/modules/README.md](../../internal/modules/README.md) | Modul-Verzeichnisstruktur                          | Schneller Modul-Überblick        |

---

## Technische Leitfäden

### Datenbank & Persistenz

| Dokument                                | Beschreibung                       | Wann Lesen                    |
| --------------------------------------- | ---------------------------------- | ----------------------------- |
| [UUID_V7_GUIDE.md](../UUID_V7_GUIDE.md) | Verwendung zeitgeordneter UUIDs    | Arbeiten mit Primärschlüsseln |
| [SOFT_DELETE.md](../SOFT_DELETE.md)     | Soft-Delete-Muster und Fallstricke | Soft-Delete implementieren    |

### Infrastruktur

| Dokument                                          | Beschreibung                             | Wann Lesen                        |
| ------------------------------------------------- | ---------------------------------------- | --------------------------------- |
| [PURGE_ARCHITECTURE.md](../PURGE_ARCHITECTURE.md) | Design des automatisierten Purge-Systems | Retention-Policies implementieren |
| [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md)   | Testen mit Redis Event Bus               | Event-gesteuerte Features testen  |
| [LOGGING.md](../LOGGING.md)                       | Strukturiertes Logging mit Kontext       | Logging zum Code hinzufügen       |

### Sicherheit & Authentifizierung

| Dokument                                              | Beschreibung                                         | Wann Lesen                         |
| ----------------------------------------------------- | ---------------------------------------------------- | ---------------------------------- |
| [AUTH_SCHEMA.md](../AUTH_SCHEMA.md)                   | Authentifizierungssystem (Registrierung, Login, JWT) | Auth-Flow verstehen                |
| [AUTHORIZATION.md](../AUTHORIZATION.md)               | RBAC-Berechtigungssystem                             | Autorisierung implementieren       |
| [CREDENTIALS.md](../CREDENTIALS.md)                   | Standardbenutzer und Rollen für dev/test             | Tests mit vordefinierten Benutzern |
| [LICENSE_ARCHITECTURE.md](../LICENSE_ARCHITECTURE.md) | Modul-Lizenzsystem mit HMAC-SHA256                   | Kommerzielle Module implementieren |

---

## Testen

| Dokument                                                      | Beschreibung                              | Wann Lesen                       |
| ------------------------------------------------------------- | ----------------------------------------- | -------------------------------- |
| [TESTING_GUIDE.md](../TESTING_GUIDE.md)                       | Vollständige Test-Strategie               | Tests schreiben                  |
| [TESTING_INFRASTRUCTURE.md](../TESTING_INFRASTRUCTURE.md)     | Test-Infrastruktur-Setup                  | Testumgebung einrichten          |
| [MOCK_GENERATION_STANDARD.md](../MOCK_GENERATION_STANDARD.md) | Einheitlicher Mock-Generierungsansatz     | Mit Repository-Mocks arbeiten    |
| [MOCK_STANDARDIZATION.md](../MOCK_STANDARDIZATION.md)         | Zusammenfassung der Mock-Standardisierung | Mock-Vereinheitlichung verstehen |
| [../../test/README.md](../../test/README.md)                  | Test-Verzeichnisstruktur                  | Test-Organisation verstehen      |

---

## Entwicklungs-Workflows

### Build & Deploy

| Dokument                                                | Beschreibung                  | Wann Lesen                 |
| ------------------------------------------------------- | ----------------------------- | -------------------------- |
| [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) | Makefile-System-Dokumentation | Make-Befehle verwenden     |
| [../../docker/README.md](../../docker/README.md)        | Docker-Setup und Deployment   | Anwendung containerisieren |

### Validierung & Qualität

| Dokument                          | Beschreibung             | Wann Lesen                    |
| --------------------------------- | ------------------------ | ----------------------------- |
| [VALIDATION.md](../VALIDATION.md) | Input-Validierungsmuster | Validierungsregeln hinzufügen |

---

## API-Dokumentation

### V1 API

- [v1/v1_docs.go](../v1/v1_docs.go) - V1 API-Dokumentation
- [v1/v1_swagger.yaml](../v1/v1_swagger.yaml) - V1 Swagger-Spezifikation (YAML)
- [v1/v1_swagger.json](../v1/v1_swagger.json) - V1 Swagger-Spezifikation (JSON)

### V2 API

- [v2/v2_docs.go](../v2/v2_docs.go) - V2 API-Dokumentation
- [v2/v2_swagger.yaml](../v2/v2_swagger.yaml) - V2 Swagger-Spezifikation (YAML)
- [v2/v2_swagger.json](../v2/v2_swagger.json) - V2 Swagger-Spezifikation (JSON)

---

## Nach Thema

### Core-Architektur

- [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) - Visueller Überblick
- [ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md) - Compliance-Review
- [ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md) - Schnellreferenz
- [../../internal/CORE.md](../../internal/CORE.md) - Core-Komponenten

### Modul-System

- [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md) - Entwicklungsleitfaden
- [MODULE_INDEPENDENCE.md](../MODULE_INDEPENDENCE.md) - Unabhängigkeitsprinzipien
- [MODULE_CONFIG_ARCHITECTURE.md](../MODULE_CONFIG_ARCHITECTURE.md) - Konfiguration
- [../../internal/modules/README.md](../../internal/modules/README.md) - Modul-Index

### Datenmanagement

- [UUID_V7_GUIDE.md](../UUID_V7_GUIDE.md) - Primärschlüssel
- [SOFT_DELETE.md](../SOFT_DELETE.md) - Soft-Delete-Muster
- [PURGE_ARCHITECTURE.md](../PURGE_ARCHITECTURE.md) - Automatisiertes Purge

### Sicherheit & Auth

- [AUTH_SCHEMA.md](../AUTH_SCHEMA.md) - Authentifizierung
- [AUTHORIZATION.md](../AUTHORIZATION.md) - RBAC-System
- [CREDENTIALS.md](../CREDENTIALS.md) - Anmeldedaten-Handhabung

### Testen & Qualität

- [TESTING_GUIDE.md](../TESTING_GUIDE.md) - Test-Strategie
- [TESTING_INFRASTRUCTURE.md](../TESTING_INFRASTRUCTURE.md) - Test-Setup
- [VALIDATION.md](../VALIDATION.md) - Input-Validierung
- [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Event-Bus-Tests

### Infrastruktur

- [LOGGING.md](../LOGGING.md) - Strukturiertes Logging
- [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) - Build-System
- [../../docker/README.md](../../docker/README.md) - Docker-Setup

---

## Lernpfade

### Pfad 1: System Verstehen (Neuer Entwickler)

1. [README.md](../../README.md) - Projekt-Überblick
2. [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) - Systemarchitektur
3. [ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md) - Gängige Muster
4. [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md) - Features erstellen
5. [TESTING_GUIDE.md](../TESTING_GUIDE.md) - Code testen

### Pfad 2: Neues Modul Erstellen

1. [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md) - Modul-Erstellungsleitfaden
2. [MODULE_INDEPENDENCE.md](../MODULE_INDEPENDENCE.md) - Design-Prinzipien
3. [MODULE_CONFIG_ARCHITECTURE.md](../MODULE_CONFIG_ARCHITECTURE.md) - Konfiguration
4. [UUID_V7_GUIDE.md](../UUID_V7_GUIDE.md) - Primärschlüssel
5. [SOFT_DELETE.md](../SOFT_DELETE.md) - Falls Soft-Delete verwendet wird
6. [PURGE_ARCHITECTURE.md](../PURGE_ARCHITECTURE.md) - Falls Purge implementiert wird

### Pfad 3: Architektur-Review (Technical Lead)

1. [ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md) - Analyse des aktuellen Zustands
2. [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) - Visuelle Diagramme
3. [../../internal/CORE.md](../../internal/CORE.md) - Core-Grenzen
4. [MODULE_INDEPENDENCE.md](../MODULE_INDEPENDENCE.md) - Modul-Isolation

### Pfad 4: Sicherheits-Implementierung

1. [AUTH_SCHEMA.md](../AUTH_SCHEMA.md) - Authentifizierungs-Flows
2. [AUTHORIZATION.md](../AUTHORIZATION.md) - RBAC-Berechtigungen
3. [CREDENTIALS.md](../CREDENTIALS.md) - Anmeldedaten-Sicherheit

### Pfad 5: Testen & Qualität

1. [TESTING_GUIDE.md](../TESTING_GUIDE.md) - Test-Strategie
2. [TESTING_INFRASTRUCTURE.md](../TESTING_INFRASTRUCTURE.md) - Test-Setup
3. [VALIDATION.md](../VALIDATION.md) - Input-Validierung
4. [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Event-Tests

---

## Schnellsuche

### Wie kann ich...

**...ein neues Modul erstellen?**
→ [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md)

**...Konfiguration zu meinem Modul hinzufügen?**
→ [MODULE_CONFIG_ARCHITECTURE.md](../MODULE_CONFIG_ARCHITECTURE.md)

**...Soft-Delete implementieren?**
→ [SOFT_DELETE.md](../SOFT_DELETE.md)

**...Retention-Policies hinzufügen?**
→ [PURGE_ARCHITECTURE.md](../PURGE_ARCHITECTURE.md)

**...UUIDs korrekt verwenden?**
→ [UUID_V7_GUIDE.md](../UUID_V7_GUIDE.md)

**...Authentifizierung implementieren?**
→ [AUTH_SCHEMA.md](../AUTH_SCHEMA.md)

**...RBAC-Berechtigungen hinzufügen?**
→ [AUTHORIZATION.md](../AUTHORIZATION.md)

**...Tests schreiben?**
→ [TESTING_GUIDE.md](../TESTING_GUIDE.md)

**...Logging hinzufügen?**
→ [LOGGING.md](../LOGGING.md)

**...Input validieren?**
→ [VALIDATION.md](../VALIDATION.md)

**...die Architektur verstehen?**
→ [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md)

**...prüfen, ob mein Code den Prinzipien folgt?**
→ [ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md)

**...eine Schnellreferenz erhalten?**
→ [ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md)

---

## Letzte Updates

### 22. Dezember 2025

- Umfassende Architektur-Dokumentation erstellt:
- [ARCHITECTURE_AUDIT.md](../ARCHITECTURE_AUDIT.md) - Vollständiger Compliance-Review
- [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) - Visuelle Diagramme
- [ARCHITECTURE_QUICKREF.md](../ARCHITECTURE_QUICKREF.md) - Schnellreferenz
- Purge-System-Dokumentation aktualisiert:
- [PURGE_ARCHITECTURE.md](../PURGE_ARCHITECTURE.md) - Registry-basierter Ansatz
- Modul-Unabhängigkeit verifiziert (15.000+ Zeilen aus Core entfernt)

---

## 🤝 Mitwirken

Beim Hinzufügen neuer Dokumentation:

1. **Den richtigen Typ wählen:**

   - `ARCHITECTURE_*.md` - Architektur und Design-Muster
   - `MODULE_*.md` - Modul-System-Dokumentation
   - `*_GUIDE.md` - Anleitungen und Tutorials
   - `*_SCHEMA.md` - Datenschemata und Strukturen
   - `README.md` - Verzeichnis-Übersichten

2. **Diesen Index aktualisieren:**

   - Zum relevanten Abschnitt hinzufügen
   - "Letzte Updates" aktualisieren
   - Zu "Schnellsuche" hinzufügen, falls zutreffend

3. **Querverweise:**

   - Auf verwandte Dokumente verlinken
   - Verwandte Dokumente mit Rückverweisen aktualisieren

4. **Aktuell halten:**
   - Bei Architekturänderungen aktualisieren
   - Veraltete Dokumente mit Präfix `DEPRECATED_` archivieren

---

## 📞 Support

- **Fragen zur Architektur?** → Lesen Sie [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md)
- **Fragen zu Modulen?** → Lesen Sie [MODULE_DEVELOPMENT.md](../MODULE_DEVELOPMENT.md)
- **Fragen zum Testen?** → Lesen Sie [TESTING_GUIDE.md](../TESTING_GUIDE.md)
- **Andere Fragen?** → Überprüfen Sie diesen Index oder das Haupt-[README.md](../../README.md)

---

**Dokumentationsversion:** 2.0  
**Zuletzt aktualisiert:** 22. Dezember 2025  
**Gewartet von:** Promenade-Entwicklungsteam

---
title: "Leitfäden"
description: "Vollständige Sammlung von Leitfäden und Tutorials für Promenade"
aliases:
  - /de/docs/guides/
---

## Leitfäden & Tutorials

Umfassende Leitfäden zum Lernen und Arbeiten mit Promenade.

---

## Erste Schritte

<div class="docs-grid">

### [Schnellstart](/promenade/de/docs/getting-started)

In 5 Minuten einsatzbereit. Abhängigkeiten installieren, Services starten und Installation verifizieren.

### [Architekturübersicht](/promenade/de/docs/architecture)

Verstehen Sie die Kernprinzipien: Clean Architecture, Core vs Module und Systemdesign.

### [Datenbank-Setup](/promenade/de/docs/database-schema)

Lernen Sie PostgreSQL-Konfiguration, UUID v7, Migrationen und Schema-Design.

</div>

---

## Entwicklungs-Leitfäden

<div class="docs-grid">

### [Modul-Entwicklung](/promenade/de/docs/module-development)

Vollständiger Leitfaden zur Erstellung benutzerdefinierter Business-Module mit Beispielen.

### [Test-Strategie](/promenade/de/docs/getting-started#testing)

Wie man Unit-, Integrations- und Smoke-Tests schreibt. Lernen Sie das manuelle Mock-Muster.

### [API-Integration](/api/v1/docs/swagger/index.html)

Erkunden Sie REST-API-Endpunkte, Authentifizierung und Integrationsmuster via Swagger UI.

</div>

---

## Architektur-Leitfäden

<div class="docs-grid">

### [Clean Architecture](/promenade/de/features/clean-architecture)

Tiefgang in Domain-, Use Case-, Adapter- und Infrastructure-Schichten.

### [Modulsystem](/promenade/de/features/module-system)

Wie das Modulsystem funktioniert: Registrierung, Lebenszyklus und Kommunikationsmuster.

### [Event-Driven Architecture](/promenade/de/features/event-bus)

Event-Bus für asynchrone Kommunikation zwischen Modulen (Memory/Redis-Adapter).

</div>

---

## Datenbank-Leitfäden

<div class="docs-grid">

### [Datenbankschema](/promenade/de/docs/database-schema)

Vollständige Schema-Dokumentation mit ER-Diagrammen für alle Tabellen.

### [UUID v7 Leitfaden](/promenade/de/features/database#uuid-v7)

Warum UUID v7 2x schneller ist als UUID v4 und wie man es verwendet.

### [Soft Delete Muster](/promenade/de/features/database#soft-delete)

Soft Delete für Benutzerinhalte mit automatischen Löschrichtlinien implementieren.

</div>

---

## Fortgeschrittene Themen

<div class="docs-grid">

### [RBAC & Berechtigungen](/promenade/de/features/authentication)

Rollenbasierte Zugriffskontrolle mit Wildcard-Berechtigungen und 4 Systemrollen.

### [Referenzdaten](/promenade/de/docs/database-schema#reference-data)

145 Länder, 124 Währungen, 30 Regionen, 17 Städte, 40+ Zahlungsmethoden.

### [Namespace-Migrationen](/promenade/de/docs/database-schema#migrations)

Unabhängige Migrationshistorie für jedes Modul ohne Konflikte.

</div>

---

## Externe Ressourcen

- [GitHub Repository](https://github.com/basilex/promenade) - Quellcode und Issues
- [Beispiel-Module](https://github.com/basilex/promenade/tree/dev/internal/modules) - Posts, Profiles, Analytics, Billing
- [API-Dokumentation](/api/v1/docs/swagger/index.html) - Interaktive Swagger UI

---

## Hilfe Benötigt?

- Siehe [Dokumentationsindex](/promenade/de/docs/) für alle verfügbaren Leitfäden
- Besuchen Sie [GitHub Discussions](https://github.com/basilex/promenade/discussions) für Fragen
- Melden Sie Probleme auf [GitHub Issues](https://github.com/basilex/promenade/issues)

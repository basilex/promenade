---
title: "Production-Ready REST API Framework"
description: "Gebaut mit Clean Architecture, modularem Plugin-System und namespace-basierten Datenbankmigrationen. Geschrieben in Go."
date: 2025-12-25
features:
  - icon: "🏗️"
    title: "Clean Architecture"
    description: "Strikte Schichtenarchitektur mit Domain-, Use Case-, Adapter- und Infrastructure-Schichten. Dependency Rule durchgesetzt."

  - icon: "🧩"
    title: "Modulares System"
    description: "Unabhängige Business-Module mit eigenen Entities, Repositories und Use Cases. Dynamisches Aktivieren/Deaktivieren von Modulen."

  - icon: "🗄️"
    title: "PostgreSQL + UUID v7"
    description: "Zeitlich sortierte UUIDs für 2x schnellere Inserts. Reines SQL mit sqlx - kein ORM-Overhead. BaseRepository-Pattern."

  - icon: "🔐"
    title: "Auth & RBAC"
    description: "JWT-Authentifizierung, rollenbasierte Zugriffskontrolle mit Wildcard-Berechtigungen. Session-Management inklusive."

  - icon: "📡"
    title: "Event-Driven"
    description: "Duale Event-Bus-Adapter (Memory/Redis). Asynchrone Kommunikation zwischen Modulen. Production-ready."

  - icon: "🔄"
    title: "Namespace-Migrationen"
    description: "Jedes Modul hat eine unabhängige Migrationshistorie. Echte Modulautonomie ohne Schema-Konflikte."

  - icon: "✅"
    title: "400+ Tests"
    description: "Umfassende Test-Suite mit Unit-, Integration- und Smoke-Tests. Manual-Mocks-Pattern für Testbarkeit."

  - icon: "📚"
    title: "20+ Dokumente"
    description: "Umfangreiche Dokumentation in EN/UK/DE. Architektur-Guides, Tutorials und API-Referenzen."

  - icon: "🚀"
    title: "Production-Ready"
    description: "Wird in realen Anwendungen verwendet. Kommerzielle Module verfügbar (billing, audit, warehouse)."
---

## Schnellstart

```bash
# Repository klonen
git clone https://github.com/basilex/promenade.git
cd promenade

# PostgreSQL + Redis starten
make docker-up

# Migrationen ausführen
make migrate

# Anwendung starten
make dev
```

Server startet auf **http://localhost:8081**

## Hauptfunktionen

### Modulare Architektur

Promenade verwendet ein **Plugin-System**, bei dem jedes Modul ein vollständiger vertikaler Slice ist:

- Eigene Domain-Entities
- Eigene Repositories
- Eigene Use Cases
- Eigene HTTP-Handler
- Eigene Datenbankmigrationen

### Verfügbare Module

**Kostenlos:**

- **Posts** - Benutzergenerierte Inhalte (Posts, Kommentare, Likes)
- **Profiles** - Benutzerprofile und Kontakte
- **Analytics** - Metriken, Berichte, Dashboards

**Kommerziell:**

- **Billing** - Abonnementverwaltung, Rechnungen, Zahlungen
- **Audit** - Unveränderliche Audit-Logs mit kryptografischen Signaturen
- **Warehouse** - Bestandsverwaltung (demnächst)

## Lernressourcen

- [Architekturübersicht](/docs/ARCHITECTURE_OVERVIEW) - Visuelle Diagramme
- [Modulentwicklung](/docs/MODULE_DEVELOPMENT) - Neue Module erstellen
- [Test-Guide](/docs/TESTING_GUIDE) - Test-Strategien
- [UUID v7 Guide](/docs/UUID_V7_GUIDE) - Zeitlich sortierte Identifikatoren

---

**Gebaut mit Clean Architecture und Go**

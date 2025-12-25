---
title: "Production-Ready REST API Framework"
description: "Enterprise-Backend-Framework auf Basis von Clean Architecture Prinzipien. Revolutionäres modulares Plugin-System, bei dem Geschäftslogik in unabhängigen, lizenzierbaren Modulen lebt. Keine gemeinsamen Domain-Entitäten, keine enge Kopplung - echte architektonische Autonomie."
date: 2025-12-25
features:
  - icon: "🏗️"
    title: "Clean Architecture"
    description: "Strikte geschichtete Architektur mit Domain-, Use Case-, Adapter- und Infrastructure-Schichten. Abhängigkeitsregel durchgesetzt."

  - icon: "🧩"
    title: "Modulares System"
    description: "Unabhängige Business-Module mit eigenen Entitäten, Repositories und Use Cases. Dynamisches Aktivieren/Deaktivieren von Modulen."

  - icon: "🗄️"
    title: "PostgreSQL + UUID v7"
    description: "Zeitbasierte UUIDs für 2x schnellere Einfügungen. Reines SQL mit sqlx - kein ORM. BaseRepository-Pattern."

  - icon: "🔐"
    title: "Auth & RBAC"
    description: "JWT-Authentifizierung, rollenbasierte Zugriffskontrolle mit Wildcard-Berechtigungen. Session-Management inklusive."

  - icon: "📡"
    title: "Event-Driven"
    description: "Duale Event-Bus-Adapter (Memory/Redis). Asynchrone Kommunikation zwischen Modulen. Production-ready."

  - icon: "🔄"
    title: "Namespace-Migrationen"
    description: "Jedes Modul hat unabhängige Migrationshistorie. Echte Modulautonomie ohne Schema-Konflikte."

  - icon: "✅"
    title: "400+ Tests"
    description: "Umfassende Test-Suite mit Unit-, Integrations- und Smoke-Tests. Manuelle Mocks-Pattern für Testbarkeit."

  - icon: "📚"
    title: "20+ Dokumente"
    description: "Umfangreiche Dokumentation in EN/UK/DE. Architektur-Guides, Tutorials und API-Referenzen."

  - icon: "🚀"
    title: "Production-Ready"
    description: "In realen Anwendungen eingesetzt. Kommerzielle Module verfügbar (billing, audit, warehouse)."
---

## Schnellstart

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
make docker-up
make migrate
make dev
```

Server startet auf **http://localhost:8081**

---

## Promenade-Philosophie

**Promenade ist nicht nur ein weiteres Go-Framework** - es ist ein komplettes Umdenken, wie Backend-Anwendungen strukturiert werden sollten.

🎯 **Core als Orchestrator, Module als Worker** - Core bietet Infrastruktur ohne Geschäftslogik

🔌 **Echte Unabhängigkeit** - Module importieren nie aus `internal/domain`

📊 **Namespace-Migrationen** - Jedes Modul hat eigene Historie ohne Konflikte

⚡ **Performance** - UUID v7 sorgt für 2x schnellere Einfügungen

🔐 **Sicherheit** - JWT, RBAC, Audit-Logs mit HMAC-Signaturen

---

## Warum Promenade Wählen?

### Für Startups

✅ Schnelle Markteinführung ✅ Kostenlose Module für MVP ✅ Einzelne Bereitstellung

### Für Enterprise

✅ Wartbarkeit ✅ Auditierbarkeit (SOC 2, GDPR) ✅ Skalierbarkeit

### Für Teams

✅ Keine Migrationskonflikte ✅ Klare Modulgrenzen ✅ Mehrsprachige Docs

### Für Entwickler

✅ Best Practices ✅ Business-Fokus ✅ Fertige Muster

---

## Dokumentation

📖 [Architektur-Überblick](/promenade/de/docs/architecture/) | [DB-Schema](/promenade/de/docs/database-schema/) | [Modul-Entwicklung](/promenade/de/docs/module-development/)

🌐 EN/UK/DE | 📝 [Swagger API](/api/v1/docs/swagger/index.html)

---

**Erstellt mit ❤️ unter Verwendung von Clean Architecture und Go**

[Erste Schritte](/promenade/de/docs/getting-started/) | [Features](/promenade/de/features/) | [GitHub](https://github.com/basilex/promenade)

---
title: "Dokumentation"
description: "Vollständige technische Dokumentation für Promenade"
---

## Technische Dokumentation

Umfassende Anleitungen, Diagramme und Referenzen für Promenade-Architektur und -Entwicklung.

---

## Architektur

<div class="docs-grid">

### [Architektur-Übersicht](/docs/architecture)

Vollständige Systemarchitektur mit Diagrammen, Schichten und Kommunikationsmustern.

- Clean Architecture Schichten
- Modulares System-Design
- Ereignisgesteuerte Kommunikation
- Deployment-Architektur

### [Datenbankschema](/docs/database-schema)

Detailliertes Datenbankschema mit ER-Diagrammen und Beziehungen.

- Kerntabellen (Benutzer, Sitzungen, RBAC)
- Referenzdaten (Länder, Währungen, Regionen, Städte)
- Modultabellen (Posts, Profile, Analytics, Billing)
- Indizes und Performance

### [Modulentwicklung](/docs/module-development)

Schritt-für-Schritt-Anleitung zum Erstellen neuer Module.

- Modulstruktur und Lebenszyklus
- Kommunikationsmuster
- Best Practices
- Beispiele und Fehlerbehebung

</div>

---

## Schnellzugriffe

### Für Entwickler

- [Erste Schritte](/docs/getting-started)
- [Modulentwicklung](/docs/module-development)
- [Datenbankschema](/docs/database-schema)
- [API-Dokumentation](/api/v1/docs/swagger/index.html)

### Für Architekten

- [Architektur-Übersicht](/docs/architecture)
- [Datenbankschema](/docs/database-schema)
- [Modul-Entwicklung](/docs/module-development)
- [Clean Architecture](/features/clean-architecture)

### API & Integration

- [Authentifizierung & RBAC](/features/authentication)
- [Event-Bus-System](/features/event-bus)
- [REST API v1](/api/v1/docs/swagger/index.html)
- [REST API v2](/api/v2/docs/swagger/index.html)

---

## Feature-Tiefgang

<div class="features-grid">

<div class="feature-card">

### [Clean Architecture](/features/clean-architecture)

Strikte Schichtenarchitektur mit klarer Trennung der Zuständigkeiten.

</div>

<div class="feature-card">

### [Modulares System](/features/module-system)

Unabhängige Geschäftsmodule mit Plugin-Architektur.

</div>

<div class="feature-card">

### [Authentifizierung & RBAC](/features/authentication)

JWT-Authentifizierung mit rollenbasierter Zugriffskontrolle.

</div>

<div class="feature-card">

### [Datenbank & Migrationen](/features/database)

PostgreSQL mit UUID v7 und namensraumbasierten Migrationen.

</div>

<div class="feature-card">

### [Test-Infrastruktur](/features/testing)

400+ Tests mit manuellen Mocks.

</div>

<div class="feature-card">

### [Ereignisgesteuerte Architektur](/features/event-bus)

Duale Event-Bus-Adapter für asynchrone Kommunikation.

</div>

</div>

---

## Externe Ressourcen

- [GitHub Repository](https://github.com/basilex/promenade)
- [Issue Tracker](https://github.com/basilex/promenade/issues)
- [Beitragsleitfaden](https://github.com/basilex/promenade/blob/dev/CONTRIBUTING.md)
- [MIT-Lizenz](https://github.com/basilex/promenade/blob/dev/LICENSE)

---

<style>
.docs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.docs-grid > div {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.docs-grid > div:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
}

.docs-grid h3 {
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.docs-grid h3 a {
  color: var(--color-primary);
  text-decoration: none;
}

.docs-grid p {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
}

.docs-grid ul {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}
</style>

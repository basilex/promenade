---
title: "Erste Schritte"
description: "Schnellstart-Anleitung für Promenade"
weight: 5
---

## Schnellstart

Starten Sie mit Promenade in 5 Minuten.

### Voraussetzungen

- **Go 1.25+**
- **Docker & Docker Compose**
- **Make**

### 1. Repository Klonen

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
```

### 2. Abhängigkeiten Installieren

```bash
make install
```

### 3. PostgreSQL und Redis Starten

```bash
make docker-up
```

### 4. Migrationen Ausführen

```bash
make migrate
```

### 5. Anwendung Starten

```bash
make dev
```

Server startet auf **http://localhost:8081**

### 6. Health Check

```bash
curl http://localhost:8081/api/v1/health
```

### 7. Swagger Dokumentation

Öffnen Sie http://localhost:8081/api/v1/docs/swagger/index.html

---

## Nächste Schritte

- [Architektur-Übersicht](/promenade/docs/architecture)
- [Modulentwicklung](/promenade/docs/module-development)
- [Datenbankschema](/promenade/docs/database-schema)

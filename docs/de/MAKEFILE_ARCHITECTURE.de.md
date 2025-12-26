[ English](../MAKEFILE_ARCHITECTURE.md) | [ Українська](../uk/MAKEFILE_ARCHITECTURE.uk.md) |  **Deutsch** | [ Português](MAKEFILE_ARCHITECTURE.pt.md) | [ Español](MAKEFILE_ARCHITECTURE.es.md)

---

# Makefile-Architektur

Modulares Makefile-System für saubere Trennung der Verantwortlichkeiten und Skalierbarkeit.

## Struktur

```
Makefile             (63 Zeilen)  - Hauptdatei: Variablen, env-Laden, help
Makefile.dev.mk      (64 Zeilen)  - Entwicklungs-Workflow
Makefile.test.mk     (57 Zeilen)  - Test-Infrastruktur
Makefile.prod.mk     (90 Zeilen)  - Produktions-/DevOps-Operationen

Gesamt:              274 Zeilen
```

## Philosophie

**Haupt-Makefile** - Nur Gemeinsames:

- Laden von Umgebungsvariablen (`.env.development`)
- Gemeinsame Variablen (`APP_NAME`, `VERSION`, `DB_URL`, `MIGRATE`)
- Modul-Includes (`include Makefile.*.mk`)
- Gruppierter help-Befehl (zeigt alle IModule)

**Makefile.dev.mk** - Entwickler-Workflow:

```bash
make install             # Tools installieren (swag, migrate, golangci-lint)
make dev                 # Dev-Server starten (postgres + migrations + app)
make build               # Binary erstellen (mit Swagger-Generierung)
make run                 # Kompilierte Binary ausführen
make lint                # golangci-lint ausführen
make fmt                 # Code formatieren (go fmt + gofmt -s)
make deps-update         # Abhängigkeiten aktualisieren
make config-show         # Aktuelle env-Konfiguration anzeigen
```

**Makefile.test.mk** - Tests:

```bash
make test                # Unit- + Integrationstests
make test-unit           # Nur Unit-Tests (domain + usecase)
make test-integration    # Integrationstests (echte DB auf Port 5433)
make test-smoke          # Smoke-Tests (end-to-end kritische Abläufe)
make test-coverage       # HTML-Coverage-Bericht generieren
make test-db-start       # Testdatenbank starten
make test-db-stop        # Testdatenbank stoppen
```

**Makefile.prod.mk** - DevOps-Operationen:

```bash
# Docker
make docker-build        # Image erstellen (VERSION=0.1.0 ENV=dev)
make docker-run          # Erstellen + Container ausführen
make docker-up           # Dienste starten
make docker-down         # Dienste stoppen
make docker-logs         # Logs anzeigen
make docker-restart      # Container neu starten
make docker-ps           # Laufende Container anzeigen
make docker-clean        # Container + Volumes entfernen

# Migrations
make migrate-create      # Migration erstellen (NAME=xxx)
make migrate-up          # Migrationen anwenden
make migrate-down        # Letzte Migration zurückrollen
make migrate-force       # Version erzwingen (VERSION=N)
make migrate-version     # Aktuelle Version anzeigen
make migrate-status      # Status anzeigen

# Dokumentation
make swagger-all         # v1 + v2 Swagger-Dokumentation generieren

# Aufräumen
make clean               # Artefakte entfernen (bin/, docs/, coverage)
```

## Vorteile

1. **Modularität** - Jede Datei hat eine einzige Verantwortung
2. **Skalierbarkeit** - Einfach `Makefile.{stage,ci,deploy}.mk` hinzufügen
3. **Lesbarkeit** - Klare Trennung nach Kontext
4. **Wartbarkeit** - Kleine fokussierte Dateien statt 188-Zeilen-Monolith
5. **Team-freundlich** - Entwickler/QA/DevOps sehen nur relevante Befehle

## Verwendungsbeispiele

**Entwicklung:**

```bash
make help          # Alle verfügbaren Befehle anzeigen
make dev           # Entwicklung starten (am häufigsten verwendet)
make build         # Für lokale Tests erstellen
make fmt lint      # Formatieren und Lint vor Commit
```

**Tests:**

```bash
make test          # Vollständige Test-Suite vor PR ausführen
make test-unit     # Schnelles Feedback während Entwicklung
make test-smoke    # Kritische Abläufe nach Änderungen verifizieren
```

**DevOps:**

```bash
make docker-run    # In lokalem Docker bereitstellen
make migrate-up    # Datenbankmigrationen anwenden
make swagger-all   # API-Dokumentation neu generieren
make clean         # Sauberer Slate vor frischer Bereitstellung
```

## Neue Befehle hinzufügen

1. Kontext identifizieren: dev/test/prod
2. Entsprechende `Makefile.{context}.mk` bearbeiten
3. `## Kommentar` für Help-Anzeige hinzufügen
4. `make help` ausführen zur Verifizierung

Beispiel:

```makefile
# In Makefile.dev.mk
watch: ## Bei Dateiänderungen beobachten und neu laden
	air -c .air.toml
```

## Variablen

Alle gemeinsamen Variablen sind im Haupt-`Makefile`:

- `APP_NAME` - Anwendungsname
- `VERSION` - Build-Version (Standard: 0.1.0)
- `ENV` - Umgebung (dev/test/prod)
- `DB_URL` - PostgreSQL-Verbindungsstring
- `MIGRATE` - Migrate-Befehl mit DB URL
- `DOCKER_COMPOSE` - Docker Compose-Befehl

Überschreiben mit:

```bash
make docker-build VERSION=1.2.3 ENV=prod
make migrate-up DB_NAME=promenade_staging
```

## Migration von alter Struktur

Vorher (188-Zeilen-Monolith):

```
Makefile  ← Alles zusammen gemischt
```

Nachher (274 Zeilen modular):

```
Makefile          ← Gemeinsam (63 Zeilen)
Makefile.dev.mk   ← Entwicklung (64 Zeilen)
Makefile.test.mk  ← Tests (57 Zeilen)
Makefile.prod.mk  ← Produktion (90 Zeilen)
```

**Keine Breaking Changes** - Alle Befehle funktionieren genau wie vorher!

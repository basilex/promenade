# Smoke Tests

Smoke Tests überprüfen, ob die kritische API-Funktionalität in einer laufenden Umgebung korrekt funktioniert. Dies sind **Black-Box-Integrationstests**, die echte HTTP-Anfragen an die API stellen.

## Was sind Smoke Tests?

Smoke Tests sind schnelle, wesentliche Tests, die die Kernfunktionalität der Anwendung überprüfen:

- ✅ Kann die API starten?
- ✅ Sind die Endpoints erreichbar?
- ✅ Funktionieren kritische Workflows End-to-End?
- ✅ Ist die Datenbank verbunden?

**Nicht von Smoke Tests abgedeckt:**

- ❌ Grenzfälle
- ❌ Performance-/Lasttests
- ❌ Unit-Test-Logik
- ❌ Fehlerbehandlungsdetails

## Schnellstart

### Voraussetzungen

- **curl** - HTTP-Client
- **jq** - JSON-Prozessor
- Laufende API-Instanz (Standard-Port 8081)

```bash
# Prüfen ob Tools installiert sind
which curl && which jq
```

### Tests ausführen

```bash
# 1. API starten (falls noch nicht gestartet)
make dev

# 2. Smoke Tests ausführen (in anderem Terminal)
make test-smoke

# Oder direkt ausführen
cd test/smoke
./smoke_test.sh
```

### Erwartete Ausgabe

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Promenade API - Smoke Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  API:         http://localhost:8081
  Environment: dev
  Date:        2025-12-25 14:30:00

ℹ Erforderliche Tools prüfen...
✓ Alle erforderlichen Tools gefunden
ℹ Warte auf API-Bereitschaft...
✓ API ist bereit!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Smoke Tests ausführen
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Gesundheitsprüfung
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ Teste GET /health...
✓ Gesundheitsprüfung bestanden
✓ Test bestanden: 01_health

...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Test-Zusammenfassung
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Gesamt:     5
  Bestanden:  5
  Fehlerhaft: 0

✓ Alle Smoke Tests bestanden!
```

## Test-Struktur

### Tests (in Reihenfolge)

1. **01_health.sh** - Health Check Endpoint
2. **02_auth.sh** - Authentifizierungsablauf (Login → Profil abrufen)
3. **03_posts.sh** - Posts CRUD-Operationen
4. **04_profiles.sh** - Profile CRUD-Operationen
5. **05_analytics.sh** - Analytics-Metriken

### Konfiguration

**config.sh** - Test-Konfiguration:

- API-URL (Standard: `http://localhost:8081`)
- Test-Anmeldedaten
- HTTP-Timeouts

**helpers.sh** - Hilfsfunktionen:

- HTTP-Wrapper (GET, POST, PUT, DELETE)
- JSON-Assertions
- Farbige Ausgabe
- Cleanup-Utilities

## Konfiguration

### Umgebungsvariablen

```bash
# API-URL (Standard: http://localhost:8081)
export PROMENADE_API_URL="http://localhost:8081"

# Umgebungsname (optional)
export PROMENADE_ENV="dev"
```

### Benutzerdefinierte API-URL

```bash
# Test gegen Staging
PROMENADE_API_URL="https://staging.example.com" ./smoke_test.sh

# Test gegen Produktion
PROMENADE_API_URL="https://api.example.com" ./smoke_test.sh
```

## Einzelne Tests ausführen

Jeder Test kann unabhängig ausgeführt werden:

```bash
cd test/smoke

# Einzelnen Test ausführen
./tests/01_health.sh

# Spezifische Tests ausführen
./tests/02_auth.sh
./tests/03_posts.sh
```

**Hinweis:** Tests 03-05 erfordern Authentifizierung, daher wird Test 02 bei Bedarf automatisch ausgeführt.

## Neue Tests schreiben

### Vorlage

```bash
#!/bin/bash

# Smoke Test: Meine Funktion

source "$(dirname "$0")/../helpers.sh"

test_my_feature() {
    print_section "Meine Funktion"

    # Test-Logik hier
    print_info "Teste GET /my-endpoint..."
    local response=$(http_get "$API_BASE/my-endpoint" 200)

    if [ $? -ne 0 ]; then
        print_error "Test fehlgeschlagen"
        return 1
    fi

    assert_json_field "$response" ".success" "true" || return 1

    print_success "Test bestanden"
    return 0
}

# Test ausführen
test_my_feature
exit $?
```

### Zum Runner hinzufügen

`smoke_test.sh` bearbeiten:

```bash
run_test "tests/06_my_feature.sh"
```

## CI/CD-Integration

### GitHub Actions Beispiel

```yaml
name: Smoke Tests

on: [push, pull_request]

jobs:
  smoke-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq

      - name: Run migrations
        run: make migrate

      - name: Build API
        run: make build

      - name: Start API
        run: ./bin/promenade &
        env:
          ENVIRONMENT: test

      - name: Wait for API
        run: sleep 5

      - name: Run smoke tests
        run: make test-smoke

      - name: Stop API
        run: pkill promenade
```

## Fehlerbehebung

### API nicht bereit

```bash
# Prüfen ob API läuft
curl http://localhost:8081/api/v1/health

# Logs prüfen
tail -f /path/to/logs/promenade.log
```

### Test-Fehler

```bash
# Mit ausführlicher Ausgabe ausführen
bash -x ./test/smoke/tests/01_health.sh

# Einzelne Anfragen prüfen
curl -v http://localhost:8081/api/v1/health
```

### Zugriff verweigert

```bash
# Skripte ausführbar machen
chmod +x test/smoke/*.sh
chmod +x test/smoke/tests/*.sh
```

### Fehlende Abhängigkeiten

```bash
# macOS
brew install curl jq

# Ubuntu/Debian
apt-get install curl jq

# CentOS/RHEL
yum install curl jq
```

## Best Practices

1. **Tests schnell halten** - Smoke Tests sollten < 2 Minuten laufen
2. **Nur kritische Pfade testen** - Nicht jeder Endpoint braucht einen Smoke Test
3. **Eindeutige Testdaten verwenden** - Zeitstempel in Test-E-Mails/Titeln verhindern Konflikte
4. **Nach Tests aufräumen** - Testdaten löschen um Verschmutzung zu vermeiden
5. **Tests idempotent machen** - Sollten auch bei mehrfacher Ausführung bestehen
6. **Assertions verwenden** - Antwortstruktur prüfen, nicht nur Statuscodes
7. **Schnell fehlschlagen** - Bei erstem Fehler stoppen um Zeit zu sparen

## Was kommt als Nächstes?

Nach bestandenen Smoke Tests erwägen:

- **Lasttests** - Siehe `test/stress/` für wrk-basierte Stresstests
- **End-to-End Tests** - Cypress/Playwright für UI-Tests
- **Sicherheitstests** - OWASP ZAP, Burp Suite
- **Chaos-Tests** - Ausfälle simulieren

## Support

- **Issues** - Bugs in Smoke Tests melden
- **Fragen** - Im Team-Chat fragen
- **Verbesserungen** - PR mit neuen Tests einreichen

---

**Erstellt:** 25. Dezember 2025
**Gepflegt von:** Promenade DevOps Team

# Stress- und Lasttests

Performance- und Lasttests für Promenade API mit **wrk** - einem modernen HTTP-Benchmarking-Tool.

## Was sind Stresstests?

Stresstests bringen die API an ihre Grenzen, um Folgendes zu finden:

-  **Maximaler Durchsatz** (Anfragen pro Sekunde)
-  **Latenz unter Last** (p50, p95, p99)
-  **Schwachstellen** (wann versagt es?)
-  **Ressourcennutzung** (CPU, Speicher, Verbindungen)

## Voraussetzungen

### wrk installieren

**macOS:**

```bash
brew install wrk
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt-get install wrk
```

**Aus dem Quellcode:**

```bash
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/
```

Installation überprüfen:

```bash
wrk --version
```

## Schnellstart

```bash
# 1. API starten
make dev

# 2. Stresstests ausführen (in anderem Terminal)
cd test/stress
./stress_test.sh

# Oder über Makefile
make stress-test
```

## Test-Szenarien

### 1. Health Check Lasttest

**Ziel:** Basis-Performance, minimale Logik

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

**Erwartete Ergebnisse:**

- RPS: 10,000+
- Latenz (p99): < 50ms
- Null Fehler

### 2. Authentifizierungs-Lasttest

**Ziel:** Auth-Flow unter Last testen

```bash
wrk -t4 -c100 -d30s -s scenarios/auth.lua http://localhost:8081
```

**Erwartete Ergebnisse:**

- RPS: 500-1,000
- Latenz (p99): < 200ms
- Fehlerrate: < 1%

### 3. Posts CRUD Lasttest

**Ziel:** Datenbankintensive Operationen

```bash
wrk -t4 -c100 -d30s -s scenarios/posts.lua http://localhost:8081
```

**Erwartete Ergebnisse:**

- RPS: 300-500
- Latenz (p99): < 300ms
- Fehlerrate: < 1%

### 4. Gleichzeitige Benutzersimulation

**Ziel:** Realistisches Benutzerverhalten

```bash
wrk -t8 -c200 -d60s -s scenarios/user_journey.lua http://localhost:8081
```

**Erwartete Ergebnisse:**

- Gleichzeitige Benutzer: 200
- Sitzungsdauer: 60s
- Realistische Denkzeit zwischen Anfragen

## Laststufen

### Leichte Last (Aufwärmen)

```bash
wrk -t2 -c10 -d10s http://localhost:8081/api/v1/health
```

- 2 Threads, 10 Verbindungen, 10 Sekunden
- Ziel: ~1,000 RPS

### Mittlere Last

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

- 4 Threads, 100 Verbindungen, 30 Sekunden
- Ziel: ~10,000 RPS

### Schwere Last

```bash
wrk -t8 -c500 -d60s http://localhost:8081/api/v1/health
```

- 8 Threads, 500 Verbindungen, 60 Sekunden
- Ziel: Grenzen finden

### Stresstest (Schwachstelle finden)

```bash
wrk -t12 -c1000 -d120s http://localhost:8081/api/v1/health
```

- 12 Threads, 1000 Verbindungen, 2 Minuten
- Ziel: System brechen, Engpässe finden

## wrk Parameter erklärt

```bash
wrk -t4 -c100 -d30s -s script.lua http://localhost:8081/api/v1/endpoint
    ↑    ↑    ↑      ↑
                   Lua-Skript für komplexe Szenarien
             Dauer (10s, 30s, 1m, 2h)
         Verbindungen (gleichzeitige Anfragen)
     Threads (zu verwendende CPU-Kerne)
```

**Empfohlen:**

- **Threads (-t):** Anzahl der CPU-Kerne (2-8)
- **Verbindungen (-c):** 10-1000 (niedrig anfangen, erhöhen)
- **Dauer (-d):** 10s-60s (länger für Produktionstests)

## Lua-Skripte

### Basis-POST-Anfrage

```lua
-- scenarios/auth_login.lua
wrk.method = "POST"
wrk.body   = '{"email":"test@example.com","password":"password123"}'
wrk.headers["Content-Type"] = "application/json"
```

### Dynamische Anfragen mit Zustand

```lua
-- scenarios/posts_crud.lua
local counter = 0
local token = "Bearer YOUR_TOKEN_HERE"

request = function()
    counter = counter + 1
    local path = "/api/v1/posts"
    local body = string.format('{"title":"Post %d","content":"Test"}', counter)

    return wrk.format("POST", path, {
        ["Content-Type"] = "application/json",
        ["Authorization"] = token
    }, body)
end

response = function(status, headers, body)
    if status ~= 201 then
        print("Fehler: " .. status .. " - " .. body)
    end
end
```

## Ergebnisse interpretieren

### Beispielausgabe

```
Running 30s test @ http://localhost:8081/api/v1/health
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.12ms    2.34ms  45.23ms   89.34%
    Req/Sec     4.89k   456.23    5.67k    76.45%
  586234 requests in 30.00s, 123.45MB read
Requests/sec:  19541.13
Transfer/sec:      4.11MB
```

**Wichtige Metriken:**

- **Latency Avg:** Durchschnittliche Antwortzeit (niedriger ist besser)
- **Latency Stdev:** Konsistenz (niedriger ist besser)
- **Latency Max:** Schlimmster Fall (auf Spitzen achten)
- **Req/Sec:** Durchsatz pro Thread
- **Requests/sec:** Gesamtdurchsatz (RPS)
- **Transfer/sec:** Verwendete Netzwerkbandbreite

### Was ist gut?

| Metrik              | Gut      | Warnung  | Kritisch |
| ------------------- | -------- | -------- | -------- |
| **Health endpoint** | >10k RPS | <5k RPS  | <1k RPS  |
| **Auth endpoints**  | >500 RPS | <200 RPS | <100 RPS |
| **CRUD endpoints**  | >300 RPS | <100 RPS | <50 RPS  |
| **p99 Latenz**      | <100ms   | <500ms   | >1s      |
| **Fehlerrate**      | 0%       | <1%      | >5%      |

### Warnsignale 

- **Hoher Stdev:** Inkonsistente Performance (untersuchen)
- **Max >> Avg:** Gelegentlich sehr langsame Anfragen (Caching-Problem?)
- **Fehler:** Datenbankverbindungspool erschöpft?
- **Lineare Degradierung:** Skaliert nicht mit Verbindungen

## Überwachung während Tests

### Terminal 1: API ausführen

```bash
make dev
```

### Terminal 2: Ressourcen überwachen

```bash
# CPU/Speicher beobachten
watch -n 1 'ps aux | grep promenade | grep -v grep'

# Oder htop verwenden
htop -p $(pgrep promenade)
```

### Terminal 3: Datenbank überwachen

```bash
# PostgreSQL-Verbindungen
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT count(*) FROM pg_stat_activity;"'

# Datenbanklast
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT * FROM pg_stat_database WHERE datname = '\''promenade_dev'\'';"'
```

### Terminal 4: wrk ausführen

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

## Optimierungstipps

### Wenn RPS niedrig ist

1. **Datenbankverbindungspool prüfen:**

   ```yaml
   # config/app.dev.yaml
   database:
     max_open_conns: 100 # Erhöhen
     max_idle_conns: 25 # Erhöhen
   ```

2. **HTTP/2 aktivieren**

3. **Indizes hinzufügen:**

   ```sql
   CREATE INDEX idx_posts_user_id ON user_posts(user_id);
   ```

4. **Caching verwenden:**
   - Redis für Session-Cache
   - In-Memory-Cache für Referenzdaten

### Wenn Latenz hoch ist

1. **Code profilen:**

   ```bash
   go tool pprof http://localhost:8081/debug/pprof/profile
   ```

2. **Langsame Abfragen prüfen:**

   ```sql
   SELECT query, mean_exec_time, calls
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;
   ```

3. **JSON-Serialisierung optimieren**

### Bei Fehlern

1. **Datenbankverbindungslimit erreicht:**

   ```
   Error: pq: sorry, too many clients already
   ```

   → `max_connections` in PostgreSQL erhöhen

2. **Context deadline exceeded:**
   → Timeout in Konfiguration erhöhen

3. **Out of memory:**
   → Verbindungspool reduzieren oder mehr RAM hinzufügen

## Ressourcen

- [wrk GitHub](https://github.com/wg/wrk)
- [wrk Lua Scripting](https://github.com/wg/wrk/blob/master/SCRIPTING)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Go Performance Tips](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)

---

**Erstellt:** 25. Dezember 2025
**Gepflegt von:** Promenade DevOps Team

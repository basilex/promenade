# UUID v7 Implementierungsleitfaden

## Überblick

Dieses Projekt wurde von UUID v4 auf UUID v7 für die Primärschlüsselgenerierung migriert. UUID v7 bietet erhebliche Leistungsvorteile und behält gleichzeitig die globale Eindeutigkeit und verteilten Generierungsfähigkeiten von UUID v4 bei.

## Warum UUID v7?

### Vorteile gegenüber UUID v4

1. **Zeitlich geordnet**: UUIDs sind natürlich nach Erstellungszeit sortiert
2. **Bessere B-tree-Leistung**: Reduziert Page-Splits in PostgreSQL-Indizes (20-50% schnellere INSERTs in Benchmarks)
3. **Reduzierte Fragmentierung**: Sequenzielle-ish IDs minimieren Index-Fragmentierung
4. **Cache-freundlich**: Bessere Referenzlokalität für kürzlich erstellte Datensätze
5. **Extrahierbarer Timestamp**: Erstellungszeit kann aus der UUID selbst abgerufen werden
6. **Weiterhin global eindeutig**: Behält die Eindeutigkeitsgarantien von UUID v4 bei

### Vergleich mit anderen Strategien

| Strategie        | Vorteile                                          | Nachteile                                            |
| ---------------- | ------------------------------------------------- | ---------------------------------------------------- |
| **UUID v7**      | [+] Zeitlich geordnet, global eindeutig, verteilt | Etwas größer als BIGINT (16 Bytes)                   |
| UUID v4          | Global eindeutig, verteilt                        | [X] Zufällig = schlechte Index-Lokalität             |
| SERIAL/BIGSERIAL | Klein, sequenziell, schnell                       | [X] Single Point of Failure, Replikationsprobleme    |
| ULID             | Ähnlich wie v7, base32-kodiert                    | Weniger Standard, begrenzte Bibliotheksunterstützung |
| Snowflake ID     | Schnell, zeitlich geordnet                        | Erfordert Koordinationsdienst                        |

## Implementierung

### PostgreSQL-Einrichtung

Das Projekt enthält zwei Migrationen:

1. **000005_add_uuidv7_support.sql**: Fügt die Funktion `uuid_generate_v7()` zu PostgreSQL hinzu
2. **000006_switch_tables_to_uuidv7.sql**: Aktualisiert Tabellenstandardwerte auf v7

```sql
-- UUID v7 in PostgreSQL generieren
SELECT uuid_generate_v7();
-- Beispiel: 018d2f07-0f3c-7000-8000-123456789abc
```

### Go-Code

Verwenden Sie das `pkg/uuidv7`-Paket:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Neue UUID v7 generieren
id := uuidv7.New()

// Mit spezifischem Timestamp generieren (nützlich für Tests)
id := uuidv7.NewWithTime(time.Now())

// Erstellungs-Timestamp aus UUID extrahieren
timestamp := uuidv7.ExtractTime(id)

// Prüfen ob UUID Version 7 ist
if uuidv7.IsV7(id) {
    // ...
}
```

### Migrationspfad

Bestehende Datensätze mit UUID v4 funktionieren weiterhin. Das System unterstützt gemischte UUIDs:

| Datensatztyp    | UUID-Version | Hinweise       |
| --------------- | ------------ | -------------- |
| Alte Datensätze | UUID v4      | Vor Migration  |
| Neue Datensätze | UUID v7      | Nach Migration |

## Leistungsaspekte

### Wann UUID v7 Verwenden

[+] **Gut geeignet:**

- Verteilte Systeme, in denen mehrere Knoten IDs generieren
- Multi-Region-Deployments
- Microservices-Architektur
- APIs, bei denen Clients IDs generieren müssen
- Tabellen mit hoher INSERT-Rate
- Wenn natürliche zeitbasierte Sortierung benötigt wird

[X] **Überdenken wenn:**

- Single-Database-System ohne Verteilungsbedarf
- Speicher extrem begrenzt ist (UUID = 16 Bytes vs BIGINT = 8 Bytes)
- Sequenzielle Lückenanalyse benötigt wird (stattdessen SERIAL verwenden)

### PostgreSQL Index-Optimierung

UUID v7 funktioniert am besten mit B-tree-Indizes aufgrund seiner zeitlichen Ordnung:

```sql
-- Gut: Primärschlüssel verwendet automatisch B-tree
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

-- Gut: Zeitbasierte Abfragen profitieren von UUID v7-Sortierung
CREATE INDEX idx_users_created ON users(created_at DESC);

-- Optimal: UUID v7 mit Timestamp für Bereichsabfragen kombinieren
SELECT * FROM users
WHERE id >= uuid_generate_v7_from_timestamp('2024-01-01'::timestamp)
ORDER BY id DESC;
```

### Leistungsüberwachung

INSERT-Leistung vor/nach Migration vergleichen:

```sql
-- Index-Bloat prüfen
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- INSERT-Leistung überwachen
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');
```

## Empfehlungen für Fremdschlüssel

### UUID v7 auch für Fremdschlüssel Verwenden

```go
type Product struct {
    ID        uuidv7.UUID `db:"id"`        // UUID v7 Primärschlüssel
    UserID    uuidv7.UUID `db:"user_id"`   // UUID v7 Fremdschlüssel
    Name      string      `db:"name"`
    CreatedAt time.Time   `db:"created_at"`
}
```

### Fremdschlüssel Indizieren

Immer Fremdschlüssel für JOIN-Leistung indizieren:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL
);

-- Kritisch: Fremdschlüssel indizieren
CREATE INDEX idx_products_user_id ON products(user_id);
```

### Zusammengesetzte Schlüssel mit UUID v7

Für Many-to-Many-Beziehungen:

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Index für umgekehrte Suche
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

## Alternative Strategien (Zukünftige Überlegung)

### Option 1: ULID (Universally Unique Lexicographically Sortable Identifier)

Ähnlich wie UUID v7, verwendet aber base32-Kodierung (kürzere String-Darstellung):

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (ULID - 26 Zeichen)
018d2f07-0f3c-7000-8000-123456789abc  (UUID v7 - 36 Zeichen)
```

**Wann verwenden**: Wenn Sie häufig IDs Benutzern anzeigen und kürzere Strings wünschen.

### Option 2: Snowflake IDs

Twitters Snowflake: 64-Bit-Integer mit Timestamp, Datacenter-ID und Sequenz.

```go
// Beispielstruktur (in diesem Projekt nicht implementiert)
// 41 Bits: Timestamp
// 10 Bits: Datacenter/Worker-ID
// 12 Bits: Sequenznummer
```

**Wann verwenden**: Extrem hoher Durchsatz (Twitter-Skala), wo Koordination akzeptabel ist.

### Option 3: PostgreSQL SERIAL/BIGSERIAL

Traditionelle Auto-Increment-Integer:

```sql
CREATE TABLE simple_table (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4, ...
    name VARCHAR(255)
);
```

**Wann verwenden**: Einfache Single-Database-Anwendungen ohne Verteilungsbedarf.

### Option 4: Zusammengesetzte Natürliche Schlüssel

Bedeutungsvolle Geschäftsdaten als Schlüssel verwenden:

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'GB', 'FR'
    name VARCHAR(100)
);
```

**Wann verwenden**: Wenn natürliche Schlüssel stabil, kurz und wirklich eindeutig sind.

## Testen

Test-Suite ausführen, um UUID v7-Implementierung zu verifizieren:

```bash
# UUID v7-Paket testen
go test -v ./pkg/uuidv7/

# Mit Benchmarks ausführen
go test -bench=. ./pkg/uuidv7/

# Integration mit Datenbank testen
make test-integration
```

Erwartete Benchmark-Ergebnisse (ungefähr):

```
BenchmarkNew-8      3000000    450 ns/op    (UUID v7)
BenchmarkNewV4-8    2800000    480 ns/op    (UUID v4)
```

## Migrations-Checkliste

- [x] Funktion `uuid_generate_v7()` zu PostgreSQL hinzufügen
- [x] Tabellenstandardwerte auf UUID v7 aktualisieren
- [x] Go-Code aktualisieren, um Paket `pkg/uuidv7` zu verwenden
- [x] Alle Use Cases aktualisieren (auth, product, role)
- [ ] Migrationen ausführen: `make migrate-up`
- [ ] UUID-Generierung testen: `go test ./pkg/uuidv7/`
- [ ] Produktionsmetriken nach Deployment überwachen

## Weiterführende Literatur

- [RFC 9562 - UUID Version 7 (Draft)](https://datatracker.ietf.org/doc/draft-ietf-uuidrev-rfc4122bis/)
- [PostgreSQL UUID Performance](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/)
- [UUID v7 in Production (Blog post)](https://buildkite.com/blog/goodbye-integers-hello-uuids)

## Fragen?

Siehe `.github/copilot-instructions.md` für Projektkonventionen oder fragen Sie im Team-Chat.

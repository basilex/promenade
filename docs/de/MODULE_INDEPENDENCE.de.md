# Modul-Unabhängigkeitsverifizierung

[ English](../MODULE_INDEPENDENCE.md) | [ Українська](../uk/MODULE_INDEPENDENCE.uk.md) |  **Deutsch** | [ Português](../pt/MODULE_INDEPENDENCE.pt.md) | [ Español](../es/MODULE_INDEPENDENCE.es.md)

## Posts-Modulstruktur

Das `posts`-Modul demonstriert vollständige Unabhängigkeit vom Kernsystem und implementiert seinen eigenen vollständigen Clean Architecture Stack:

```
internal/modules/posts/
 domain/
    entity/          # Post-Entity + Domain-Fehler
    repository/      # Repository-Interface
 usecase/             # Geschäftslogik-Ebene
 adapter/
    http/           # HTTP-Handler & DTOs
    repository/     # Datenbankimplementierung
 tests/              # Modulspezifische Tests
 module.go           # Modulintegration
 register.go         # Auto-Registrierung
```

## Unabhängigkeitsanalyse

### Keine Core-Abhängigkeiten

Verifiziert durch Suche: `grep -r "github.com/basilex/promenade/internal/(domain|usecase|adapter)" internal/modules/posts/`

**Ergebnis**: NULL Treffer - Modul importiert keine internen Core-Pakete.

### Nur gemeinsame Abhängigkeiten

Das Modul importiert nur:

- `pkg/*` - Gemeinsame Utilities (logger, uuidv7, pagination, bus, response, module SDK)
- `github.com/gin-gonic/gin` - HTTP-Framework
- `github.com/jmoiron/sqlx` - Datenbankbibliothek
- Standard-Library-Pakete

### Vollständiger vertikaler Slice

Jede Ebene innerhalb des Moduls implementiert:

| Ebene                 | Ort                                    | Abhängigkeiten              |
| --------------------- | -------------------------------------- | --------------------------- |
| **Domain-Entity**     | `domain/entity/post.go`                | Nur `pkg/uuidv7`            |
| **Domain-Repository** | `domain/repository/post_repository.go` | Domain-Entity, pkg          |
| **Use Case**          | `usecase/post_usecase.go`              | Nur Domain                  |
| **Repository-Impl**   | `adapter/repository/postgres/`         | Domain, pkg/database        |
| **HTTP-Handler**      | `adapter/http/handler/`                | Use case, DTO, pkg/response |
| **DTOs**              | `adapter/http/dto/`                    | Domain-Entity               |

### Vorteile der Modulisolierung

1. **Unabhängige Entwicklung**: Kann isoliert entwickelt/getestet werden
2. **Wiederverwendbarkeit**: Kann mit pkg/ in ein anderes Projekt kopiert werden
3. **Keine Breaking Changes**: Core-Refactoring beeinflusst Modul nicht
4. **Klare Grenzen**: Alle Abhängigkeiten explizit und minimal
5. **Einfaches Testen**: Nur Domain-Interfaces mocken, nicht Core-Services

## Neue IModule erstellen

Um ein neues unabhängiges Modul zu erstellen:

1. Verzeichnisstruktur erstellen:

   ```
   internal/modules/{name}/
    domain/entity/
    domain/repository/
    usecase/
    adapter/http/handler/
    adapter/http/dto/
    adapter/repository/postgres/
    tests/
    module.go
    register.go
   ```

2. Implementierung aus Core kopieren oder von Grund auf neu schreiben
3. Alle Imports auf Modulpfade aktualisieren
4. Modulspezifische Fehler in `domain/entity/errors.go` definieren
5. `module.IModule`-Interface in `module.go` implementieren
6. Auto-Registrierung in `register.go` mittels `init()` durchführen

## Verifizierungsbefehl

```bash
# Auf interne Core-Abhängigkeiten prüfen
grep -r "github.com/basilex/promenade/internal/\(domain\|usecase\|adapter\)" \
  internal/modules/posts/ || echo " Modul ist unabhängig"
```

## Aktueller Status

- **posts**: Vollständig unabhängig, vollständige Clean Architecture
- **comments**: Benötigt Refactoring (derzeit Wrapper)
- **warehouse**: Benötigt Refactoring (derzeit Wrapper)

## Nächste Schritte

1. `comments`-Modul nach `posts`-Muster refactoren
2. `warehouse`-Modul refactoren
3. Integrationstests zu `tests/`-Verzeichnissen hinzufügen
4. Modulspezifische APIs dokumentieren
5. Modulentwicklungsleitfaden mit Vorlagen erstellen

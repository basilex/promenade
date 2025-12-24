# Promenade Dokumentationsübersetzungen

[🇬🇧 English](../TRANSLATIONS.md) | [🇺🇦 Українська](../uk/TRANSLATIONS.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/TRANSLATIONS.pt.md) | [🇪🇸 Español](../es/TRANSLATIONS.es.md)

---

 **Available Languages / Доступні мови / Verfügbare Sprachen / Idiomas Disponíveis**

Die Promenade-Dokumentation ist in mehreren Sprachen verfügbar, um sie für Entwickler weltweit zugänglich zu machen.

---

## Verfügbare Übersetzungen

### 🇬🇧 English (Primär)

- Status: **Vollständig**
- Ort: `/docs/` und `README.md`
- Betreuer: Kernteam

### 🇺🇦 Українська (Ukrainian)

- Status: **Aktiv** 🟢
- Ort: `/docs/uk/` und `README.uk.md`
- Abdeckung: README (100%), Kerndokumente (in Bearbeitung)
- [Ukrainische Dokumente durchsuchen →](uk/README.md)

### 🇩🇪 Deutsch (German)

- Status: **Geplant** 📋
- Ort: `/docs/de/` und `README.de.md`
- Abdeckung: Struktur bereit, Übersetzungen benötigt
- [Deutsche Dokumente durchsuchen →](de/README.md)

### 🇵🇹 Português (Portuguese)

- Status: **Geplant** 📋
- Ort: `/docs/pt/` und `README.pt.md`
- Abdeckung: Struktur bereit, Übersetzungen benötigt
- [Portugiesische Dokumente durchsuchen →](pt/README.md)

### 🇪🇸 Español (Spanish)

- Status: **Geplant** 📋
- Ort: `/docs/es/` und `README.es.md`
- Abdeckung: Struktur bereit, Übersetzungen benötigt
- [Spanische Dokumente durchsuchen →](es/README.md)

---

## Übersetzungspriorität

Dokumente werden in der folgenden Reihenfolge übersetzt:

###  Hohe Priorität (Wesentlich für den Einstieg)

1. `README.md` - Projektübersicht und Schnellstart
2. `docs/ARCHITECTURE_QUICKREF.md` - Architektur-Kurzreferenz
3. `docs/MODULE_DEVELOPMENT.md` - Modul-Entwicklungshandbuch

### 🔶 Mittlere Priorität (Wichtig für die Entwicklung)

4. `docs/TESTING_GUIDE.md` - Best Practices für Tests
5. `docs/MODULE_INDEPENDENCE.md` - Modulprinzipien
6. `docs/ARCHITECTURE_OVERVIEW.md` - Detaillierte Architektur

### 🔷 Niedrige Priorität (Fortgeschrittene Themen)

7. `docs/PURGE_ARCHITECTURE.md` - Purge-System
8. `docs/REDIS_BUS_TESTING.md` - Event Bus Testing
9. Technische Referenzdokumente

---

## Dateinamenskonvention

Alle übersetzten Dokumente folgen einem konsistenten Namensmuster:

```
Original:     docs/FILENAME.md
Ukrainian:    docs/uk/FILENAME.uk.md
German:       docs/de/FILENAME.de.md
Portuguese:   docs/pt/FILENAME.pt.md
Spanish:      docs/es/FILENAME.es.md
```

**Sprachcodes** folgen dem ISO 639-1 Standard:

- `uk` - Ukrainian (українська)
- `de` - German (Deutsch)
- `pt` - Portuguese (Português)
- `es` - Spanish (Español)

---

## Wie man eine Übersetzung hinzufügt

### 1. Wählen Sie ein Dokument

Wählen Sie ein nicht übersetztes Dokument aus der obigen Prioritätsliste.

### 2. Erstellen Sie eine Übersetzungsdatei

```bash
# Example: Translate ARCHITECTURE_QUICKREF.md to Ukrainian
touch docs/uk/ARCHITECTURE_QUICKREF.uk.md
```

### 3. Fügen Sie den Sprachwähler hinzu

Am Anfang des **ursprünglichen englischen Dokuments** fügen Sie hinzu:

```markdown
🇬🇧 **English** | [🇺🇦 Українська](uk/FILENAME.uk.md) | [🇩🇪 Deutsch](de/FILENAME.de.md) | [🇵🇹 Português](pt/FILENAME.pt.md) | [🇪🇸 Español](es/FILENAME.es.md)
```

Am Anfang des **übersetzten Dokuments** fügen Sie hinzu:

```markdown
🇬🇧 [English](../FILENAME.md) | 🇩🇪 **Deutsch**
```

### 4. Aktualisieren Sie den Fortschritt

Aktualisieren Sie die entsprechende Sprach-README (`docs/{lang}/README.md`), um das Dokument als abgeschlossen zu markieren.

### 5. Erstellen Sie einen Pull Request

- Titel: `docs: Add [Language] translation for [Document]`
- Beispiel: `docs: Add Ukrainian translation for ARCHITECTURE_QUICKREF`

---

## Übersetzungsrichtlinien

###  TUN SIE:

- Übersetzen Sie technische Begriffe konsistent (verwenden Sie das Glossar unten)
- Lassen Sie Code-Beispiele unverändert (Code ist universell)
- Bewahren Sie alle Links (aktualisieren Sie Pfade zu übersetzten Versionen, wenn verfügbar)
- Behalten Sie die gleiche Dokumentstruktur bei
- Verwenden Sie native Sprachkonventionen (z.B. Datumsformate, Anführungszeichen)

###  TUN SIE NICHT:

- Übersetzen Sie keine Dateinamen oder Pfade im Code
- Ändern Sie keine Code-Beispiele oder Befehlsausgaben
- Entfernen oder überspringen Sie keine Abschnitte
- Übersetzen Sie keine Markennamen (Promenade, PostgreSQL, Redis)
- Übersetzen Sie keine Programmier-Schlüsselwörter (`func`, `type`, `interface`)

---

## Glossar technischer Begriffe

Um Konsistenz über Übersetzungen hinweg zu gewährleisten:

| English            | 🇺🇦 Українська            | 🇩🇪 Deutsch         | 🇵🇹 Português      | 🇪🇸 Español          |
| ------------------ | ------------------------ | ------------------ | ----------------- | ------------------- |
| Module             | Модуль                   | Modul              | Módulo            | Módulo              |
| Repository         | Репозиторій              | Repository         | Repositório       | Repositorio         |
| Use Case           | Use Case / Бізнес-логіка | Anwendungsfall     | Caso de Uso       | Caso de Uso         |
| Handler            | Хендлер                  | Handler            | Manipulador       | Manejador           |
| Entity             | Сутність                 | Entität            | Entidade          | Entidad             |
| Migration          | Міграція                 | Migration          | Migração          | Migración           |
| Testing            | Тестування               | Testen             | Teste             | Pruebas             |
| Clean Architecture | Clean Architecture       | Clean Architecture | Arquitetura Limpa | Arquitectura Limpia |

---

## Mitwirkende

Besonderer Dank an die Übersetzungsmitwirkenden:

- 🇺🇦 Ukrainian: [@basilex](https://github.com/basilex) and AI Assistant
- 🇩🇪 German: _Mitwirkende willkommen!_
- 🇵🇹 Portuguese: _Mitwirkende willkommen!_
- 🇪🇸 Spanish: _Mitwirkende willkommen!_

**Möchten Sie mitwirken?** Schauen Sie sich den [Contributing Guide](../README.md#contributing) an und wählen Sie ein Dokument aus der obigen Prioritätsliste!

---

## Support

- **Dokumentationsprobleme**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Übersetzungsfragen**: alexander.vasilenko@gmail.com
- **Community**: Beteiligen Sie sich an Diskussionen in Ihrer Sprache!

---

**Promenade für Entwickler weltweit zugänglich machen**

# Promenade Documentation Translations

🌍 **Available Languages / Доступні мови / Verfügbare Sprachen / Idiomas Disponíveis**

Promenade documentation is available in multiple languages to make it accessible to developers worldwide.

---

## Available Translations

### 🇬🇧 English (Primary)

- Status: **Complete** ✅
- Location: `/docs/` and `README.md`
- Maintainer: Core team

### 🇺🇦 Українська (Ukrainian)

- Status: **Active** 🟢
- Location: `/docs/uk/` and `README.uk.md`
- Coverage: README (100%), Core docs (in progress)
- [Browse Ukrainian docs →](uk/README.md)

### 🇩🇪 Deutsch (German)

- Status: **Planned** 📋
- Location: `/docs/de/` and `README.de.md`
- Coverage: Structure ready, translations needed
- [Browse German docs →](de/README.md)

### 🇵🇹 Português (Portuguese)

- Status: **Planned** 📋
- Location: `/docs/pt/` and `README.pt.md`
- Coverage: Structure ready, translations needed
- [Browse Portuguese docs →](pt/README.md)

### 🇪🇸 Español (Spanish)

- Status: **Planned** 📋
- Location: `/docs/es/` and `README.es.md`
- Coverage: Structure ready, translations needed
- [Browse Spanish docs →](es/README.md)

---

## Translation Priority

Documents are translated in the following priority order:

### 🔥 High Priority (Essential for getting started)

1. `README.md` - Project overview and quick start
2. `docs/ARCHITECTURE_QUICKREF.md` - Architecture quick reference
3. `docs/MODULE_DEVELOPMENT.md` - Module development guide

### 🔶 Medium Priority (Important for development)

4. `docs/TESTING_GUIDE.md` - Testing best practices
5. `docs/MODULE_INDEPENDENCE.md` - Module principles
6. `docs/ARCHITECTURE_OVERVIEW.md` - Detailed architecture

### 🔷 Low Priority (Advanced topics)

7. `docs/PURGE_ARCHITECTURE.md` - Purge system
8. `docs/REDIS_BUS_TESTING.md` - Event bus testing
9. Technical reference docs

---

## File Naming Convention

All translated documents follow a consistent naming pattern:

```
Original:     docs/FILENAME.md
Ukrainian:    docs/uk/FILENAME.uk.md
German:       docs/de/FILENAME.de.md
Portuguese:   docs/pt/FILENAME.pt.md
Spanish:      docs/es/FILENAME.es.md
```

**Language codes** follow ISO 639-1 standard:

- `uk` - Ukrainian (українська)
- `de` - German (Deutsch)
- `pt` - Portuguese (Português)
- `es` - Spanish (Español)

---

## How to Add a Translation

### 1. Choose a Document

Pick an untranslated document from the priority list above.

### 2. Create Translation File

```bash
# Example: Translate ARCHITECTURE_QUICKREF.md to Ukrainian
touch docs/uk/ARCHITECTURE_QUICKREF.uk.md
```

### 3. Add Language Selector

At the top of the **original English document**, add:

```markdown
🇬🇧 **English** | [🇺🇦 Українська](uk/FILENAME.uk.md) | [🇩🇪 Deutsch](de/FILENAME.de.md) | [🇵🇹 Português](pt/FILENAME.pt.md) | [🇪🇸 Español](es/FILENAME.es.md)
```

At the top of the **translated document**, add:

```markdown
🇬🇧 [English](../FILENAME.md) | 🇺🇦 **Українська**
```

### 4. Update Progress

Update the relevant language README (`docs/{lang}/README.md`) to mark the document as complete.

### 5. Submit Pull Request

- Title: `docs: Add [Language] translation for [Document]`
- Example: `docs: Add Ukrainian translation for ARCHITECTURE_QUICKREF`

---

## Translation Guidelines

### ✅ DO:

- Translate technical terms consistently (use glossary below)
- Keep code examples unchanged (code is universal)
- Preserve all links (update paths to translated versions when available)
- Maintain the same document structure
- Use native language conventions (е.g., date formats, quotes)

### ❌ DON'T:

- Translate file names or paths in code
- Change code examples or command outputs
- Remove or skip sections
- Translate brand names (Promenade, PostgreSQL, Redis)
- Translate programming keywords (`func`, `type`, `interface`)

---

## Technical Term Glossary

To ensure consistency across translations:

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

## Contributors

Special thanks to translation contributors:

- 🇺🇦 Ukrainian: [@basilex](https://github.com/basilex) and AI Assistant
- 🇩🇪 German: _Contributors welcome!_
- 🇵🇹 Portuguese: _Contributors welcome!_
- 🇪🇸 Spanish: _Contributors welcome!_

**Want to contribute?** Check the [Contributing Guide](../README.md#contributing) and pick a document from the priority list above!

---

## Support

- **Documentation Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Translation Questions**: alexander.vasilenko@gmail.com
- **Community**: Join discussions in your language!

---

**Making Promenade accessible to developers worldwide** 🌍

# Переклади документації Promenade

[🇬🇧 English](../TRANSLATIONS.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](TRANSLATIONS.de.md) | [🇵🇹 Português](TRANSLATIONS.pt.md) | [🇪🇸 Español](TRANSLATIONS.es.md)

---

🌍 **Available Languages / Доступні мови / Verfügbare Sprachen / Idiomas Disponíveis**

Документація Promenade доступна багатьма мовами, щоб зробити її доступною для розробників з усього світу.

---

## Доступні переклади

### 🇬🇧 English (Основна)

- Статус: **Повна** ✅
- Розташування: `/docs/` та `README.md`
- Супроводжувач: Основна команда

### 🇺🇦 Українська (Ukrainian)

- Статус: **Активна** 🟢
- Розташування: `/docs/uk/` та `README.uk.md`
- Покриття: README (100%), Основні документи (в процесі)
- [Переглянути українські документи →](uk/README.md)

### 🇩🇪 Deutsch (German)

- Статус: **Заплановано** 📋
- Розташування: `/docs/de/` та `README.de.md`
- Покриття: Структура готова, потрібні переклади
- [Переглянути німецькі документи →](de/README.md)

### 🇵🇹 Português (Portuguese)

- Статус: **Заплановано** 📋
- Розташування: `/docs/pt/` та `README.pt.md`
- Покриття: Структура готова, потрібні переклади
- [Переглянути португальські документи →](pt/README.md)

### 🇪🇸 Español (Spanish)

- Статус: **Заплановано** 📋
- Розташування: `/docs/es/` та `README.es.md`
- Покриття: Структура готова, потрібні переклади
- [Переглянути іспанські документи →](es/README.md)

---

## Пріоритет перекладів

Документи перекладаються в наступному порядку пріоритетності:

### 🔥 Високий пріоритет (Необхідні для початку роботи)

1. `README.md` - Огляд проєкту та швидкий старт
2. `docs/ARCHITECTURE_QUICKREF.md` - Коротка довідка з архітектури
3. `docs/MODULE_DEVELOPMENT.md` - Посібник з розробки модулів

### 🔶 Середній пріоритет (Важливі для розробки)

4. `docs/TESTING_GUIDE.md` - Кращі практики тестування
5. `docs/MODULE_INDEPENDENCE.md` - Принципи модулів
6. `docs/ARCHITECTURE_OVERVIEW.md` - Детальна архітектура

### 🔷 Низький пріоритет (Просунуті теми)

7. `docs/PURGE_ARCHITECTURE.md` - Система очищення
8. `docs/REDIS_BUS_TESTING.md` - Тестування event bus
9. Технічні довідкові документи

---

## Конвенція іменування файлів

Усі перекладені документи дотримуються єдиного шаблону іменування:

```
Original:     docs/FILENAME.md
Ukrainian:    docs/uk/FILENAME.uk.md
German:       docs/de/FILENAME.de.md
Portuguese:   docs/pt/FILENAME.pt.md
Spanish:      docs/es/FILENAME.es.md
```

**Коди мов** відповідають стандарту ISO 639-1:

- `uk` - Ukrainian (українська)
- `de` - German (Deutsch)
- `pt` - Portuguese (Português)
- `es` - Spanish (Español)

---

## Як додати переклад

### 1. Виберіть документ

Оберіть неперекладений документ зі списку пріоритетів вище.

### 2. Створіть файл перекладу

```bash
# Example: Translate ARCHITECTURE_QUICKREF.md to Ukrainian
touch docs/uk/ARCHITECTURE_QUICKREF.uk.md
```

### 3. Додайте перемикач мов

На початку **оригінального англійського документа** додайте:

```markdown
🇬🇧 **English** | [🇺🇦 Українська](uk/FILENAME.uk.md) | [🇩🇪 Deutsch](de/FILENAME.de.md) | [🇵🇹 Português](pt/FILENAME.pt.md) | [🇪🇸 Español](es/FILENAME.es.md)
```

На початку **перекладеного документа** додайте:

```markdown
🇬🇧 [English](../FILENAME.md) | 🇺🇦 **Українська**
```

### 4. Оновіть прогрес

Оновіть відповідний README мови (`docs/{lang}/README.md`), щоб позначити документ як завершений.

### 5. Відправте Pull Request

- Заголовок: `docs: Add [Language] translation for [Document]`
- Приклад: `docs: Add Ukrainian translation for ARCHITECTURE_QUICKREF`

---

## Керівні принципи перекладу

### ✅ РОБІТЬ:

- Перекладайте технічні терміни послідовно (використовуйте глосарій нижче)
- Залишайте приклади коду незмінними (код універсальний)
- Зберігайте всі посилання (оновлюйте шляхи до перекладених версій, коли доступно)
- Підтримуйте ту саму структуру документа
- Використовуйте конвенції рідної мови (наприклад, формати дат, лапки)

### ❌ НЕ РОБІТЬ:

- Не перекладайте імена файлів або шляхи в коді
- Не змінюйте приклади коду чи вивід команд
- Не видаляйте та не пропускайте розділи
- Не перекладайте назви брендів (Promenade, PostgreSQL, Redis)
- Не перекладайте ключові слова програмування (`func`, `type`, `interface`)

---

## Глосарій технічних термінів

Для забезпечення послідовності в перекладах:

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

## Контриб'ютори

Особлива подяка контриб'юторам перекладів:

- 🇺🇦 Ukrainian: [@basilex](https://github.com/basilex) and AI Assistant
- 🇩🇪 German: _Запрошуємо контриб'юторів!_
- 🇵🇹 Portuguese: _Запрошуємо контриб'юторів!_
- 🇪🇸 Spanish: _Запрошуємо контриб'юторів!_

**Хочете долучитися?** Перегляньте [Посібник з внеску](../README.md#contributing) та оберіть документ зі списку пріоритетів вище!

---

## Підтримка

- **Проблеми з документацією**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Питання про переклади**: alexander.vasilenko@gmail.com
- **Спільнота**: Приєднуйтесь до обговорень вашою мовою!

---

**Робимо Promenade доступною для розробників з усього світу** 🌍

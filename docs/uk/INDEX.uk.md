# Індекс Документації Promenade

[🇬🇧 English](../INDEX.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/INDEX.de.md) | [🇵🇹 Português](../pt/INDEX.pt.md) | [🇪🇸 Español](../es/INDEX.es.md)

Ця тека містить вичерпну документацію щодо архітектури додатку Promenade, робочих процесів розробки та найкращих практик.

> **Нове!** Документація тепер доступна кількома мовами. Див. [TRANSLATIONS.md](../TRANSLATIONS.md) для статусу перекладів та рекомендацій щодо внесення змін.

---

## Почніть Тут

### Новачок у Promenade?

1. **[ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md)** - Візуальна діаграма архітектури та огляд компонентів
2. **[ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md)** - Швидкий довідник для розробників
3. **[README.md](../../README.md)** - Головний README проєкту

### Огляд Архітектури

- **[ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md)** - Повний аудит відповідності архітектури (Core vs Модулі)

---

## Основні Концепції

### Архітектура & Дизайн

| Документ                                                | Опис                                     | Коли Читати                     |
| ------------------------------------------------------- | ---------------------------------------- | ------------------------------- |
| [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) | Повна візуальна архітектура з діаграмами | Розуміння структури системи     |
| [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md)       | Звіт про відповідність архітектури       | Перевірка принципів дизайну     |
| [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md) | Швидкий довідник для типових патернів    | Щоденна розробка                |
| [../../internal/CORE.md](../../internal/CORE.md)        | Документація основних компонентів        | Розуміння відповідальності core |

### Модулі

| Документ                                                             | Опис                                 | Коли Читати                   |
| -------------------------------------------------------------------- | ------------------------------------ | ----------------------------- |
| [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md)                    | Повний посібник зі створення модулів | Створення нових модулів       |
| [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.uk.md)                  | Принципи незалежності модулів        | Розуміння меж модулів         |
| [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.uk.md)    | Управління конфігурацією для модулів | Налаштування конфігів модулів |
| [../../internal/modules/README.md](../../internal/modules/README.md) | Структура теки модулів               | Швидкий огляд модулів         |

---

## Технічні Посібники

### База Даних & Персистентність

| Документ                                | Опис                                   | Коли Читати                 |
| --------------------------------------- | -------------------------------------- | --------------------------- |
| [UUID_V7_GUIDE.md](UUID_V7_GUIDE.uk.md) | Використання часово-впорядкованих UUID | Робота з первинними ключами |
| [SOFT_DELETE.md](../SOFT_DELETE.md)     | Патерни soft delete та підводні камені | Реалізація soft delete      |

### Інфраструктура

| Документ                                          | Опис                                 | Коли Читати                     |
| ------------------------------------------------- | ------------------------------------ | ------------------------------- |
| [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md) | Дизайн автоматизованої системи purge | Реалізація політик збереження   |
| [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md)   | Тестування з Redis event bus         | Тестування event-driven функцій |
| [LOGGING.md](../LOGGING.md)                       | Структуроване логування з контекстом | Додавання логування до коду     |

### Безпека & Автентифікація

| Документ                                              | Опис                                              | Коли Читати                                       |
| ----------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- |
| [AUTH_SCHEMA.md](AUTH_SCHEMA.uk.md)                   | Система автентифікації (реєстрація, вхід, JWT)    | Розуміння потоку автентифікації                   |
| [AUTHORIZATION.md](AUTHORIZATION.uk.md)               | Система дозволів RBAC                             | Реалізація авторизації                            |
| [CREDENTIALS.md](../CREDENTIALS.md)                   | Користувачі та ролі за замовчуванням для dev/test | Тестування з попередньо визначеними користувачами |
| [LICENSE_ARCHITECTURE.md](LICENSE_ARCHITECTURE.uk.md) | Система ліцензування модулів з HMAC-SHA256        | Реалізація комерційних модулів                    |

---

## Тестування

| Документ                                                      | Опис                                   | Коли Читати                       |
| ------------------------------------------------------------- | -------------------------------------- | --------------------------------- |
| [TESTING_GUIDE.md](TESTING_GUIDE.uk.md)                       | Повна стратегія тестування             | Написання тестів                  |
| [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.uk.md)     | Налаштування інфраструктури тестування | Налаштування тестового середовища |
| [MOCK_GENERATION_STANDARD.md](../MOCK_GENERATION_STANDARD.md) | Уніфікований підхід до генерації моків | Робота з моками репозиторіїв      |
| [MOCK_STANDARDIZATION.md](../MOCK_STANDARDIZATION.md)         | Підсумок стандартизації моків          | Розуміння уніфікації моків        |
| [../../test/README.md](../../test/README.md)                  | Структура теки тестів                  | Розуміння організації тестів      |

---

## Робочі Процеси Розробки

### Збірка & Деплой

| Документ                                                | Опис                          | Коли Читати              |
| ------------------------------------------------------- | ----------------------------- | ------------------------ |
| [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) | Документація системи Makefile | Використання команд make |
| [../../docker/README.md](../../docker/README.md)        | Налаштування Docker і деплой  | Контейнеризація додатку  |

### Валідація & Якість

| Документ                          | Опис                            | Коли Читати                |
| --------------------------------- | ------------------------------- | -------------------------- |
| [VALIDATION.md](../VALIDATION.md) | Патерни валідації вхідних даних | Додавання правил валідації |

---

## API Документація

### V1 API

- [v1/v1_docs.go](../v1/v1_docs.go) - V1 API документація
- [v1/v1_swagger.yaml](../v1/v1_swagger.yaml) - V1 Swagger специфікація (YAML)
- [v1/v1_swagger.json](../v1/v1_swagger.json) - V1 Swagger специфікація (JSON)

### V2 API

- [v2/v2_docs.go](../v2/v2_docs.go) - V2 API документація
- [v2/v2_swagger.yaml](../v2/v2_swagger.yaml) - V2 Swagger специфікація (YAML)
- [v2/v2_swagger.json](../v2/v2_swagger.json) - V2 Swagger специфікація (JSON)

---

## За Темами

### Основна Архітектура

- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Візуальний огляд
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md) - Огляд відповідності
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md) - Швидкий довідник
- [../../internal/CORE.md](../../internal/CORE.md) - Компоненти Core

### Система Модулів

- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md) - Посібник з розробки
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.uk.md) - Принципи незалежності
- [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.uk.md) - Конфігурація
- [../../internal/modules/README.md](../../internal/modules/README.md) - Індекс модулів

### Управління Даними

- [UUID_V7_GUIDE.md](UUID_V7_GUIDE.uk.md) - Первинні ключі
- [SOFT_DELETE.md](../SOFT_DELETE.md) - Патерни soft delete
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md) - Автоматизована purge

### Безпека & Автентифікація

- [AUTH_SCHEMA.md](AUTH_SCHEMA.uk.md) - Автентифікація
- [AUTHORIZATION.md](AUTHORIZATION.uk.md) - Система RBAC
- [CREDENTIALS.md](../CREDENTIALS.md) - Обробка облікових даних

### Тестування & Якість

- [TESTING_GUIDE.md](TESTING_GUIDE.uk.md) - Стратегія тестування
- [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.uk.md) - Налаштування тестів
- [VALIDATION.md](../VALIDATION.md) - Валідація вхідних даних
- [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Тестування event bus

### Інфраструктура

- [LOGGING.md](../LOGGING.md) - Структуроване логування
- [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) - Система збірки
- [../../docker/README.md](../../docker/README.md) - Налаштування Docker

---

## Шляхи Навчання

### Шлях 1: Розуміння Системи (Новий Розробник)

1. [README.md](../../README.md) - Огляд проєкту
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Архітектура системи
3. [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md) - Типові патерни
4. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md) - Створення функцій
5. [TESTING_GUIDE.md](TESTING_GUIDE.uk.md) - Тестування вашого коду

### Шлях 2: Створення Нового Модуля

1. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md) - Посібник зі створення модуля
2. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.uk.md) - Принципи дизайну
3. [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.uk.md) - Конфігурація
4. [UUID_V7_GUIDE.md](UUID_V7_GUIDE.uk.md) - Первинні ключі
5. [SOFT_DELETE.md](../SOFT_DELETE.md) - Якщо використовуєте soft delete
6. [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md) - Якщо реалізуєте purge

### Шлях 3: Огляд Архітектури (Техлід)

1. [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md) - Аналіз поточного стану
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Візуальні діаграми
3. [../../internal/CORE.md](../../internal/CORE.md) - Межі Core
4. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.uk.md) - Ізоляція модулів

### Шлях 4: Реалізація Безпеки

1. [AUTH_SCHEMA.md](AUTH_SCHEMA.uk.md) - Потоки автентифікації
2. [AUTHORIZATION.md](AUTHORIZATION.uk.md) - Дозволи RBAC
3. [CREDENTIALS.md](../CREDENTIALS.md) - Безпека облікових даних

### Шлях 5: Тестування & Якість

1. [TESTING_GUIDE.md](TESTING_GUIDE.uk.md) - Стратегія тестування
2. [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.uk.md) - Налаштування тестів
3. [VALIDATION.md](../VALIDATION.md) - Валідація вхідних даних
4. [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Тестування подій

---

## Швидкий Пошук

### Як мені...

**...створити новий модуль?**
→ [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md)

**...додати конфігурацію до мого модуля?**
→ [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.uk.md)

**...реалізувати soft delete?**
→ [SOFT_DELETE.md](../SOFT_DELETE.md)

**...додати політики збереження?**
→ [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md)

**...правильно використовувати UUID?**
→ [UUID_V7_GUIDE.md](UUID_V7_GUIDE.uk.md)

**...реалізувати автентифікацію?**
→ [AUTH_SCHEMA.md](AUTH_SCHEMA.uk.md)

**...додати дозволи RBAC?**
→ [AUTHORIZATION.md](AUTHORIZATION.uk.md)

**...писати тести?**
→ [TESTING_GUIDE.md](TESTING_GUIDE.uk.md)

**...додати логування?**
→ [LOGGING.md](../LOGGING.md)

**...валідувати вхідні дані?**
→ [VALIDATION.md](../VALIDATION.md)

**...зрозуміти архітектуру?**
→ [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md)

**...перевірити чи мій код відповідає принципам?**
→ [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md)

**...отримати швидкий довідник?**
→ [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md)

---

## Останні Оновлення

### 22 Грудня 2025

- Створено вичерпну архітектурну документацію:
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md) - Повний огляд відповідності
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Візуальні діаграми
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.uk.md) - Швидкий довідник
- Оновлено документацію системи purge:
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md) - Підхід на основі реєстру
- Перевірено незалежність модулів (видалено 15,000+ рядків з core)

---

## **Contributing** Внесок

Коли додаєте нову документацію:

1. **Виберіть правильний тип:**

   - `ARCHITECTURE_*.md` - Архітектура та патерни дизайну
   - `MODULE_*.md` - Документація системи модулів
   - `*_GUIDE.md` - Посібники та навчальні матеріали
   - `*_SCHEMA.md` - Схеми даних та структури
   - `README.md` - Огляди тек

2. **Оновіть цей індекс:**

   - Додайте до відповідного розділу
   - Оновіть "Останні Оновлення"
   - Додайте до "Швидкий Пошук", якщо застосовно

3. **Перехресні посилання:**

   - Посилання на пов'язані документи
   - Оновіть пов'язані документи зворотними посиланнями

4. **Підтримуйте актуальність:**
   - Оновлюйте при зміні архітектури
   - Архівуйте застарілі документи з префіксом `DEPRECATED_`

---

## Підтримка

- **Питання про архітектуру?** → Читайте [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md)
- **Питання про модулі?** → Читайте [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md)
- **Питання про тестування?** → Читайте [TESTING_GUIDE.md](TESTING_GUIDE.uk.md)
- **Інші питання?** → Перевірте цей індекс або головний [README.md](../../README.md)

---

**Версія Документації:** 2.0
**Останнє Оновлення:** 22 Грудня 2025
**Підтримується:** Командою Розробки Promenade

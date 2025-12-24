# Верифікація незалежності модулів

[🇬🇧 English](MODULE_INDEPENDENCE.uk.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/MODULE_INDEPENDENCE.de.md) | [🇵🇹 Português](../pt/MODULE_INDEPENDENCE.pt.md) | [🇪🇸 Español](../es/MODULE_INDEPENDENCE.es.md)

## Структура модуля Posts

Модуль `posts` демонструє повну незалежність від основної системи, реалізуючи власний повний стек Clean Architecture:

```
internal/modules/posts/
├── domain/
│   ├── entity/          # Сутність Post + доменні помилки
│   └── repository/      # Інтерфейс репозиторію
├── usecase/             # Рівень бізнес-логіки
├── adapter/
│   ├── http/           # HTTP обробники та DTO
│   └── repository/     # Реалізація бази даних
├── tests/              # Тести специфічні для модуля
├── module.go           # Інтеграція модуля
└── register.go         # Авто-реєстрація
```

## Аналіз незалежності

### Без залежностей від Core

Перевірено пошуком: `grep -r "github.com/basilex/promenade/internal/(domain|usecase|adapter)" internal/modules/posts/`

**Результат**: НУЛЬ співпадінь - модуль не імпортує жодних внутрішніх пакетів core.

### Тільки спільні залежності

Модуль імпортує лише:

- `pkg/*` - Спільні утиліти (logger, uuidv7, pagination, bus, response, module SDK)
- `github.com/gin-gonic/gin` - HTTP фреймворк
- `github.com/jmoiron/sqlx` - Бібліотека бази даних
- Пакети стандартної бібліотеки

### Повний вертикальний зріз

Кожен рівень реалізований всередині модуля:

| Рівень                     | Розташування                           | Залежності                  |
| -------------------------- | -------------------------------------- | --------------------------- |
| **Доменна сутність**       | `domain/entity/post.go`                | Тільки `pkg/uuidv7`         |
| **Доменний репозиторій**   | `domain/repository/post_repository.go` | Доменна сутність, pkg       |
| **Випадок використання**   | `usecase/post_usecase.go`              | Тільки Domain               |
| **Реалізація репозиторію** | `adapter/repository/postgres/`         | Domain, pkg/database        |
| **HTTP обробник**          | `adapter/http/handler/`                | Use case, DTO, pkg/response |
| **DTO**                    | `adapter/http/dto/`                    | Доменна сутність            |

### Переваги ізоляції модуля

1. **Незалежна розробка**: Можна розробляти/тестувати ізольовано
2. **Повторне використання**: Можна скопіювати в інший проєкт з pkg/
3. **Без критичних змін**: Рефакторинг core не впливає на модуль
4. **Чіткі межі**: Всі залежності явні та мінімальні
5. **Легке тестування**: Мокати потрібно тільки доменні інтерфейси, а не core сервіси

## Створення нових модулів

Щоб створити новий незалежний модуль:

1. Створити структуру директорій:

   ```
   internal/modules/{name}/
   ├── domain/entity/
   ├── domain/repository/
   ├── usecase/
   ├── adapter/http/handler/
   ├── adapter/http/dto/
   ├── adapter/repository/postgres/
   ├── tests/
   ├── module.go
   └── register.go
   ```

2. Скопіювати реалізацію з core або написати з нуля
3. Оновити всі імпорти, щоб вони вказували на шляхи модуля
4. Визначити специфічні для модуля помилки в `domain/entity/errors.go`
5. Реалізувати інтерфейс `module.Module` в `module.go`
6. Авто-реєструвати в `register.go` за допомогою `init()`

## Команда верифікації

```bash
# Перевірити на будь-які внутрішні залежності core
grep -r "github.com/basilex/promenade/internal/\(domain\|usecase\|adapter\)" \
  internal/modules/posts/ || echo "✅ Модуль незалежний"
```

## Поточний стан

- **posts**: Повністю незалежний, повна Clean Architecture
- **comments**: Потребує рефакторингу (наразі обгортка)
- **warehouse**: Потребує рефакторингу (наразі обгортка)

## Наступні кроки

1. Рефакторити модуль `comments` за зразком `posts`
2. Рефакторити модуль `warehouse`
3. Додати інтеграційні тести в директорії `tests/`
4. Документувати специфічні для модуля API
5. Створити посібник з розробки модулів з шаблонами

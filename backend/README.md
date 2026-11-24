# Mentor backend

REST API платформы «Куратор». Go, PostgreSQL (pgx), стандартный `net/http` роутер, миграции goose.
Контракт API: [`api/openapi.yaml`](api/openapi.yaml) (v1.2.0), базовый путь `/api/v1`.

## Запуск

```bash
cp .env.example .env   # заполнить DB_* и JWT_SECRET (не короче 32 байт: openssl rand -hex 32)
make db-up             # PostgreSQL в docker
make migrate-up        # миграции
make run
```

Проверка: `GET /health` → `{"status":"ok"}`.

## Тесты

```bash
make test
```

Интеграционные тесты репозиториев ходят в настоящую базу и запускаются, только если задана
`TEST_DATABASE_URL` (база с накатанными миграциями; тесты очищают таблицы, поэтому нужна отдельная тестовая база):

```bash
TEST_DATABASE_URL=postgres://user:pass@127.0.0.1:5432/mentor_test?sslmode=disable make test
```

Без переменной они пропускаются, юнит-тесты выполняются всегда.

## Модули

| Пакет | Что делает | Ручки |
|---|---|---|
| `internal/auth` | регистрация, вход, JWT (ручки ещё не подключены в `app.go`) | `/auth/*` |
| `internal/requests` | заявки на менторство: матрица статусов v1.2, лимиты, создание пары | `POST /requests`, `GET /requests/my`, `PATCH /requests/{id}/status`, `GET/POST /curator/requests` |
| `internal/catalog` | каталог наставников с фильтрами, карточка, отзывы, словарь навыков | `GET /mentors`, `GET /mentors/{id}`, `GET /mentors/{id}/reviews`, `GET /skills` |
| `internal/curator` | панель куратора: «светофор» пар, статус пары, сообщение паре, очереди помощи и верификации | `/curator/workspaces*`, `/curator/profiles*` |
| `internal/chat` | чат внутри пары | `GET/POST /workspaces/{id}/messages`, `POST /workspaces/{id}/messages/read` |
| `internal/faq` | FAQ-бот по ключевым словам, база в `internal/faq/faq.json` | `GET /faq`, `POST /faq/ask` |
| `internal/notifications` | запись уведомлений (модули пишут их в своих транзакциях) | — |

### Общие договорённости

- **Авторизация** — `transport/rest/middleware`: `auth.Require(handler, roles...)` проверяет Bearer-токен и роль,
  `auth.Optional` — для публичных ручек, которые авторизованным отвечают иначе (каталог). Пользователь
  достаётся через `middleware.PrincipalFrom(ctx)`.
- **Ошибки** — формат контракта `{"error_code": "...", "message": "...", "details": [...]}`:
  `httpx.SendAPIError`, `httpx.SendBadRequest`; коды — константы `httpx.Code*`.
- **Пагинация** — `httpx.ParsePage(query, defaultLimit, maxLimit)`, ответ `{"total", "page", "items"}`.
- **Транзакции** — правила, которые нельзя нарушить параллельными запросами (лимит 3 заявок ученика,
  свободные места наставника при подтверждении заявки или возврате пары в работу), проверяются под
  `SELECT ... FOR UPDATE` в одной транзакции с изменением.

### Точки интеграции

- **Процент совместимости.** Каталог принимает `catalog.Matcher` вторым аргументом `catalog.NewService`.
  Пока модуль подбора не подключён (`nil`), карточки идут без `match`, а сортировка `match` заменяется на `rating`.
- **Системные сообщения в чат** (например, «Встреча назначена») — `chat.InsertSystemMessage(ctx, tx, workspaceID, text)`.

## Миграции

`migrations/00001–00003` — первая версия схемы, `00004–00010` приводят её к ER-диаграмме от 16.11,
`00011` добавляет `requests.suggested_by`. Этого поля нет в ER-диаграмме: оно отличает заявки,
которые создал куратор (они не входят в лимит 3 активных заявок ученика).

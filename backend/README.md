# Mentor backend

REST API платформы «Куратор». Go, PostgreSQL (pgx), стандартный `net/http` роутер, миграции goose.
Контракт API: [`api/openapi.yaml`](api/openapi.yaml) (v1.2.0), базовый путь `/api/v1`.

## Запуск

```bash
cp .env.example .env   # заполнить DB_*, JWT_SECRET и OTP_SECRET (не короче 32 байт: openssl rand -hex 32)
make db-up             # PostgreSQL в docker
make migrate-up        # миграции
make run
```

Проверка: `GET /health` → `{"status":"ok"}`.

- **Коды и сессии.** Коды подтверждения почты и сброса пароля (`CODE_*`) и refresh-токены (`REFRESH_*`)
  пока хранятся в памяти процесса (`internal/platform/memory`): после перезапуска сервера нужно войти заново.
  По контракту в проде это Redis — достаточно реализовать интерфейсы из `internal/auth`.
- **Почта.** Письма с кодами пока не отправляются, а пишутся в лог сервера (`platform/mailer.LogMailer`).
- **Загруженные файлы** (документы для верификации) лежат в `UPLOADS_DIR` и раздаются по `GET /uploads/...`;
  `UPLOADS_BASE_URL` — адрес, по которому их увидит фронтенд.

## Тесты

```bash
make test
```

Интеграционные тесты репозиториев ходят в настоящую базу и запускаются, только если задана
`TEST_DATABASE_URL` или `TEST_DB_DSN` (база с накатанными миграциями; тесты очищают таблицы, поэтому нужна отдельная тестовая база):

```bash
TEST_DATABASE_URL=postgres://user:pass@127.0.0.1:5432/mentor_test?sslmode=disable make test
```

Без переменной они пропускаются, юнит-тесты выполняются всегда.

## Модули

| Пакет | Что делает | Ручки |
|---|---|---|
| `internal/auth` | регистрация с кодом из письма, вход, refresh-токены, восстановление пароля | `/auth/register`, `/auth/verify-email`, `/auth/resend-code`, `/auth/login`, `/auth/refresh`, `/auth/password/*` |
| `internal/account` | данные аккаунта | `GET /users/me` |
| `internal/profile` | профиль: создание при регистрации, анкета по шагам с проверками по ролям, материалы для верификации (ручки — в `profile/profilehttp`) | `GET/PATCH /profiles/me`, `POST /profiles/me/verification-assets`, `DELETE /profiles/me/verification-assets/{assetId}` |
| `internal/workspaces` | список пар пользователя (для чата) | `GET /workspaces` |
| `internal/matching` | процент совместимости по формуле из контракта | — (используют каталог и заявки) |
| `internal/requests` | заявки на менторство: матрица статусов v1.2, лимиты, создание пары | `POST /requests`, `GET /requests/my`, `PATCH /requests/{id}/status`, `GET/POST /curator/requests` |
| `internal/catalog` | каталог наставников с фильтрами, карточка, отзывы, словарь навыков | `GET /mentors`, `GET /mentors/{id}`, `GET /mentors/{id}/reviews`, `GET /skills` |
| `internal/curator` | панель куратора: «светофор» пар, статус пары, сообщение паре, очереди помощи и верификации | `/curator/workspaces*`, `/curator/profiles*` |
| `internal/chat` | чат внутри пары | `GET/POST /workspaces/{id}/messages`, `POST /workspaces/{id}/messages/read` |
| `internal/notifications` | запись уведомлений (модули пишут их в своих транзакциях) | — |

### Общие договорённости

- **Авторизация** — `transport/rest/middleware`: `auth.Require(handler, roles...)` проверяет Bearer-токен и роль,
  `auth.Optional` — для публичных ручек, которые авторизованным отвечают иначе (каталог). Пользователь
  достаётся через `middleware.PrincipalFrom(ctx)`.
- **Ошибки** — формат контракта `{"error_code": "...", "message": "...", "details": [...]}`:
  `httpx.SendError`, `httpx.SendBadRequest`; коды — константы `httpx.Err*`.
- **Пагинация** — `httpx.ParsePage(query, defaultLimit, maxLimit)`, ответ `{"total", "page", "items"}`.
- **Транзакции** — правила, которые нельзя нарушить параллельными запросами (лимит 3 заявок ученика,
  свободные места наставника при подтверждении заявки или возврате пары в работу), проверяются под
  `SELECT ... FOR UPDATE` в одной транзакции с изменением.

### Точки интеграции

- **Процент совместимости** считает `internal/matching`: навыки 50%, цели 20%, время 20%, стиль 10%.
  Каталог и заявки получают его через интерфейсы, поэтому `GET /mentors/recommendations` можно собрать на этом же расчёте.
- **Системные сообщения в чат** (например, «Встреча назначена») — `chat.InsertSystemMessage(ctx, tx, workspaceID, text)`.

## Миграции

`migrations/00001–00003` — первая версия схемы, `00004–00010` приводят её к ER-диаграмме от 16.11,
`00011` добавляет `requests.suggested_by`. Этого поля нет в ER-диаграмме: оно отличает заявки,
которые создал куратор (они не входят в лимит 3 активных заявок ученика). `00012` исправляет
`profiles.availability_slots`: по контракту это объект по дням недели, а не массив.

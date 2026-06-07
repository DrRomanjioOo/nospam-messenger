# NoSpam Messenger — Backend

Go backend for a single global chat with sync spam rules, rate limiting, and async OpenRouter moderation.

## Stack

- **HTTP:** `net/http` (Go 1.23+)
- **DB:** PostgreSQL (parameterized queries via `pgx`)
- **Sessions & rate limits:** Redis
- **Real-time:** WebSocket (`/ws`)
- **AI worker:** in-process goroutine pool + queue

## Quick start

```bash
cp .env.example .env
# optional: set OPENROUTER_API_KEY in .env or docker-compose

docker compose up --build
```

- API: http://localhost:8080
- Swagger UI: http://localhost:8081 (openapi: http://localhost:8081/openapi.yaml)
- Health: http://localhost:8080/health

## API overview

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | — | Register (sets `session_id` cookie) |
| POST | `/auth/login` | — | Login |
| GET | `/messages` | cookie | History, newest first (`before_id`, `limit≤50`) |
| POST | `/messages` | cookie | Send message |
| DELETE | `/messages/{id}` | cookie | Soft-delete own message |
| GET | `/ws` | cookie | WebSocket chat + presence |
| GET | `/health` | — | Postgres + Redis health |

Full contract: [`openapi.yaml`](openapi.yaml).

## Spam pipeline

1. **Sync (before DB write):** stop-words, empty/whitespace, meaningless text; **1 msg/sec** per user (Redis).
2. **Persist & broadcast** if sync checks pass.
3. **Async AI:** worker pool calls OpenRouter; on spam → soft delete (`deleted_by_ai`, empty `content`) + `message.updated` event; audit row in `spam_audit`.
4. **OpenRouter errors:** message stays; job retried with backoff.

Set `OPENROUTER_SPAM_PROMPT` when ready (placeholder used until then).

## Тесты

### Быстрые unit-тесты (без Postgres)

Нужен установленный Go 1.23+:

```bash
cd backend
go mod download
go test ./... -count=1
```

Покрытие по пакетам: `domain`, `spam`, `service`, `handler`, `middleware`, `ws`, `worker`, `repository` (Redis через miniredis).

### Все тесты, включая Postgres (integration)

Поднять инфраструктуру и прогнать тесты в контейнере Go:

```bash
cd backend
docker compose up -d postgres redis
docker compose run --rm test
```

Или локально с Go:

```bash
docker compose up -d postgres redis
export DATABASE_URL=postgres://messenger:messenger@localhost:5432/messenger?sslmode=disable
export REDIS_URL=redis://localhost:6379/0
export SESSION_SECRET=test-secret
go test ./... -count=1 -tags=integration
```

### Поднять всю систему

```bash
cd backend
docker compose up --build
```

Проверка API:

```bash
curl http://localhost:8080/health
```

## Local development (without Docker)

```bash
# start postgres + redis (or use compose for only infra)
export $(cat .env | xargs)  # or set vars manually
go run ./cmd/api
```

## Project layout

```
cmd/api/          entrypoint
internal/
  handler/        HTTP + WS handlers
  service/        business logic
  repository/     postgres + redis
  spam/           rules + openrouter client
  worker/         AI worker pool
  ws/             hub, presence
migrate/          goose SQL migrations
openapi.yaml
docker-compose.yml
```

# NoSpam Messenger

Веб-мессенджер с одним общим чатом, антиспам-фильтрами и ИИ-модерацией сообщений в реальном времени.

## Функционал

- **Регистрация и вход** — сессии в cookie, логин 4–16 символов (латиница/цифры).
- **Общий чат** — отправка и просмотр сообщений, история с подгрузкой при прокрутке вверх.
- **Real-time** — новые сообщения, удаления и системные уведомления доставляются по WebSocket.
- **Антиспам (синхронно)** — стоп-слова, пустые/бессмысленные сообщения, лимит 1 сообщение в секунду на пользователя.
- **ИИ-модерация (асинхронно)** — проверка через [OpenRouter](https://openrouter.ai/); спам мягко удаляется с пометкой «Сообщение удалено модерацией ИИ».
- **Удаление своих сообщений** — мягкое удаление автором.
- **Присутствие в чате** — уведомления «username подключился к чату» / «username вышел из чата».

## Стек

| Часть | Технологии |
|-------|------------|
| Backend | Go 1.23, PostgreSQL, Redis, WebSocket |
| Frontend | HTML, CSS, JavaScript (без сборщика) |
| AI | OpenRouter API (опционально) |

## Зависимости

Для локального запуска через Docker достаточно:

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

Для запуска фронтенда отдельно:

- [Node.js](https://nodejs.org/) 18+ (нужен `npx` для статического сервера на порту 3000)

Опционально для разработки бэкенда без Docker:

- Go 1.23+
- PostgreSQL 16
- Redis 7

## Локальный запуск (Docker)

### 1. Бэкенд и инфраструктура

```bash
cd backend
cp .env.example .env
```

При необходимости укажите в `.env` ключ OpenRouter для ИИ-модерации:

```env
OPENROUTER_API_KEY=sk-or-v1-...
```

Запуск API, PostgreSQL, Redis и Swagger UI:

```bash
docker compose up --build
```

Сервисы после старта:

| Сервис | URL |
|--------|-----|
| API | http://localhost:8080 |
| Health-check | http://localhost:8080/health |
| Swagger UI | http://localhost:8081 |
| PostgreSQL | `localhost:5432` (user/pass/db: `messenger`) |
| Redis | `localhost:6379` |

Остановка:

```bash
docker compose down
```

### 2. Фронтенд

В отдельном терминале:

```bash
cd frontend
npm start
```

Откройте в браузере: **http://localhost:3000**

Фронтенд обращается к API на `http://localhost:8080` (настраивается в `frontend/src/index.js`).

### 3. Проверка

```bash
curl http://localhost:8080/health
```

Зарегистрируйтесь в UI, отправьте сообщение. Для проверки ИИ-модерации нужен валидный `OPENROUTER_API_KEY` в `backend/.env` и перезапуск контейнера `api`.

## Переменные окружения

Основные переменные — в [`backend/.env.example`](backend/.env.example):

| Переменная | Описание |
|------------|----------|
| `DATABASE_URL` | Подключение к PostgreSQL |
| `REDIS_URL` | Подключение к Redis |
| `SESSION_SECRET` | Секрет для сессий (обязательно сменить в production) |
| `CORS_ORIGIN` | Разрешённые origin фронтенда |
| `OPENROUTER_API_KEY` | Ключ OpenRouter; без него ИИ-модерация отключена |
| `OPENROUTER_MODEL` | Модель (по умолчанию `deepseek/deepseek-chat`) |

## Структура репозитория

```
nospam-messenger/
├── backend/          # Go API, WebSocket, миграции, docker-compose
│   ├── cmd/api/      # Точка входа
│   ├── internal/     # handlers, services, spam, worker, ws
│   └── openapi.yaml  # Описание HTTP/WS API
└── frontend/         # Статический UI чата
```

Подробнее о бэкенде, тестах и API: [`backend/README.md`](backend/README.md).

## Тесты

```bash
cd backend
docker compose run --rm test
```

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: messenger
      POSTGRES_PASSWORD: messenger
      POSTGRES_DB: messenger
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U messenger -d messenger"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 10

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      HTTP_ADDR: ":8080"
      DATABASE_URL: postgres://messenger:messenger@postgres:5432/messenger?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      SESSION_SECRET: dev-change-me-in-production
      CORS_ORIGIN: http://localhost:3000,http://127.0.0.1:3000
      COOKIE_SECURE: "false"
      OPENROUTER_API_KEY: ${OPENROUTER_API_KEY:-}
      OPENROUTER_MODEL: deepseek/deepseek-chat
      OPENROUTER_SPAM_PROMPT: ""
      AI_WORKER_COUNT: "4"
      AI_QUEUE_SIZE: "256"
      AI_RETRY_DELAY_SEC: "30"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  swagger:
    image: swaggerapi/swagger-ui:v5.18.2
    ports:
      - "8081:8080"
    environment:
      SWAGGER_JSON_URL: /openapi.yaml
    volumes:
      - ./openapi.yaml:/usr/share/nginx/html/openapi.yaml:ro

  test:
    image: golang:1.23-alpine
    working_dir: /app
    volumes:
      - .:/app
    environment:
      DATABASE_URL: postgres://messenger:messenger@postgres:5432/messenger?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      SESSION_SECRET: test-secret
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: sh -c "go mod download && go test ./... -count=1"

  test-integration:
    image: golang:1.23-alpine
    working_dir: /app
    volumes:
      - .:/app
    environment:
      DATABASE_URL: postgres://messenger:messenger@postgres:5432/messenger?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      SESSION_SECRET: test-secret
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: sh -c "go mod download && go test ./... -count=1 -tags=integration"

volumes:
  pgdata:

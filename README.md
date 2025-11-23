# Сервис назначения ревьюеров

Микросервис автоматически назначает ревьюеров на PR, управляет командами и активностью пользователей, отдает статистику и массово деактивирует команду с безопасной переассигнацией.

## Запуск
- Требования: Go 1.23+, Docker + docker-compose.
- Переменные окружения читаются из `config/.env` (опционально). См. типы в `config/models.go`. Основные:
  - `POSTGRES_HOST/PORT/USER/PASSWORD/DB_NAME/SSL_MODE`
  - `SERVER_HOST/SERVER_PORT`
  - таймауты: `HTTP_REQUEST_TIMEOUT`, `POSTGRES_QUERY_TIMEOUT`, `POSTGRES_MIGRATE_TIMEOUT`, `SERVER_SHUTDOWN_TIMEOUT`

Быстрый старт (применит миграции через goose при старте сервиса):

```bash
docker-compose up --build
```

Локально без контейнеров (Postgres должен быть доступен по env):

```bash
make generate   # oapi-codegen из openapi.yml
go run ./cmd
```

Проверки:

```bash
make ci         # fmt + lint + generate + test + diff
make test       # go test ./...

# линты/форматирование
make lint       # golangci-lint run (конфиг .golangci.yml)
make fmt        # gofumpt -w
```

## API и документация
- OpenAPI спецификация: `openapi.yml` (а также вшита в бинарник через oapi-codegen). Импортируйте в Swagger UI или постман.
- Основные эндпоинты (см. спецификацию для полей/кодов):
  - `POST /team/add` — создать команду и участников.
  - `GET /team` — получить команду по имени.
  - `POST /pull-request/create` — создать PR, автоназначение до 2 ревьюеров из команды автора.
  - `POST /pull-request/merge` — идемпотентный merge.
  - `POST /pull-request/reassign` — переассайн одного ревьюера.
  - `GET /users/get-review` — список PR, где пользователь ревьюер.
  - `POST /users/set-is-active` — включить/выключить пользователя.
  - `POST /deactivate/team` — массовая деактивация команды и безопасная переассигнация.
  - `GET /stats` и `GET /stats/summary` — агрегированная статистика.
  - `GET /healthz` — health-check.

Примеры (curl):

```bash
# создать команду
curl -X POST http://localhost:8080/team/add -H "Content-Type: application/json" \
  -d '{"team_name":"backend","members":[{"id":"u1","username":"Alice","is_active":true}]}'

# создать PR
curl -X POST http://localhost:8080/pull-request/create -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr1","pull_request_name":"Init","author_id":"u1"}'

# merge PR (идемпотентно)
curl -X POST http://localhost:8080/pull-request/merge -H "Content-Type: application/json" \
  -d '{"pull_request_id":"pr1"}'

# статистика c фильтром
curl "http://localhost:8080/stats/summary?limit=5"
```

## Допущения
- Выбор ревьюеров и переассайны выполняются случайно, при недоступности crypto/rand используется детерминированный fallback (срез кандидатов).
- Если нет активных кандидатов при деактивации команды, ревьюер удаляется с записью в историю.
- Таймауты запросов к БД/HTTP задаются конфигом и применяются на уровне сервисов и репозитория.
- Миграции через goose выполняются при старте; при ошибке сервис не поднимается.

## Полезное
- Логи структурированы (zap), включают request_id и категории (`service.*`, `repo.postgres`).
- Make таргеты: `generate`, `test`, `lint`, `fmt`, `ci`.
- Линтер: `golangci-lint` с gofumpt, goconst, gocyclo, misspell, revive (см. `.golangci.yml`).

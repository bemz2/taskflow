# Taskflow

## Обзор проекта

Taskflow это Go backend для регистрации пользователей, аутентификации и управления персональными задачами.

Сервис предоставляет HTTP API на Echo, хранит пользователей и задачи в PostgreSQL и использует Redis как кэш для чтения отдельных задач.

Основной стек:

- Go 1.25+ (целевая версия модуля) / локально проверено на Go 1.26
- Echo v4
- PostgreSQL через `pgx/v5`
- Redis через `go-redis/v9`
- Собственный HS256 token service в JWT-подобном формате
- Docker Compose
- Swaggo (`swag`, `echo-swagger`)

## Обзор архитектуры

Проект использует layered architecture:

- `internal/http` отвечает за transport-уровень, binding, ответы и middleware
- `internal/service` содержит use case-логику, аутентификацию, выпуск/проверку токенов и работу с кэшем
- `internal/repository` содержит логику доступа к PostgreSQL
- `internal/domain` содержит сущности, инварианты и бизнес-правила
- `internal/application` собирает зависимости и настраивает HTTP сервер

Схема:

```text
HTTP Client
    |
    v
Echo Router + Middleware
    |
    v
Handlers (internal/http/handler)
    |
    v
Services (internal/service)
    |              \
    |               \-- Redis cache (cache-aside)
    v
Repositories (internal/repository)
    |
    v
PostgreSQL
```

Используемые паттерны:

- Layered architecture
- Dependency injection через `internal/application.Container`
- Cache-aside для `TaskService.GetTask`
- Repository pattern для PostgreSQL
- Graceful shutdown для HTTP процесса

## Технологический стек

| Компонент | Детали |
| --- | --- |
| Go | `go 1.25.0` в `go.mod` |
| HTTP framework | Echo v4 |
| База данных | PostgreSQL 16 |
| Кэш | Redis 7 |
| Аутентификация | Собственный HS256 token service + bearer middleware |
| SQL builder | `github.com/Masterminds/squirrel` |
| Контейнеры | Docker Compose |
| API документация | Swaggo + Swagger UI |

## Возможности

- Регистрация пользователя
- Логин пользователя
- Защищённый `GET /me`
- Создание задачи
- Получение задачи по ID
- Список задач с пагинацией, фильтрацией, поиском и сортировкой
- Смена статуса задачи
- Удаление задачи
- Хеширование паролей через bcrypt
- Redis-кэш для чтения отдельных задач
- Автоматическое создание dev-пользователя при старте
- Graceful shutdown по `SIGINT` / `SIGTERM`
- Запуск SQL миграций через `cmd/postgres-migrations`

## Быстрый старт

### Требования

- Go 1.25 или новее
- Docker и Docker Compose
- Установленный `swag` (`go install github.com/swaggo/swag/cmd/swag@latest`)

### Рекомендуемый сценарий через Make

```bash
make setup
make run
```

Что делает `make setup`:

- очищает директорию `bin`
- создаёт `.env` из `.env.sample`, если `.env` ещё не существует
- скачивает Go-зависимости
- устанавливает `mockery`
- поднимает `postgres` и `redis` через Docker Compose
- выполняет миграции

`make run` запускает HTTP API.

API по sample-конфигу будет доступен на:

```text
http://localhost:1323
```

### Дополнительные команды Make

```bash
make infra-up
make infra-down
make migrate
make swagger
make test
make build
```

### Worker

Каталог `cmd/taskflow-worker` существует, но сейчас это заглушка без рабочего `main()`. Команда `make run-worker` завершится ошибкой до тех пор, пока worker не будет реализован.

### Swagger

Генерация Swagger-артефактов:

```bash
make swagger
```

После запуска API Swagger UI доступен по адресу:

```text
http://localhost:1323/swagger/index.html
```

## Переменные окружения

| Переменная | Обязательна | Значение по умолчанию | Назначение |
| --- | --- | --- | --- |
| `PUBLIC_SERVER_PORT` | Нет | `1323` | Порт HTTP сервера |
| `DB_ADAPTER` | Да | Нет | Схема подключения PostgreSQL, обычно `postgres` |
| `POSTGRES_HOST` | Да | Нет | Хост PostgreSQL |
| `POSTGRES_PORT` | Да | Нет | Порт PostgreSQL |
| `POSTGRES_DB` | Да | Нет | Имя базы данных |
| `POSTGRES_USER` | Да | Нет | Пользователь PostgreSQL |
| `POSTGRES_PASSWORD` | Да | Нет | Пароль PostgreSQL |
| `POSTGRES_SSLMODE` | Да | Нет | Режим SSL для PostgreSQL |
| `REDIS_HOST` | Да | Нет | Хост Redis |
| `REDIS_PORT` | Нет | `6379` | Порт Redis |
| `REDIS_PASSWORD` | Нет | пусто | Пароль Redis |
| `REDIS_DB` | Нет | `0` | Индекс Redis DB |
| `JWT_SECRET` | Нет | `taskflow-dev-secret` | Секрет подписи токена |
| `JWT_EXPIRATION_HOURS` | Нет | `24` | TTL токена в часах |

Примечания:

- `POSTGRES_CONNECT_TIMEOUT` встречался в старых sample-файлах, но приложение его не читает.
- Корневой `.env.sample` и `samples/.env.sample` синхронизированы с `internal/config.go`.

## API Endpoints

| Method | Path | Description | Auth Required |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/register` | Создать пользователя и вернуть bearer token | Нет |
| `POST` | `/api/v1/auth/login` | Аутентифицировать пользователя и вернуть bearer token | Нет |
| `POST` | `/api/v1/users` | Создать пользователя без выдачи токена | Нет |
| `GET` | `/api/v1/me` | Получить текущего пользователя | Да |
| `GET` | `/api/v1/tasks` | Получить список задач с фильтрами | Да |
| `POST` | `/api/v1/task` | Создать задачу | Да |
| `GET` | `/api/v1/task/:id` | Получить задачу по ID | Да |
| `PATCH` | `/api/v1/tasks/:id/status` | Изменить статус задачи | Да |
| `DELETE` | `/api/v1/tasks/:id` | Удалить задачу | Да |
| `GET` | `/swagger/*` | Swagger UI и OpenAPI-артефакты | Нет |

## Структура проекта

| Путь | Назначение |
| --- | --- |
| `cmd/taskflow-api` | Основной HTTP API процесс |
| `cmd/postgres-migrations` | CLI для запуска SQL миграций |
| `cmd/taskflow-worker` | Заглушка под worker-процесс |
| `internal/application` | Инициализация приложения, DI-контейнер, запуск сервера |
| `internal/client` | Инициализация PostgreSQL и Redis клиентов |
| `internal/domain` | Сущности и бизнес-инварианты |
| `internal/http/dto` | HTTP request/response модели |
| `internal/http/handler` | Echo handlers |
| `internal/http/middleware` | HTTP middleware |
| `internal/repository` | PostgreSQL repositories |
| `internal/service` | Use cases, токены и кэш |
| `internal/lib/logger` | Абстракция логгера и реализация |
| `migrations/postgres-migrations` | SQL миграции |
| `docs` | Swagger-артефакты и архитектурная документация |
| `samples` | Примеры env и compose файлов |

## Known Issues

- `sort_by` попадает в `ORDER BY` без whitelist в `internal/repository/task/task.go`, это создаёт риск SQL injection.
- В домене используется статус `cancelled`, а в PostgreSQL enum в миграции используется `canceled`. Обновление задачи в отменённое состояние может падать на уровне БД.
- Handlers используют `c.Get("userID").(uuid.UUID)`. Если middleware не отработает или контекст будет повреждён, возможна паника.
- `cmd/taskflow-worker` объявлен в репозитории и Makefile, но рабочей реализации worker нет.
- Таблица `task_analytics` создаётся миграцией, но не используется приложением.

## Возможные улучшения

- Добавить явную валидацию запросов вместо одного `Echo.Bind`
- Ввести структурированные DTO для ошибок вместо raw JSON string
- Добавить интеграционные тесты для PostgreSQL и Redis
- Добавить учёт состояния миграций вместо безусловного запуска всех `*.up.sql`
- Реализовать worker или удалить заглушку
- Ограничить список допустимых колонок для сортировки
- При необходимости совместимости заменить самописный token service на стандартную JWT-библиотеку

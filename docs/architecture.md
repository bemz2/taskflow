# Архитектура

## Обзор

Taskflow использует layered architecture с явными границами между transport-слоем, прикладной логикой, persistence-слоем и доменными правилами.

Основная цель такой структуры: изолировать HTTP, кэш, Kafka и SQL от самих правил работы с пользователями и задачами.

```text
cmd/taskflow-api
    |
    v
internal/application
    |
    v
internal/http -> internal/service -> internal/repository
                     |
                     v
               internal/domain

cmd/taskflow-worker -> Kafka -> task_analytics
```

## Почему выбрана layered architecture

Репозиторий небольшой, но в нём уже есть несколько инфраструктурных подсистем:

- HTTP transport (Echo)
- PostgreSQL persistence
- Redis caching
- Kafka для событий аналитики
- token-based authentication
- startup / shutdown lifecycle

Layered architecture позволяет разделить эти обязанности без лишнего усложнения в стиле full DDD с bounded context, domain events и отдельными aggregate roots.

Для этого проекта такая структура подходит, потому что:

- handlers остаются тонкими и отвечают только за HTTP I/O
- services централизуют use case-логику
- repositories остаются сфокусированными на SQL
- domain-типы держат инварианты рядом с данными, которые они защищают

## Ответственность слоёв

### Domain

`internal/domain` содержит:

- `User`
- `Task`
- `TaskFilter`
- проверки инвариантов и правил перехода состояний

Примеры:

- `Task.ChangeStatus` проверяет допустимость смены статуса
- `Task.Rename` запрещает пустой заголовок
- `NormalizeUserEmail` нормализует email
- `NormalizeStatus` приводит legacy-значение `cancelled` к каноническому `canceled`

Этот слой не должен знать ни про Echo, ни про Redis, ни про PostgreSQL, ни про Kafka.

### Service

`internal/service` содержит прикладные use cases:

- `AuthService`
- `UserService`
- `TaskService`
- `TokenService`
- `AnalyticsPublisher`

Этот слой координирует:

- вызовы repository
- хеширование паролей
- выпуск и разбор токенов
- cache-aside чтение и инвалидацию кэша
- публикацию событий аналитики

Именно service-слой является правильным местом для Redis и Kafka, потому что:

- кэширование это прикладная оптимизация, а не обязанность persistence-слоя
- публикация событий это orchestration-задача use case, а не SQL-операция

### Repository

`internal/repository` содержит PostgreSQL-специфичную логику:

- построение SQL через Squirrel
- выполнение запросов через `pgxpool`
- mapping между storage model и domain model

Repository намеренно не знает о Redis и Kafka. Это сохраняет его узким и предсказуемым:

- repository читает и пишет постоянное состояние
- service решает, нужно ли читать через кэш, инвалидировать кэш или публиковать событие

Если бы кэш и event publishing жили в repository, persistence-слой стал бы зависеть от внешних volatile-систем и его контракт был бы менее прозрачным.

### HTTP

`internal/http` содержит:

- request DTO
- response DTO
- handlers
- middleware

Handlers делают:

- binding запроса
- чтение path/query параметров
- вызов service-слоя
- преобразование domain model в response DTO

Бизнес-правила здесь не должны жить, кроме transport-level проверок.

## Сборка зависимостей

`internal/application.Container` это composition root.

Он создаёт:

- logger
- PostgreSQL pool
- Redis client
- Kafka writer для аналитики
- repositories
- services
- handlers

Также он создаёт dev-пользователя через `UserService.EnsureDevUser`.

```text
Config
  |
  v
Container.Init()
  |-- Postgres pool
  |-- Redis client
  |-- Kafka writer -> AnalyticsPublisher
  |-- TokenService
  |-- UserRepository -> UserService -> UserHandler
  |-- AuthService -> AuthHandler
  `-- TaskRepository -> TaskService(cache) -> TaskHandler(events)
```

## Аутентификация

Аутентификация реализована через собственный HS256 token service в `TokenService`.

Формат токена JWT-подобный:

```text
base64url(header).base64url(payload).base64url(signature)
```

Поля payload:

- `sub`: идентификатор пользователя
- `exp`: timestamp истечения токена

Поток работы:

1. `AuthHandler` принимает credentials.
2. `AuthService` либо создаёт пользователя (`Register`), либо проверяет пароль (`Login`).
3. `TokenService.Issue` подписывает токен через HMAC-SHA256.
4. Клиент передаёт `Authorization: Bearer <token>`.
5. `AuthMiddleware` извлекает bearer token.
6. `TokenService.Parse` проверяет подпись и срок действия.
7. Middleware кладёт `userID` в Echo context.
8. Handlers читают `userID` через безопасный helper `UserIDFromContext`, без panic на type assertion.

```text
Client -> /auth/login
       <- token

Client -> Authorization: Bearer <token>
       -> AuthMiddleware
       -> TokenService.Parse
       -> handler/service
```

## Cache-Aside в TaskService

Redis используется только для чтения одной задачи (`GetTask`).

Текущее поведение:

- `CreateTask`: пишет в PostgreSQL, затем кладёт задачу в Redis
- `GetTask`: сначала проверяет Redis; при miss читает PostgreSQL и заново прогревает кэш
- `ChangeStatus`: обновляет PostgreSQL, затем обновляет кэш
- `UpdateTask`: обновляет PostgreSQL, затем обновляет кэш
- `DeleteTask`: удаляет из PostgreSQL, затем удаляет ключ из Redis
- `ListTasks`: всегда читает PostgreSQL, list-cache не используется

Формат ключа:

```text
task:<user_id>:<task_id>
```

TTL:

```text
5 минут
```

Почему кэш находится в service-слое:

- политика кэширования это решение use case
- repository остаётся persistence-only
- service может отдельно решать, какие операции кэшировать, а какие нет

```text
TaskService.GetTask
    |
    |-- Redis GET task:<user>:<task>
    |     |
    |     `-- hit -> вернуть задачу
    |
    `-- miss -> TaskRepository.Get
              -> Redis SET
              -> вернуть задачу
```

## Пайплайн аналитики

Аналитика обновляется асинхронно через Kafka.

Текущий поток:

- `TaskHandler.Create` публикует `task_created`
- `TaskHandler.ChangeStatus` публикует `task_completed`, если новый статус `done`
- `TaskHandler.Delete` публикует `task_deleted`
- `cmd/taskflow-worker` читает события из `KAFKA_TOPIC`
- worker обновляет агрегаты в `task_analytics`

```text
HTTP request
    |
    v
TaskHandler
    |
    |-- TaskService -> PostgreSQL / Redis
    |
    `-- Kafka publish (task event)
             |
             v
        taskflow-worker
             |
             v
        task_analytics
```

Это убирает обновление аналитики из синхронного request path. Цена такого решения:

- eventual consistency
- необходимость мониторить ошибки publish/consume
- необходимость отдельно управлять lifecycle worker-процесса

## Почему repository не знает о кэше

Repository должен быть слоем работы с постоянным хранилищем, а не местом orchestration.

Если repository начнёт:

- читать через Redis
- инвалидировать Redis
- публиковать Kafka events

то он перестанет быть чистым persistence-слоем и станет смешивать несколько разных обязанностей.

Текущая граница выбрана осознанно:

- repository отвечает за PostgreSQL
- service отвечает за прикладную координацию
- handlers отвечают за transport

Это упрощает тестирование и делает слой данных предсказуемым.

## PostgreSQL схема

Начальная миграция создаёт:

- `users`
- `tasks`
- `task_analytics`
- enum `task_status`

Структура:

```text
users
  id UUID PK
  email TEXT UNIQUE NOT NULL
  password_hash TEXT NOT NULL
  created_at TIMESTAMPTZ NOT NULL

tasks
  id UUID PK
  user_id UUID FK -> users.id ON DELETE CASCADE
  title TEXT NOT NULL
  description TEXT NOT NULL
  status task_status NOT NULL
  created_at TIMESTAMPTZ NOT NULL
  completed_at TIMESTAMPTZ NULL

task_analytics
  user_id UUID PK/FK -> users.id ON DELETE CASCADE
  tasks_created BIGINT
  tasks_completed BIGINT
  updated_at TIMESTAMPTZ NOT NULL
```

Сейчас `task_analytics` обновляется только worker-процессом.

## Middleware

В `PublicServer.Configure` регистрируются:

- `Recover`
- `RequestID`
- `RequestLogger`
- кастомный `AuthMiddleware` для защищённых маршрутов

Эффекты:

- panic в handler-слое не роняет процесс
- каждому запросу назначается request ID
- метаданные запроса логируются
- защищённые маршруты получают `userID` в контексте после успешной проверки токена

## Жизненный цикл приложения

Последовательность старта API:

1. `cmd/taskflow-api/main.go` загружает `.env`
2. `signal.NotifyContext` подписывается на `SIGINT` и `SIGTERM`
3. `Container.Init` создаёт инфраструктурные зависимости
4. `PublicServer.Configure` настраивает Echo, middleware, routes и Swagger endpoint
5. `App.Run` запускает Echo в отдельной goroutine
6. Главная goroutine ждёт `<-ctx.Done()`

```text
main
  -> load config
  -> init container
  -> configure server
  -> start server goroutine
  -> wait for signal
```

## Graceful Shutdown

Для API-процесса реализован graceful shutdown.

При получении сигнала завершения:

1. signal context отменяется
2. создаётся новый timeout context на 5 секунд
3. `App.ShutDown` вызывает `Echo.Shutdown`
4. закрывается PostgreSQL pool
5. закрывается Redis client
6. закрывается Kafka publisher

```text
SIGINT / SIGTERM
      |
      v
signal.NotifyContext cancelled
      |
      v
context.WithTimeout(5s)
      |
      v
Echo.Shutdown -> close DB pool -> close Redis -> close Kafka writer
```

## Lifecycle worker-процесса

Worker живёт отдельно от API.

Последовательность:

1. загружает тот же `AppConfig`
2. подключается к PostgreSQL
3. создаёт Kafka reader
4. вступает в consumer group
5. читает события аналитики
6. обновляет `task_analytics`
7. коммитит offset
8. завершает работу по `SIGINT` / `SIGTERM`

## Текущие ограничения

- Ошибки публикации аналитических событий пока не пробрасываются в HTTP response
- Аналитика асинхронная, поэтому значения в `task_analytics` обновляются с задержкой
- `ListTasks` не использует list-cache
- Ошибки API всё ещё возвращаются как raw JSON string, а не как структурированные error DTO
- Система миграций не хранит applied-state и выполняет все `*.up.sql` при запуске команды миграций

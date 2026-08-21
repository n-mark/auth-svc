# Auth Service (Сервис аутентификации)

Микросервис аутентификации и регистрации пользователей для маркетплейса объявлений (финальный проект). Написан на Go.

- **GitHub:** https://github.com/n-mark/auth-svc
- **DockerHub:** [`mblkuta/auth-service`](https://hub.docker.com/r/mblkuta/auth-service)

## Возможности

- Регистрация пользователей (`/api/v1/register`)
- Вход и выдача JWT-токенов (`/api/v1/login`)
- Подтверждение аккаунта (`/api/v1/confirm`)
- Публикация события `user.created` в Kafka (топик `auth`)

## Технологии

- Go, PostgreSQL, Kafka
- JWT для авторизации
- Docker / docker-compose

## Структура проекта

```text
cmd/           # точка входа
internal/      # бизнес-логика, обработчики, хранилище
pkg/           # переиспользуемые пакеты
init.sql       # инициализация БД
```

## Переменные окружения

| Переменная | Описание | Пример |
|---|---|---|
| `APP_PORT` | Порт HTTP-сервера | `:8080` |
| `DB_HOST` / `DB_PORT` | Хост/порт PostgreSQL | `postgres-shared` / `5432` |
| `DB_NAME` | Имя БД | `usersdb` |
| `DB_USER` / `DB_PASSWORD` | Учётные данные БД | `postgres` |
| `DB_MAX_RETRIES` / `DB_RETRY_INTERVAL` | Ретраи подключения к БД | `30` / `3` |
| `JWT_SECRET` | Секрет для подписи JWT | — |
| `BROKER_TYPE` | Тип брокера | `KAFKA` |
| `KAFKA_BROKERS` | Адреса брокеров Kafka | `kafka:9092` |
| `KAFKA_AUTH_TOPIC` | Топик событий аутентификации | `auth` |
| `KAFKA_USER_CREATED_EVENT_TYPE` | Тип события создания пользователя | `user.created` |
| `CONFIRM_BASE_URL` | Базовый URL для ссылок подтверждения | `http://localhost/api/v1` |

## Запуск

### Docker Compose

```bash
docker compose up -d
```

### Локально

```bash
go run ./cmd/...
```

## Эндпоинты

- `GET /health` — health-check
- `POST /api/v1/register` — регистрация
- `POST /api/v1/login` — вход (JWT)
- `GET /api/v1/confirm` — подтверждение аккаунта

## Связанные репозитории

Инфраструктура всего проекта (k8s, Helm, docker-compose всего стека): https://github.com/n-mark/final-project
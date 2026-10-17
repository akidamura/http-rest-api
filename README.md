# REST API и gRPC сервис на Go

Учебный бэкенд-сервис для управления пользователями. Поддерживает два транспорта: классический REST (JSON/HTTP) и высокопроизводительный gRPC (Protobuf/HTTP2). Построен на стандартной библиотеке Go без использования тяжелых веб-фреймворков.

> *Примечание: проект начинался как учебный REST API, поэтому имя репозитория (`http-rest-api`) осталось историческим. На данный момент это полноценный сервис с поддержкой двух транспортов.*

## Стек

- **Go** 1.27 — `net/http`, `log/slog`, `database/sql`
- **gRPC** + **Protobuf** — бинарный RPC-транспорт, строгий контракт
- **MySQL** 8 — основная БД
- **Nginx** — TLS-терминация для HTTP, reverse proxy
- **Docker** + **docker-compose** — контейнеризация
- **golang-migrate** — миграции БД
- **bcrypt** — хеширование паролей
- **gorilla/sessions** — сессии через куки

## Возможности

- CRUD пользователей через HTTP (REST) и gRPC (Protobuf)
- Аутентификация по email/паролю с выдачей сессионной куки (HTTP)
- Защищённые эндпоинты через middleware (HTTP)
- gRPC Server Reflection (для удобного тестирования через `grpcurl` и GoLand)
- HTTPS через Nginx (TLS-терминация)
- Health-check с проверкой БД
- Логирование запросов с `X-Request-Id`
- Graceful shutdown (HTTP и gRPC)
- Пул соединений MySQL
- Unit-тесты (HTTP + gRPC) и интеграционные тесты (БД)

## Быстрый старт

### Через Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone https://github.com/akidamura/http-rest-api.git
cd http-rest-api

# 2. Создать .env из шаблона
cp .env.example .env
# Отредактировать .env — задать DB_PASSWORD и SESSION_KEY

# 3. Сгенерировать самоподписанный сертификат для HTTPS
mkdir -p nginx/certs
openssl req -x509 -newkey rsa:2048 \
  -keyout nginx/certs/key.pem \
  -out nginx/certs/cert.pem \
  -days 365 -nodes \
  -subj "/CN=localhost"

# 4. Запустить всё
docker compose up --build
```

Сервисы поднимутся автоматически:

- **MySQL** на порту `3306`
- **Приложение (HTTP)** на порту `8080`
- **Приложение (gRPC)** на порту `50051`
- **Nginx** на портах `80` и `443`
- **Миграции** применятся при старте

**Проверка HTTP:**

```bash
curl -k https://localhost/health
# {"status":"ok"}
```

### Локально (без Docker)

**Требования:** Go 1.27, MySQL 8, `golang-migrate`.

```bash
# 1. Установить зависимости
go mod download

# 2. Создать базу
mysql -u root -p -e "CREATE DATABASE restapi_dev;"

# 3. Применить миграции
make migrate-up

# 4. Запустить сервер (HTTP + gRPC)
make run
```

## API

### HTTP REST API (порт 8080)

**Health**
- **`GET /health`** — проверка живости сервиса и БД

**Пользователи**
- **`POST /users`** — регистрация
- **`GET /users/{id}`** — получить пользователя
- **`PUT /users/{id}`** — обновить
- **`DELETE /users/{id}`** — удалить

**Сессии**
- **`POST /sessions`** — логин, устанавливает куку
- **`GET /private/whoami`** — текущий пользователь (нужна кука)

**Примеры:**

*Регистрация:*
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'
```

*Логин:*
```bash
curl -c .cookies -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'
```

*Получить текущего пользователя:*
```bash
curl -b .cookies http://localhost:8080/private/whoami
```

### gRPC API (порт 50051)

**Сервис:** `user.v1.UserService`
- `CreateUser` — регистрация
- `GetUser` — получение по ID
- `UpdateUser` — обновление
- `DeleteUser` — удаление

Для тестирования можно использовать `grpcurl`:
```bash
grpcurl -plaintext -d '{"email":"test@example.org", "password":"secret123"}' \
  localhost:50051 user.v1.UserService/CreateUser
```

Или встроенный HTTP Client в GoLand (файл `requests.grpc.http`).

## Usage

Типичный сценарий: регистрация → логин → работа с защищёнными эндпоинтами.

```bash
# 1. Регистрация нового пользователя
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'

# 2. Логин — получить сессионную куку (сохраняется в .cookies)
curl -c .cookies -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'

# 3. Защищённый запрос с кукой
curl -b .cookies http://localhost:8080/private/whoami

# 4. Обновить пользователя
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"email":"new@example.org","password":"newsecret123"}'

# 5. Удалить пользователя
curl -X DELETE http://localhost:8080/users/1
```

**Проверка всех сценариев сразу:**

```bash
make smoke
```

Полный список HTTP-эндпоинтов — в разделе [API](#api). Примеры всех сценариев — в файлах `requests.http` и `requests.grpc.http`.

## HTTPS через Nginx

Nginx принимает HTTPS на порту `443`, расшифровывает и проксирует запросы в Go-сервис по HTTP внутри Docker-сети. gRPC-трафик (порт `50051`) проксируется напрямую, минуя Nginx (в учебных целях).

```
Клиент (HTTP) → HTTPS (443) → Nginx → HTTP (8080) → Go
Клиент (gRPC) → TCP (50051) ──────────────────────→ Go
```

### Самоподписанный сертификат

Используется для локальной разработки. Сертификаты не коммитятся — каждый генерирует свои (см. «Быстрый старт»).

### Проверка HTTPS

```bash
# HTTP редиректит на HTTPS
curl -i http://localhost/health
# HTTP/1.1 301 Moved Permanently
# Location: https://localhost/health

# HTTPS с игнорированием самоподписанного сертификата
curl -k https://localhost/health
# {"status":"ok"}
```

### В продакшене

Вместо самоподписанного сертификата используется Let's Encrypt или корпоративный CA. Конфигурация Nginx и схема проксирования остаются те же.

## Тесты

```bash
make test              # все тесты (Unit + Интеграционные)
make test-units        # быстрые Unit-тесты (Model + HTTP + gRPC)
make test-grpc         # только тесты gRPC
make test-http         # только тесты HTTP
make test-integrations # интеграционные тесты с реальной MySQL
make test-cover        # отчёт о покрытии
make smoke             # smoke-тесты HTTP API
```

## Структура проекта

```
.
├── api/
│   └── proto/user/v1/      # Protobuf-контракты (user.proto)
├── cmd/apiserver/          # точка входа (запуск HTTP и gRPC)
├── internal/
│   ├── app/
│   │   ├── apiserver/      # HTTP-сервер, хендлеры, middleware
│   │   ├── grpcserver/     # gRPC-сервер, хендлеры, конвертеры
│   │   ├── model/          # модели данных
│   │   └── store/          # работа с БД
│   │       ├── sqlstore/   # MySQL-реализация
│   │       └── teststore/  # in-memory для тестов
│   ├── pb/                 # Сгенерированный Go-код из protobuf
│   └── config/             # общие утилиты
├── migrations/             # SQL-миграции
├── scripts/                # smoke-тесты
├── nginx/                  # конфиг Nginx + сертификаты
│   ├── nginx.conf
│   └── certs/              # сертификаты (не в Git)
├── requests.http           # примеры HTTP-запросов
├── requests.grpc.http      # примеры gRPC-запросов
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Переменные окружения

**Сервер:**

- `BIND_ADDR` — адрес HTTP-сервера (по умолчанию `:8080`)
- `LOG_LEVEL` — уровень логов (`debug`, `info`, `warn`, `error`)
- `LOG_FORMAT` — формат логов (`text` или `json`)

**MySQL:**

- `DB_HOST` — хост (по умолчанию `localhost`)
- `DB_PORT` — порт (по умолчанию `3306`)
- `DB_USER` — пользователь
- `DB_PASSWORD` — пароль
- `DB_NAME` — имя базы (по умолчанию `restapi_dev`)

**Сессии:**

- `SESSION_KEY` — ключ шифрования сессий (64 hex-символа)

## Миграции

```bash
make migrate-up        # применить к dev
make migrate-down      # откатить
make migrate-create name=add_column   # создать новую
```

## Лицензия

Учебный проект. Свободно используйте в образовательных целях.
# REST API на Go

Учебный REST API для управления пользователями: CRUD, аутентификация через сессии, health-check. Построен на стандартной библиотеке Go без внешних фреймворков.

## Стек

- **Go** 1.27 — `net/http`, `log/slog`, `database/sql`
- **MySQL** 8 — основная БД
- **Docker** + **docker-compose** — контейнеризация
- **golang-migrate** — миграции БД
- **bcrypt** — хеширование паролей
- **gorilla/sessions** — сессии через куки

## Возможности

- CRUD пользователей
- Аутентификация по email/паролю с выдачей сессионной куки
- Защищённые эндпоинты через middleware
- Health-check с проверкой БД
- Логирование запросов с `X-Request-Id`
- Graceful shutdown
- Пул соединений MySQL
- Unit + интеграционные тесты

## Быстрый старт

### Через Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone https://github.com/akidamura/http-rest-api.git
cd http-rest-api

# 2. Создать .env из шаблона
cp .env.example .env
# Отредактировать .env — задать DB_PASSWORD и SESSION_KEY

# 3. Запустить всё
docker compose up
```

Сервисы поднимутся автоматически:
- **MySQL** на порту `3306`
- **Приложение** на порту `8080`
- **Миграции** применятся при старте

**Проверка:**

```bash
curl http://localhost:8080/health
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

# 4. Запустить
make run
```

## API

### Health

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `GET` | `/health` | Проверка живости сервиса и БД |

### Пользователи

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/users` | Регистрация |
| `GET` | `/users/{id}` | Получить пользователя |
| `PUT` | `/users/{id}` | Обновить |
| `DELETE` | `/users/{id}` | Удалить |

### Сессии

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/sessions` | Логин, устанавливает куку |
| `GET` | `/private/whoami` | Текущий пользователь (нужна кука) |

### Примеры

**Регистрация:**

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'
```

**Логин:**

```bash
curl -c .cookies -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.org","password":"secret123"}'
```

**Получить текущего пользователя:**

```bash
curl -b .cookies http://localhost:8080/private/whoami
```

## Тесты

```bash
make test              # все тесты
make test-units        # только model
make test-integrations # только store (с MySQL)
make test-cover        # отчёт о покрытии
make smoke             # smoke-тесты API
```

## Структура проекта

```
.
├── cmd/apiserver/          # точка входа
├── internal/
│   ├── app/
│   │   ├── apiserver/      # HTTP-сервер, хендлеры, middleware
│   │   ├── model/          # модели данных
│   │   └── store/          # работа с БД
│   │       ├── sqlstore/   # MySQL-реализация
│   │       └── teststore/  # in-memory для тестов
│   └── config/             # общие утилиты
├── migrations/             # SQL-миграции
├── scripts/                # smoke-тесты
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `BIND_ADDR` | `:8080` | Адрес сервера |
| `LOG_LEVEL` | `info` | Уровень логов |
| `LOG_FORMAT` | `text` | Формат логов (`text` / `json`) |
| `DB_HOST` | `localhost` | Хост MySQL |
| `DB_PORT` | `3306` | Порт MySQL |
| `DB_USER` | `root` | Пользователь |
| `DB_PASSWORD` | — | Пароль |
| `DB_NAME` | `restapi_dev` | Имя базы |
| `SESSION_KEY` | — | Ключ шифрования сессий |

## Миграции

```bash
make migrate-up        # применить к dev
make migrate-down      # откатить
make migrate-create name=add_column   # создать новую
```

## Лицензия

Учебный проект. Свободно используйте в образовательных целях.
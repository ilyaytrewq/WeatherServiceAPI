# WeatherServiceAPI

Лёгкий сервис на Go для хранения пользователей и сбора погодных метрик.
Использует **PostgreSQL** для пользователей и **ClickHouse** для метрик, получает координаты/погоду через OpenWeather.

Сервис работает как HTTP API и слушает порт, указанный в переменных окружения (по умолчанию `8080`).

---

## Возможности

* Регистрация/удаление/обновление данных пользователя (email, пароль, города).
* Периодический сбор текущей погоды для городов и запись в ClickHouse.
* Логи входящих запросов, вызовов внешних API и ошибок.

---

## Стек

* Go
* PostgreSQL
* ClickHouse
* OpenWeather API
* Docker

---

## Переменные окружения (обязательные)

```bash
# Postgres
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgrespw
POSTGRES_DB=weatherdb

# ClickHouse (драйвер использует нативный порт)
CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=logs
CLICKHOUSE_PASSWORD=logs
CLICKHOUSE_DB=logs

# OpenWeather
API_WEATHER_KEY=your_openweather_api_key

# HTTP
HTTP_PORT=8080
```

---

## Быстрый запуск

```bash
# (в каталоге с compose.yml)
docker compose up --build
```

---

## HTTP API

Базовый префикс: `http://localhost:8080/v1`

### 1) `POST /v1/createUser`

Создать пользователя.

**Тело (JSON):**

```json
{
  "email": "user@example.com",
  "password": "secret",
  "cities": ["Tokyo", "Ufa", "Moscow"]
}
```

**curl:**

```bash
curl -X POST http://localhost:8080/v1/createUser \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "secret",
    "cities": ["Tokyo", "London", "Moscow"]
  }'
```

**Успех (201):**

```json
{"message":"User registered successfully"}
```

---

### 2) `POST /v1/changeUserData`

Изменить список городов (нужно указать `email` и `password` для авторизации).

**Тело (JSON):**

```json
{
  "email": "user@example.com",
  "password": "secret",
  "cities": ["Berlin", "Amsterdam"]
}
```

**curl:**

```bash
curl -X POST http://localhost:8080/v1/changeUserData \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "password":"secret",
    "cities":["Berlin","Amsterdam"]
  }'
```

**Успех (200):**

```json
{"message":"User data updated successfully"}
```

---

### 3) `POST /v1/getUserData`

**Тело (JSON):**

```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

**curl:**

```bash
curl -X POST http://localhost:8080/v1/getUserData \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "password":"secret"
  }'
```

**Успех (200):**

```json
{"email":"user@example.com","cities":["Berlin","Amsterdam"]}
```

---

### 4) `DELETE /v1/deleteUser`

Удалить пользователя (нужны `email` и `password` в теле).

**curl:**

```bash
curl -X DELETE http://localhost:8080/v1/deleteUser \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "password":"secret"
  }'
```

**Успех (200):**

```json
{"message":"User deleted successfully"}
```

---

## Логи и отладка

Сервис использует `log.Printf` для логирования:

* входящие запросы (метод, путь, IP и sample тела),
* результат парсинга JSON (email + cities),
* вызовы OpenWeather (URL, статус, короткий сэмпл ответа),
* ошибки DB/ClickHouse/внешних вызовов.


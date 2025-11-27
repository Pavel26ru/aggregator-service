# Сервис-агрегатор

Сервис `aggregator-service` предназначен для асинхронной агрегации данных, поступающих через Kafka, и предоставления доступа к ним через gRPC и REST API.

## Архитектура

Проект построен на основе принципов **Чистой Архитектуры** (Clean Architecture).

- **Асинхронная агрегация:** Данные генерируются, отправляются в топик Kafka, обрабатываются пулом воркеров-консьюмеров (которые находят максимальное значение в пакете) и сохраняются в базу данных PostgreSQL.
- **Синхронный доступ:** Клиенты могут запрашивать агрегированные данные по `uuid` или за определенный период времени через API.

## Как запустить

### 1. Конфигурация

Перед запуском необходимо создать файл `.env` в корневой директории проекта. Можете скопировать и использовать пример ниже:

```env
# .env

# === App Config ===
ENV=local
WORKERS=5
INTERVAL=100ms # интервал генерации новых сообщений в Kafka

# === Ports ===
HTTP_PORT=8080
GRPC_PORT=9090

# === Postgres ===
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=agg
POSTGRES_SSL=disable

# === Kafka ===
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=records
KAFKA_GROUP=agg-workers
```

### 2. Запуск через Docker Compose

Все компоненты сервиса (приложение, PostgreSQL, Kafka, Zookeeper) упакованы в Docker. Для запуска выполните команду:

```bash
docker-compose up --build
```

Сервис будет доступен после того, как все контейнеры запустятся.

## Как проверить работоспособность API

После запуска сервис начинает генерировать данные. Вы можете наблюдать за логами контейнера `aggregator-service`, чтобы увидеть UUID созданных записей.

```bash
docker-compose logs -f aggregator-service
```

Вы увидите логи, похожие на эти, откуда можно взять `uuid` для тестов:
```
... "msg":"incoming request","method":"GET","path":"/max" ...
... "msg":"record not found","uuid":"a1b2c3d4-..." ...
```

### 1. Проверка REST API (порт 8080)

Вы можете использовать `curl` для отправки запросов.

#### Получить запись по UUID

Подставьте `uuid` из логов в команду ниже:
```bash
# Замените a1b2c3d4-... на реальный UUID
curl "http://localhost:8080/max?uuid=a1b2c3d4-e5f6-a7b8-c9d0-e1f2a3b4c5d6"
```

#### Получить записи за период времени

Для запроса по времени используйте формат `RFC3339`.
```bash
# Пример запроса за последние 5 минут
FROM_TIME=$(date -v-5M -u +'%Y-%m-%dT%H:%M:%SZ')
TO_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

curl "http://localhost:8080/max?from=${FROM_TIME}&to=${TO_TIME}"
```

### 2. Проверка gRPC API (порт 9090)

Для проверки gRPC удобно использовать утилиту `grpcurl`.

#### Показать доступные сервисы
```bash
grpcurl -plaintext localhost:9090 list
# Вывод: aggregator.AggregatorService
```

#### Получить запись по UUID
```bash
# Замените a1b2c3d4-... на реальный UUID
grpcurl -plaintext -d '{"uuid": "a1b2c3d4-e5f6-a7b8-c9d0-e1f2a3b4c5d6"}' \
  localhost:9090 aggregator.AggregatorService/GetMax
```

#### Получить записи за период времени
```bash
# Пример запроса за последние 5 минут
FROM_TS=$(date -v-5M +%s)
TO_TS=$(date +%s)

grpcurl -plaintext -d '{"from": {"seconds": '$FROM_TS'}, "to": {"seconds": '$TO_TS'}}' \
  localhost:9090 aggregator.AggregatorService/GetMax
```
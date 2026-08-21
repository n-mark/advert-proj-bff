# BFF (Backend for Frontend)

BFF-сервис для маркетплейса объявлений (финальный проект). Агрегирует данные из всех микросервисов в единый API для фронтенда. Написан на Go.

- **GitHub:** https://github.com/n-mark/advert-proj-bff
- **DockerHub:** [`mblkuta/advert-proj-bff`](https://hub.docker.com/r/mblkuta/advert-proj-bff)

## Возможности

- Агрегация объявлений, заказов и кабинета пользователя в единые эндпоинты
- Прокси/оркестрация запросов к сервисам: auth, profile, advert-cmd, advert-query, order, billing, delivery, dialog

## Технологии

- Go
- Docker / docker-compose

## Структура проекта

```text
cmd/           # точка входа
internal/      # обработчики, клиенты к downstream-сервисам
```

## Переменные окружения

| Переменная | Описание | Пример |
|---|---|---|
| `APP_PORT` | Порт HTTP-сервера | `8080` |
| `AUTH_URL` | URL сервиса аутентификации | `http://auth-service:8080` |
| `PROFILE_URL` | URL сервиса профилей | `http://profile-service:8080` |
| `ADVERT_CMD_URL` | URL сервиса команд объявлений | `http://advert-cmd-svc:8080` |
| `ADVERT_QUERY_URL` | URL сервиса поиска | `http://advert-query:8080` |
| `ORDER_URL` | URL сервиса заказов | `http://order-service:8080` |
| `BILLING_URL` | URL сервиса биллинга | `http://billing-service:8080` |
| `DELIVERY_URL` | URL сервиса доставки | `http://delivery-service:8080` |
| `DIALOG_URL` | URL сервиса диалогов | `http://dialog-service:8080` |

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
- `GET /api/v1/bff/adverts/{id}` — агрегированные данные объявления
- `GET /api/v1/bff/orders/{id}` — агрегированные данные заказа
- `GET /api/v1/bff/users/{id}/cabinet` — кабинет пользователя

## Связанные репозитории

Инфраструктура всего проекта (k8s, Helm, docker-compose всего стека): https://github.com/n-mark/final-project
# Сервер

## Запуск

```bash
go mod tidy
go run ./cmd/api
```

По умолчанию слушает `:8080`.

## Переменные

| Переменная | По умолчанию | Зачем |
| --- | --- | --- |
| `LISTEN` | `:8080` | адрес и порт |
| `DATABASE_URL` | пусто | PostgreSQL |
| `AI_BASE_URL` | пусто | адрес ИИ-провайдера |
| `AI_API_KEY` | пусто | ключ |
| `AI_VISION_MODEL` | пусто | модель для картинок |
| `AI_TEXT_MODEL` | пусто | модель для ссылок |

Без ключа API работает: формальные проверки доступны, семантика — «не проверено».

## Миграции

Схема и миграции — в `migrations/` (goose, встроены в бинарник командой `cmd/migrate`).

```bash
# поднять базу (Docker, порт 5432 проброшен на хост)
docker compose -f deploy/docker-compose.yml up -d db

# применить миграции
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/accessible?sslmode=disable"
go run ./cmd/migrate up

# статус / откат
go run ./cmd/migrate status
go run ./cmd/migrate down
```

## Структура

```
cmd/api/            точка входа
internal/api/       HTTP-роуты и middleware
internal/store/     память сейчас, PostgreSQL потом
internal/crawler/   Colly
internal/checks/    проверки
internal/pdf/       PDFium + Tesseract
internal/ai/        vision и текст
internal/report/    CSV и PDF
migrations/         SQL-миграции
```

## Тесты

```bash
go vet ./...
go test ./...
go build ./...
```

## PDF

Пока PDF-анализ не подключён. Задача — в `internal/pdf/README.md`.
Python не используется.

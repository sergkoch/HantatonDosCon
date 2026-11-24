# Деплой

## Локально

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Сайт: `http://localhost:8080`.

## Что внутри

- `db` — PostgreSQL (порт 5432 проброшен на хост для локальных миграций);
- `migrate` — применяет миграции (goose) и завершается;
- `api` — Go-сервис (ждёт успешных миграций);
- `web` — nginx + Svelte.

Наружу открыт только `web` (порт 8080).

## Переменные ИИ (необязательно)

Скопируй `.env.example` в `.env` и заполни, если нужен ИИ:

```
AI_BASE_URL=https://api.openai.com/v1
AI_API_KEY=...
AI_TEXT_MODEL=gpt-4o-mini
AI_VISION_MODEL=gpt-4o-mini
```

Без ключа сервис работает: формальные проверки доступны, семантика — «не проверено».

## Сервер

1. Поставить Docker и Docker Compose.
2. Склонировать репозиторий.
3. Создать `.env` на сервере.
4. Запустить:

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

5. Настроить домен: A-запись на IP сервера.
6. Поставить TLS: Let's Encrypt (`certbot`) + nginx или Caddy перед `web`.
7. Проверить:

```bash
curl -I https://<домен>
curl https://<домен>/api/health
```

## Обновление

```bash
git pull
docker compose -f deploy/docker-compose.yml up -d --build
```

## Бэкап

```bash
docker compose -f deploy/docker-compose.yml exec db \
  pg_dump -U postgres accessible > backup.sql
```

## Переменные

- `SCAN_ALLOWED_HOSTS` — хосты, которым разрешено ходить на приватные адреса.
  Для локального демо там стоит `demo`. В проде оставьте пустым.
- `RATE_LIMIT_PER_MINUTE` — сколько сканов принимает API за минуту, по умолчанию 20.
- `AI_BASE_URL`, `AI_API_KEY`, `AI_TEXT_MODEL`, `AI_VISION_MODEL` — доступ к ИИ.
  Без ключа проверки ИИ пропускаются.

## Безопасность

- Наружу — только 80/443.
- API не публикуем напрямую.
- Сканер не ходит на `localhost`, приватные и link-local IP: проверка идёт и при
  валидации URL, и при подключении (защита от DNS rebinding).
- Тело запроса ограничено 1 МиБ, сканы — 20 в минуту.
- Секреты — в `.env`, не в образе.
- Логи не содержат ключей.
- PDF-анализатор — отдельный контейнер (при подключении).

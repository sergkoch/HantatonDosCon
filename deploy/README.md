# Деплой

## Локально

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Сайт: `http://localhost:8080`.

## Что внутри

- `db` — PostgreSQL;
- `api` — Go-сервис;
- `web` — nginx + Svelte.

Наружу открыт только `web`.

## Сервер

1. Поставить Docker и Docker Compose.
2. Склонировать репозиторий.
3. Создать `.env` на сервере.
4. Запустить:

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

5. Настроить домен: A-запись на IP сервера.
6. Поставить TLS: Let's Encrypt + nginx или Caddy.
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

## Безопасность

- Наружу — только 80/443.
- API не публикуем напрямую.
- Секреты — в `.env`, не в образе.
- Логи не содержат ключей.
- PDF-анализатор — отдельный контейнер.

# Production deploy

## Что нужно

1. Сервер с Docker и Docker Compose.
2. DNS: A-запись `not-domain.ru` → IP сервера (уже настроено).
3. Открытые порты **80** и **443**.
4. Репозиторий на сервере, например `/opt/accessible-content`.
5. `.env` в корне репозитория:

```env
POSTGRES_PASSWORD=смените_пароль
AI_BASE_URL=
AI_API_KEY=
AI_TEXT_MODEL=
AI_VISION_MODEL=
SCAN_ALLOWED_HOSTS=
RATE_LIMIT_PER_MINUTE=20
```

`SCAN_ALLOWED_HOSTS` в проде оставьте пустым. Он нужен только для локального
демо-сайта `demo`.

## Запуск

```bash
cd /opt/accessible-content
git fetch origin main
git reset --hard origin/main
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

Caddy сам получит HTTPS-сертификат Let's Encrypt для `not-domain.ru`.

## Проверка

```bash
curl -I https://not-domain.ru
curl https://not-domain.ru/api/health
```

Ожидаемо:

- `HTTP/2 200`;
- `{"status":"ok",...}`.

## Обновление

```bash
cd /opt/accessible-content
git fetch origin main
git reset --hard origin/main
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

## Логи

```bash
docker compose -f deploy/docker-compose.prod.yml logs -f caddy api
```

# AGENTS.md

Гайд для ИИ-агентов и людей. Читается автоматически в каждой сессии.

## Что за проект

«Доступный контент» — сервис проверки доступности сайтов. Проверяет:

- `alt` изображений, в том числе осмысленность по самой картинке (vision-модель);
- ссылки: доступность и понятность текста (ИИ);
- формы, таблицы, заголовки, язык страницы;
- PDF: текстовый слой, теги структуры, `/Alt`, `Title`, `Lang`.

Результат — отчёт с оценкой, находками и рекомендациями. Есть выгрузка CSV и PDF,
история сканов, версии и сравнение с предыдущей проверкой.

## Структура

```
client/      Svelte 5 + TypeScript + Vite; статика, nginx проксирует /api
server/      Go + Gin; cmd/api, cmd/migrate; миграции в server/migrations
deploy/      docker-compose (db, migrate, api, web) + nginx
demo-site/   сайт с намеренными ошибками для демо
docs/        план, архитектура, API, состояние проекта
tasks/       задачи по дням
```

## Команды

### API

```bash
cd server
go run ./cmd/api      # http://localhost:8080
go test ./...
go vet ./...
```

### Клиент

```bash
cd client
bun install --frozen-lockfile
bun run dev           # http://localhost:5173, /api проксируется на :8080
bun run check         # svelte-check
bun run test          # vitest
bun run build
```

### Всё в Docker

```bash
docker compose -f deploy/docker-compose.yml up -d --build
docker compose -f deploy/docker-compose.yml logs -f api
# приложение: http://localhost:8080
```

### Линт

```bash
bun run lint          # biome
```

## Окружение

- `.env` в корне — для локального запуска.
- `deploy/.env` — для Docker Compose (именно его читает compose).
- Переменные: `DATABASE_URL`, `AI_BASE_URL`, `AI_API_KEY`, `AI_TEXT_MODEL`,
  `AI_VISION_MODEL`, `SCAN_ALLOWED_HOSTS`, `RATE_LIMIT_PER_MINUTE`.
- Секреты не коммитим. `.env` и `deploy/.env` в `.gitignore`.

## Правила работы

1. Одна задача = один issue = одна ветка = один PR. В `main` только через PR.
2. Ветки: `task/<slug>`, для PR — видимый алиас `team/<slug>`.
   Аккаунт `heisenbaf` скрыт GitHub, поэтому PR открывает `sergkoch` или
   `shrekichevsk`.
3. Тексты интерфейса, ошибки и документация — по-русски.
4. Перед PR обязательно: `go test ./...`, `go vet ./...`, `bun run check`,
   `bun run test`, `bun run build`.
5. Не трогать чужие незакоммиченные файлы. Незнакомый код сначала разобрать,
   потом менять.
6. Коммиты: `feat(scope): ...`, `fix(scope): ...`, `test(scope): ...`,
   `docs(scope): ...`.

## Как проверить, что всё работает

```bash
curl -s http://localhost:8080/api/health
# {"status":"ok","ai":"configured"}

curl -s -X POST http://localhost:8080/api/scans \
  -H 'Content-Type: application/json' \
  -d '{"url":"http://demo/","depth":1}'
# затем GET /api/scans/{id} и отчёт /report/{id}
```

## Сервер

- Хост: `84.54.57.218`, пользователь `andrey`, репозиторий `/opt/accessible-content`.
- Git на сервере — от `serega`: `sudo -u serega git ...`.
- Docker — только через `sudo docker ...`.
- Обновление: `sudo -u serega git fetch origin main && sudo -u serega git reset --hard origin/main`,
  затем `sudo docker compose -f deploy/docker-compose.yml up -d --build`.
- Ключи ИИ лежат в `deploy/.env` на сервере.

## Состояние

Актуальный статус, блокеры и следующие шаги — `docs/state.md`
(подключён через `instructions` в `opencode.jsonc`, загружается сам).

Коротко: основная функциональность в `main`; не смержена ветка
`team/review-fixes`; домен `not-domain.ru` не открывается, потому что Cloud.ru
не пускает 80/443 — снаружи доступен только SSH 22, поэтому используется туннель.

## Стиль ответов

- Кратко, по делу, без пересказа очевидного.
- Сначала результат и проверка, потом детали.
- Все команды — готовые к копированию.

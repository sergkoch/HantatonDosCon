# Состояние проекта

Обновлено: 25.11.2026. **Обновляй этот файл после каждой крупной задачи** —
он загружается в контекст ИИ автоматически и заменяет длинную историю чата.

## Готово

- Каркас: Svelte 5 + Go/Gin, Docker Compose, демо-сайт с ошибками.
- Краулер: обход домена, глубина, `robots.txt`.
- Проверки: `alt`, ссылки, формы, таблицы, заголовки, язык.
- PDF: текстовый слой, теги, `/Alt`, `Title`, `Lang` (go-pdfium + pdfcpu).
- ИИ: текст ссылок и `alt` по картинке (DeepSeek, vision).
- Отчёт: оценка, фильтры, покрытие, CSV, PDF, автооткрытие после скана.
- История: версии сканов, журнал, diff с предыдущей проверкой.
- PostgreSQL-хранилище при заданном `DATABASE_URL`.
- Безопасность: SSRF-защита, лимиты, CORS, allowlist.
- SEO: `robots.txt`, метатеги; Lighthouse 100/100/100.

## Влитые PR

`#73` история, `#74` PDF-экспорт, `#76` автооткрытие отчёта,
`#77` новый экран отчёта, `#78` ИИ-проверка картинок, `#68` SEO,
`#69` PDF-проверки, `#70` безопасность.

`main` = `aa39948`.

## Не влито

- `team/review-fixes` — правки по документу ревью: Docker/`go.sum`, таймауты,
  скрытие внутренних ошибок, русские ошибки API, CSV-безопасность, polling,
  a11y роутера и форм, `aria-current`, лимит памяти, CORS, CSP в nginx,
  favicon/noscript, CI.
- PR нужно открыть от `sergkoch` или `shrekichevsk`:
  `https://github.com/sergkoch/HantatonDosCon/compare/main...team/review-fixes?expand=1`

## Сервер

- `84.54.57.218`, пользователь `andrey`, репозиторий `/opt/accessible-content`.
- Стек развёрнут, `ai=configured`, история работает, `/` и `/history` → 200.
- Nginx настроен на `not-domain.ru`.

## Блокеры

1. **Домен.** Cloud.ru не пускает 80/443: снаружи открыт только SSH 22.
   Нужно открыть порты в панели Cloud.ru (Security Group / NAT), затем
   `certbot --nginx -d not-domain.ru -d www.not-domain.ru`.
2. **Публичный доступ.** Пока работает через туннель Cloudflare Quick.
   Адрес меняется при перезапуске, большие ответы иногда обрезаются.
   Текущий адрес:
   ```bash
   ssh andrey@84.54.57.218 "sudo journalctl -u cloudflared-quick --no-pager -o cat | grep -oE 'https://[a-z0-9-]+\.trycloudflare\.com' | tail -1"
   ```

## Осталось по ревью

- `ScanPage`: `AbortController` и повтор запроса.
- Компонентные тесты клиента.
- Ограничения в БД: `CHECK`, `updated_at`.
- Чистка мёртвого CSS.
- CI на Docker-сборку, biome в CI.

## Ключи

Ключи ИИ — только в `.env` и `deploy/.env` (в `.gitignore`). Оба ключа
засветились в чате, после проекта их нужно перевыпустить.

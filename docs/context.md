# Контекст окружения

Долговременные факты, которые нельзя потерять между сессиями. Загружается
автоматически. Статус задач — в `docs/state.md`.

Обновлено: 28.11.2026.

---

## 1. GitHub

- Репозиторий: `https://github.com/sergkoch/HantatonDosCon`
- Основная ветка: `main`, на 28.11.2026 — `8983af6`.
- В `main` только через PR и зелёный CI.
- Ветка задачи: `task/<slug>`. Для PR создаётся видимый алиас `team/<slug>`.

### Теневой бан (важно)

Аккаунт `heisenbaf` в теневом бане GitHub:

- `https://api.github.com/users/heisenbaf` → 404;
- его issues и PR публично отдают 404;
- ветки и коммиты при этом видны.

Поэтому **PR открывает `sergkoch` или `shrekichevsk`**, а не `heisenbaf`.
Если PR не виден — причина именно в этом, а не в правах.

### Issues

- Видимый набор — issues `sergkoch` #33–#59.
- Соответствие: #39 = T07, #47 = T15 (PDF), #50 = T18, #55 = T23 (robots),
  #58 = T26 (безопасность), #48 = T16 (ИИ-картинки).
- **Issues пока не закрывать** — так просил владелец репозитория.
- Ранее часть issues закрыли и потом переоткрыли (#37, #39–#45, #47–#59).

### Влитые PR

`#68` robots/SEO, `#69` PDF-проверки, `#70` безопасность, `#73` история и diff,
`#74` PDF-отчёт, `#76` автооткрытие отчёта, `#77` новый экран отчёта,
`#78` ИИ-проверка картинок, `#95` правки ревью 27.11, `#96` UI-полировка,
`#97` язык отчёта, `#98` финальная полировка.

---

## 2. Сервер

- Хост: `84.54.57.218`, пользователь `serega`.
- Приватный IP `10.0.0.5`, исходящий IP совпадает с публичным.
- Ключ: `~/.ssh/id_ed25519_hackathon`.
- ОС: Ubuntu 24.04, KVM, провайдер **Cloud.ru**.
- Репозиторий: `/opt/accessible-content`, владелец `serega`.
- Git на сервере — только от `serega`: `sudo -u serega git ...`.
- Docker — только через `sudo docker ...`.
- Compose-файл: `deploy/docker-compose.yml`; сервисы: `db`, `migrate`, `api`,
  `web`, `demo` (статика `demo-site/`, доступна сканеру как `http://demo/`).
- Переменные окружения compose читает из `deploy/.env` (не из корневого `.env`).
- Порт **80 открыт снаружи** и отдаёт приложение: `http://not-domain.ru` работает.
  Порт 443 до 25.11.2026 не слушался вовсе (не было сертификата), поэтому
  подключение отклонялось — это выглядело как «закрытые порты», но блокировки
  на стороне Cloud.ru не было.

### Обновление сервера

```bash
ssh -i ~/.ssh/id_ed25519_hackathon serega@84.54.57.218
cd /opt/accessible-content
sudo -u serega git fetch origin main
sudo -u serega git reset --hard origin/main
sudo docker compose -f deploy/docker-compose.yml up -d --build
curl -s http://localhost:8080/api/health
```

---

## 3. Домен и публичный доступ

- Домен `not-domain.ru` работает: `http://` и `https://` отдают приложение
  (`200`), включая `www`.
- Nginx на сервере настроен на `not-domain.ru` и проксирует на `localhost:8080`
  (конфиг лежит в `/etc/nginx/sites-enabled/cool-domain.online`).
- Сертификат Let's Encrypt выпущен 25.11.2026, действует до 23.02.2027,
  продление автоматическое. Принудительный редирект HTTP→HTTPS не включён.

```bash
sudo certbot certificates                      # проверить срок
sudo certbot renew --dry-run                   # проверить продление
curl -I https://not-domain.ru                  # проверить снаружи
```

### Временный туннель

`cloudflared` quick tunnel на `http://localhost:8080` — **больше не нужен**,
домен открывается напрямую. Сервис можно погасить:

```bash
sudo systemctl disable --now cloudflared-quick
```

Ограничения, уже проверенные на практике:

- адрес меняется при каждом перезапуске;
- большие файлы иногда обрезаются (JS 84 КБ доходил только частично),
  из-за этого возможен белый экран;
- независимые прокси получают 522 — значит проблема на стороне туннеля,
  а не приложения.

Пробовали и отбросили: `serveo` и `pinggy` (показывают страницу-предупреждение),
`localhost.run` (работает, но адрес тоже плавающий), Cloudflare named tunnel
(нужен аккаунт Cloudflare и перенос NS домена).

Сеть сервера: у `enp3s0` был MTU 9000 (jumbo frames), вручную поставлен 1500.
Изменение не сохранено в конфиге — после перезагрузки вернётся 9000.

---

## 4. ИИ

- Провайдер: **DeepSeek**, OpenAI-совместимый API.
- `AI_BASE_URL=https://api.deepseek.com/v1`
- Модели: `deepseek-flash` (текст и vision), `deepseek-v4-pro`.
- Vision работает: модель принимает изображение в `image_url` как data URL.
- Проверка ссылок и проверка `alt` по картинке реализованы.
- Ключ Google Gemini проверен и валиден, но Google блокирует регион:
  `User location is not supported` — использовать нельзя.
- Ключи хранятся только в `.env` и `deploy/.env` (в `.gitignore`).
- Оба ключа засветились в переписке — после проекта перевыпустить.

---

## 5. Технические решения

- PDF: `go-pdf/go-pdfium` (WebAssembly, без cgo и DLL) + `pdfcpu` как fallback.
- PDF-отчёт: `go-pdf/fpdf` со встроенным `goregular.TTF` (кириллица работает).
- Хранилище: PostgreSQL при заданном `DATABASE_URL`, иначе память.
- История: `version`, `previousId`, `changes{added,fixed,remaining,scoreDelta}`,
  эндпоинт `GET /api/scans`, страница `/history`.
- Безопасность: SSRF-защита `internal/safehttp`, rate limit, лимит тела,
  allowlist хостов, CORS через env.
- Демо-сайт внутри compose доступен как `http://demo/`, скан запускать на него.

---

## 6. Локальная работа

Главная рабочая папка: `C:\Users\l123r\Downloads\accessible-content\accessible-content`.

Старые worktree (`ac-polish`, `ac-history`, `ac-pdf`, `accessible-content-t26`,
`ac-prod`, `ac-vision`) не удалять без разбора: их ветки уже влиты в `main`,
кроме исторических локальных копий.

Документ ревью: `C:\Users\Андрей\Desktop\Чекпоинт_Доступный_контент_23-11-2026`,
читаемые копии — `%TEMP%\checkpoint-doc\00.txt`–`07.txt`.

---

## 7. Чего не делать

- Не закрывать issues без просьбы владельца.
- Не коммитить `.env`, `deploy/.env`, ключи и пароли.
- Не пушить напрямую в `main` — только ветка и PR.
- Не удалять чужие worktree и незакоммиченные файлы.
- Не перезаписывать `git reset --hard` локально без предупреждения.
- Не менять настройки Cloud.ru «наугад» — доступ только у владельца.

---

## 8. Короткий бриф для нового чата

Если контекст почему-то не загрузился, вставь это первым сообщением:

```text
Проект «Доступный контент»: проверка доступности сайтов (alt, ссылки, формы,
таблицы, заголовки, PDF) + отчёт HTML/CSV/PDF + история и diff.
Стек: Svelte 5 + Go/Gin + PostgreSQL + Docker Compose.
Репозиторий sergkoch/HantatonDosCon, main 8983af6, PR только через ветку.
Аккаунт heisenbaf в теневом бане: PR открывает sergkoch или shrekichevsk.
Сервер 84.54.57.218, пользователь serega, /opt/accessible-content,
git от serega, docker через sudo, env в deploy/.env.
Домен not-domain.ru работает: http и https (Let's Encrypt, до 23.02.2027), www тоже.
Сайт открыт снаружи, туннель cloudflared-quick больше не нужен.
ИИ — DeepSeek (deepseek-flash, vision работает).
Подробности: AGENTS.md, docs/context.md, docs/state.md.
```

# Инфраструктура: зеркало GitHub ↔ GitLab и деплой на сервер

## Схема

```
GitHub (sergkoch/HantatonDosCon)  ──push main──►  GitHub Actions
   │                                                  ├── mirror-gitlab.yml  →  GitLab (зеркало)
   │                                                  └── deploy.yml         →  сервер: git pull + docker compose up
   ▼
GitLab (зеркало, read-only)
```

- **Источник правды — GitHub.** `main` пушится в GitLab как зеркало (односторонне).
- **Деплой:** push в `main` → GitHub Actions по SSH заходит на сервер → `git pull` + `docker compose up -d --build`.

## Переключатели (repository variables)

Чтобы воркфлоу были безопасны до настройки доступов, они включаются переменными
(Settings → Secrets and variables → Actions → Variables):

| Переменная | Значение | Что делает |
| --- | --- | --- |
| `DEPLOY_ENABLED` | `true` | включает деплой на сервер |
| — | — | зеркало включается автоматически, если заданы `GITLAB_*` секреты |

Пока `DEPLOY_ENABLED` не задан — деплой-джоб пропускается (main остаётся зелёным).

## Требуемые GitHub Secrets

| Секрет | Что | Пример |
| --- | --- | --- |
| `GITLAB_TOKEN` | Personal Access Token GitLab, scope `write_repository` | `glpat-...` |
| `GITLAB_REPO_URL` | HTTPS-URL GitLab-проекта | `https://gitlab.com/group/project.git` |
| `DEPLOY_HOST` | IP/домен сервера | `84.54.57.218` |
| `DEPLOY_USER` | SSH-логин | `serega` |
| `DEPLOY_SSH_KEY` | приватный SSH-ключ (с `-----BEGIN...`) | содержимое `id_ed25519_hackathon` |
| `DEPLOY_PATH` | путь к репозиторию на сервере | `/opt/accessible-content` |

CLI:

```bash
gh secret set GITLAB_REPO_URL --body "https://gitlab.com/group/project.git"
gh secret set GITLAB_TOKEN            # вставить токен
gh secret set DEPLOY_HOST --body "84.54.57.218"
gh secret set DEPLOY_USER --body "serega"
gh secret set DEPLOY_PATH --body "/opt/accessible-content"
gh secret set DEPLOY_SSH_KEY < id_ed25519_hackathon
gh variable set DEPLOY_ENABLED --body "true"
```

## Сервер: что должно быть готово

1. Установлены Docker и Docker Compose.
2. Репозиторий склонирован в `DEPLOY_PATH`, `origin` смотрит на GitHub.
3. Создан `.env` (ключи ИИ и т.п.).
4. Первый запуск: `docker compose -f deploy/docker-compose.yml up -d --build`.

## ⚠️ Важные оговорки

- **VPN.** Если SSH к серверу доступен только через VPN AmneziaWG, облачный
  GitHub Actions до сервера не достучится. Тогда — публичный SSH, self-hosted
  runner в контуре, либо серверный `cron` вместо Actions.
- **Приватность ключа.** `DEPLOY_SSH_KEY` — ключ с sudo-доступом. Репозиторий
  публичный; секреты шифруются, но безопаснее завести **отдельный деплой-ключ**
  с ограниченными правами, а не root-ключ.
- **GitLab.** Для зеркала на `main` должен быть разрешён **force-push**
  (Settings → Repository → Protected branches → `main` → Allow force push),
  иначе пуш отклоняется: «not allowed to force push to a protected branch».
  Альтернатива — встроенный **Pull Mirror** в GitLab (Settings → Repository →
  Mirroring), тогда `mirror-gitlab.yml` не нужен.
- **Локальный Docker.** Docker Desktop в текущей среде не стартует (нет WSL2-
  дистрибутива). Docker-стек и зеркало запускаются на сервере и в GitHub Actions.

## Ручной деплой (если Actions недоступны)

```bash
ssh -i ~/.ssh/id_ed25519_hackathon serega@84.54.57.218
cd /opt/accessible-content
git fetch origin main && git reset --hard origin/main
docker compose -f deploy/docker-compose.yml up -d --build
```

# API

Базовый адрес в разработке: `http://localhost:8080`.

## GET /api/health

Проверка, что сервис жив.

```json
{
  "status": "ok",
  "version": "0.1.0",
  "ai": "missing",
  "time": "2026-11-22T10:00:00Z"
}
```

`ai` принимает `configured` или `missing`.

## POST /api/scans

Запускает проверку.

Запрос:

```json
{
  "url": "https://example.ru",
  "depth": 2
}
```

Правила:

- `url` — только `http` или `https`;
- `url` должен вести на публичный адрес: `localhost`, приватные и link-local IP
  отклоняются (`400`, `url must be public`);
- исключение — хосты из `SCAN_ALLOWED_HOSTS` (нужно для локального демо-сайта);
- `depth` — от 1 до 3;
- по умолчанию `1`;
- тело запроса — не больше 1 МиБ (`413`);
- не больше 20 сканов в минуту с одного инстанса (`429`, `Retry-After: 60`).

Ответ `201 Created`:

```json
{
  "id": "0f7f0f2e-6a0b-4a3e-9f4e-7d1b2c3d4e5f",
  "url": "https://example.ru",
  "depth": 2,
  "status": "queued",
  "score": null,
  "stats": { "critical": 0, "warning": 0, "ok": 0 },
  "findings": [],
  "createdAt": "2026-11-22T10:00:00Z"
}
```

## GET /api/scans/{id}

Возвращает скан.

```json
{
  "id": "0f7f0f2e-6a0b-4a3e-9f4e-7d1b2c3d4e5f",
  "url": "https://example.ru",
  "depth": 2,
  "status": "done",
  "version": 2,
  "previousId": "9f8e7d6c-5b4a-4321-8765-fedcba987654",
  "score": 42,
  "pages": 3,
  "checkedLinks": 18,
  "unverifiedLinks": 0,
  "stats": { "critical": 2, "warning": 1, "ok": 1 },
  "changes": {
    "added": [],
    "fixed": [
      {
        "id": "f2",
        "code": "LINK_BROKEN",
        "kind": "link",
        "severity": "critical",
        "title": "Битая ссылка",
        "element": "<a href=\"/old\">old</a>",
        "recommendation": "Исправьте адрес ссылки.",
        "wcag": "2.4.4"
      }
    ],
    "remaining": 2,
    "scoreDelta": 12
  },
  "findings": [
    {
      "id": "f1",
      "code": "IMG_NO_ALT",
      "kind": "image",
      "severity": "critical",
      "title": "Нет alt у изображения",
      "element": "<img src=\"photo.svg\">",
      "recommendation": "Добавьте описание, что на картинке.",
      "wcag": "1.1.1"
    }
  ],
  "createdAt": "2026-11-22T10:00:00Z"
}
```

`version` — номер проверки этого URL. Первая проверка получает `1`, каждая
следующая — `+1`. `previousId` указывает на предыдущую завершённую проверку.

`changes` сравнивает текущие находки с предыдущей проверкой по паре
`code + element`:

- `added` — новые находки;
- `fixed` — находки, которых больше нет;
- `remaining` — сколько находок осталось;
- `scoreDelta` — изменение оценки относительно предыдущей проверки.

Для первой проверки URL `changes` отсутствует.

`status`:

- `queued` — создан;
- `running` — идёт проверка;
- `done` — готово;
- `failed` — ошибка; причина в поле `error`.

`score` — от 0 до 100 или `null`, пока проверка не закончена.

`kind` принимает `image`, `pdf`, `link`, `page`, `form`, `table`.

`pages` — сколько страниц удалось обойти. Если 0, скан получает `failed`.

`checkedLinks` — сколько уникальных ссылок проверили по HTTP.
`unverifiedLinks` — сколько ссылок не удалось проверить: сервер ответил 401/403/429/5xx
или не ответил. Такие ссылки не считаются битыми и не влияют на оценку.

## GET /api/scans

Журнал проверок, новые сверху.

Параметры:

- `url` — точный адрес сайта; без него возвращаются все проверки;
- `limit` — от 1 до 100, по умолчанию 20.

```json
[
  {
    "id": "0f7f0f2e-6a0b-4a3e-9f4e-7d1b2c3d4e5f",
    "url": "https://example.ru",
    "status": "done",
    "version": 2,
    "score": 42,
    "stats": { "critical": 2, "warning": 1, "ok": 1 },
    "findings": [],
    "createdAt": "2026-11-22T10:00:00Z"
  }
]
```

`findings` в списке может быть пустым: это краткая история, полный отчёт
запрашивается через `GET /api/scans/{id}`.

## Ошибки

```json
{
  "error": "invalid_input",
  "detail": "url must be http or https"
}
```

Коды:

- `400` — неверный запрос, приватный или неразрешимый адрес;
- `404` — скан не найден;
- `413` — тело запроса больше 1 МиБ;
- `429` — слишком много запросов;
- `500` — внутренняя ошибка.

## Правила контракта

- Поля не переименовываем без согласования с командой.
- Новые поля добавляем, старые не ломаем.
- Все даты — ISO 8601 UTC.
- Пустые массивы — `[]`, а не `null`.
- Секреты и ключи в ответы не попадают.

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
- `depth` — от 1 до 3;
- по умолчанию `1`.

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
  "score": 42,
  "stats": { "critical": 2, "warning": 1, "ok": 1 },
  "findings": [
    {
      "id": "f1",
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

`status`:

- `queued` — создан;
- `running` — идёт проверка;
- `done` — готово;
- `failed` — ошибка.

`score` — от 0 до 100 или `null`, пока проверка не закончена.

## Ошибки

```json
{
  "error": "invalid_input",
  "detail": "url must be http or https"
}
```

Коды:

- `400` — неверный запрос;
- `404` — скан не найден;
- `429` — слишком много запросов;
- `500` — внутренняя ошибка.

## Правила контракта

- Поля не переименовываем без согласования с командой.
- Новые поля добавляем, старые не ломаем.
- Все даты — ISO 8601 UTC.
- Пустые массивы — `[]`, а не `null`.
- Секреты и ключи в ответы не попадают.

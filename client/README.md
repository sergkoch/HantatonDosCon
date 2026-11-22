# Клиент

## Запуск

```bash
bun install
bun run dev
```

Открывается `http://localhost:5173`. Запросы `/api` уходят на `http://localhost:8080`.

## Проверки

```bash
bun run check
bun run test
bun run build
```

## Структура

```
src/
  app.tsx?          нет, точка входа в main.tsx
  App.tsx           маршруты
  api.ts            запросы к API
  types.ts          типы данных
  components/       Layout, Button, Card, Badge
  pages/            Home, Scan, Report, NotFound
  lib/              score, подписи
  styles.css        токены и стили
```

## Правила

- Строки интерфейса — по-русски.
- Цвета — только через `var(--color-...)`.
- Один шрифт.
- Каждый интерактивный элемент доступен с клавиатуры.
- У полей есть `label`.
- Ошибки показываются текстом, а не только цветом.
- Никаких `console.log` и `any`.

## Задачи

- `T04` лендинг.
- `T05` форма проверки.
- `T09` экран прогресса.
- `T18` отчёт.
- `T19` фильтры и CSV.
- `T23` доступность.

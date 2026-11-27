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
  main.ts              точка входа
  App.svelte           маршруты и ленивые страницы
  api.ts               запросы к API
  types.ts             типы данных
  lib/
    router.svelte.ts   история и навигация
    routes.ts          кеш и предзагрузка ленивых страниц
    i18n.svelte.ts     язык интерфейса
    messages.ts        словари RU/EN
    insights.ts        аналитика отчёта и сводка истории
    score.ts           оценка и подписи
    clientId.ts        анонимный ID браузера
    cookieConsent.svelte.ts  состояние баннера cookie
    theme.svelte.ts    тема
  components/          Layout, RouteSkeleton, CookieConsent, Button, Card, Badge
  pages/               Home, History, Report, NotFound
  styles.css           токены и стили
```

## Правила

- Строки интерфейса — в `messages.ts` на русском и английском, без хардкода.
- Цвета — только через `var(--...)`.
- Один шрифт.
- Каждый интерактивный элемент доступен с клавиатуры.
- У полей есть `label`.
- Ошибки показываются текстом, а не только цветом.
- Никаких `console.log` и `any`.

## Что есть

- главная с формой проверки и секцией «Как это работает»;
- история: фильтр, сводка, повторная проверка, очистка;
- отчёт: оценка, diff, аналитика, фильтры, CSV/PDF/JSON, CI-гейт;
- cookie: компактная карточка, «Что хранится», повторное открытие из футера;
- RU/EN, тёмная и светлая темы.

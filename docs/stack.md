# Стек: всё на Go, без Python

## Коротко

| Слой | Выбор | Почему |
| --- | --- | --- |
| API | **Go + Gin** | простой роутинг, один бинарник, понятен новичкам |
| Краулер | **Colly + goquery** | обход домена, глубина, `robots.txt`, конкурентность |
| HTML | **goquery** | знакомый CSS-селекторный API |
| PDF-структура | **go-pdfium** (PDFium в WASM) | текст, теги, `/Alt`, `Title`, `Lang`; лицензия BSD |
| PDF-валидация | **pdfcpu** | Apache 2.0, проверка файла, текст, fallback |
| OCR | **Tesseract CLI — план** | в образ не добавлен; PDF без текста сразу `critical` |
| ИИ | **HTTP-клиент на Go** | vision и текст через API, без SDK |
| БД | **PostgreSQL + pgx + goose** | надёжно, миграции, без ORM-магии |
| Отчёт | **go-pdf/fpdf + encoding/csv** | PDF и CSV на чистом Go |
| Интерфейс | **Svelte 5 + TypeScript + Vite** | браузер не умеет Go; Svelte компактнее и проще |
| Деплой | **Docker Compose + nginx** | один сервер, без Kubernetes |

## PDF без Python

Сейчас работают три вещи:

1. **Текстовый слой.** `go-pdfium` извлекает текст. Если текста нет — это скан,
   и находка сразу `critical`.
2. **Теги структуры.** PDFium отдаёт `FPDF_StructTree_*`, `FPDF_StructElement_GetAltText`,
   `FPDF_StructElement_GetType`.
3. **Метаданные.** `pdfcpu` — fallback для `Title`, `Lang` и признака тегов.

Рендер страницы и OCR — план, не подключены. Tesseract CLI понадобится, если
решим распознавать сканы: `tesseract page.png stdout -l rus+eng`.

`pdfcpu` — запасной вариант: быстрая проверка файла, текста и метаданных, если
PDFium не справился.

### Почему PDFium, а не pdfcpu

`pdfcpu` — отличный чистый Go, но он не даёт удобного доступа к дереву тегов и
`/Alt`. PDFium — движок Chrome, у него есть C API доступности. Поэтому:

- **PDFium** — глубокая проверка доступности;
- **pdfcpu** — быстрая базовая проверка и fallback.

### Почему Tesseract CLI, а не Go-обёртка

Это план на случай подключения OCR:

- CLI не тянет cgo в основной сервис.
- Падение OCR не роняет API.
- Проще обновлять и чинить.
- Русский язык — пакет `tesseract-ocr-rus`.

## ИИ

Один интерфейс:

```go
type Client interface {
    CheckImage(ctx context.Context, req ImageRequest) (ImageResult, error)
    CheckLink(ctx context.Context, req LinkRequest) (LinkResult, error)
}
```

Провайдер — любой OpenAI-совместимый API с vision. Ключ и модель — в `.env`.
Без ключа сервис работает: формальные проверки остаются, семантические
помечаются «не проверено».

## Отчёт

- **PDF:** `go-pdf/fpdf` — текст, отступы, встроенный Go-шрифт с кириллицей.
- **CSV:** стандартный `encoding/csv`.
- **JSON:** стандартный `encoding/json`.

Не тащим Chromium только ради PDF.

## Интерфейс

Браузерная часть не может быть на Go. Варианты:

- Svelte 5 + TypeScript + Vite — **выбран**;
- React — тоже можно, но в шаблоне Svelte;
- Go-шаблоны + HTMX — можно, но меньше учебных материалов.

Если команда решит сменить фреймворк — меняется только папка `client`, API остаётся.

## Что не берём

- Python, FastAPI, PyMuPDF.
- Kubernetes, Kafka, Redis.
- GraphQL, gRPC.
- Тяжёлые ORM.
- Облачный OCR как обязательный.
- Browser extension в MVP.

## Лицензии

| Компонент | Лицензия | Что значит |
| --- | --- | --- |
| PDFium | BSD-3 | можно использовать свободно |
| pdfcpu | Apache 2.0 | можно |
| Tesseract | Apache 2.0 | можно |
| Gin | MIT | можно |
| go-pdf/fpdf | MIT | можно |
| Svelte | MIT | можно |

## PDF-спайк на день 1

Проверено на реальном PDF:

- [x] открыть файл;
- [x] получить количество страниц;
- [x] извлечь текст;
- [x] найти `/StructTreeRoot` и `/MarkInfo`;
- [x] пройти по дереву и достать `/Alt`;
- [ ] отрендерить страницу в PNG;
- [ ] прогнать PNG через Tesseract (OCR — план).

PDFium в режиме WASM закрыл основную задачу: текст, теги, `/Alt`, `Title`,
`Lang`. Если PDFium падает на конкретном файле, остаётся fallback на `pdfcpu`.

## Docker для PDFium

`libpdfium.so`, `CGO_ENABLED=1` и `LD_LIBRARY_PATH` не нужны: PDFium
встроен в бинарник как WebAssembly (Wazero), образ собирается с `CGO_ENABLED=0`.

Если позже включим OCR, в образ нужно добавить `tesseract-ocr`,
`tesseract-ocr-rus` и `ca-certificates`.

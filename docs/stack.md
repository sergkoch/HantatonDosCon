# Стек: всё на Go, без Python

## Коротко

| Слой | Выбор | Почему |
| --- | --- | --- |
| API | **Go + Gin** | простой роутинг, один бинарник, понятен новичкам |
| Краулер | **Colly + goquery** | обход домена, глубина, `robots.txt`, конкурентность |
| HTML | **goquery** | знакомый CSS-селекторный API |
| PDF-структура | **go-pdfium** (PDFium) | текст, теги, `/Alt`, рендер; лицензия BSD |
| PDF-валидация | **pdfcpu** | Apache 2.0, проверка файла, текст, изображения |
| OCR | **Tesseract CLI** | офлайн, русский, без Python |
| ИИ | **HTTP-клиент на Go** | vision и текст через API, без SDK |
| БД | **PostgreSQL + pgx + goose** | надёжно, миграции, без ORM-магии |
| Отчёт | **maroto + encoding/csv** | PDF и CSV на чистом Go |
| Интерфейс | **React + TypeScript + Vite** | браузер не умеет Go; React проще для команды |
| Деплой | **Docker Compose + nginx** | один сервер, без Kubernetes |

## PDF без Python

Нужны четыре вещи:

1. **Текстовый слой.** `go-pdfium` извлекает текст. Если текста нет — это скан.
2. **Теги структуры.** PDFium отдаёт `FPDF_StructTree_*`, `FPDF_StructElement_GetAltText`,
   `FPDF_StructElement_GetType`.
3. **Рендер страницы.** PDFium рисует страницу в PNG.
4. **OCR.** Tesseract CLI распознаёт PNG: `tesseract page.png stdout -l rus+eng`.

`pdfcpu` — запасной вариант: быстрая проверка файла, текста и метаданных, если
PDFium не справился.

### Почему PDFium, а не pdfcpu

`pdfcpu` — отличный чистый Go, но он не даёт удобного доступа к дереву тегов и
`/Alt`. PDFium — движок Chrome, у него есть C API доступности. Поэтому:

- **PDFium** — глубокая проверка доступности;
- **pdfcpu** — быстрая базовая проверка и fallback.

### Почему Tesseract CLI, а не Go-обёртка

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

- **PDF:** `maroto/v2` — таблицы, заголовки, цвета.
- **CSV:** стандартный `encoding/csv`.
- **JSON:** стандартный `encoding/json`.

Не тащим Chromium только ради PDF.

## Интерфейс

Браузерная часть не может быть на Go. Варианты:

- React + TypeScript + Vite — **выбран**;
- SolidJS — тоже можно, но в шаблоне React;
- Go-шаблоны + HTMX — можно, но меньше учебных материалов.

Если команда уже знает Solid — меняется только папка `client`, API остаётся.

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
| maroto | MIT | можно |
| React | MIT | можно |

## PDF-спайк на день 1

Проверить на реальном PDF:

- [ ] открыть файл;
- [ ] получить количество страниц;
- [ ] извлечь текст;
- [ ] найти `/StructTreeRoot` и `/MarkInfo`;
- [ ] пройти по дереву и достать `/Alt`;
- [ ] отрендерить страницу в PNG;
- [ ] прогнать PNG через Tesseract.

Если всё получается — двигаемся на PDFium. Если нет — упрощаем проверки до
текстового слоя и метаданных, а глубину PDF/UA оставляем на бонус.

## Docker для PDFium

В образе нужны:

- `libpdfium.so`;
- `tesseract-ocr` и `tesseract-ocr-rus`;
- `ca-certificates`.

Сборка с `CGO_ENABLED=1`, `CGO_LDFLAGS`, `LD_LIBRARY_PATH`. Это задача для ответственного за PDF.

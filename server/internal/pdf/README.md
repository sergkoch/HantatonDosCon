# PDF без Python

## Задача

Понять, доступен ли PDF для скринридера.

## Что проверяем

1. Есть ли текстовый слой.
2. Есть ли теги структуры (`/StructTreeRoot`, `/MarkInfo /Marked`).
3. Есть ли `/Alt` у фигур.
4. Есть ли `Title` и `Lang`.
5. Можно ли отрендерить страницу для OCR.

## Инструменты

- `github.com/klippa-app/go-pdfium` — PDFium в режиме WebAssembly (Wazero),
  без cgo и без `pdfium.dll`. Лицензия BSD.
- `github.com/pdfcpu/pdfcpu` — fallback для `Title`, `Lang` и признака тегов.
  Apache 2.0.
- `tesseract` CLI — OCR, Apache 2.0.

## План

### День 1. Спайк

Результат: `docs/pdf-spike.md`. Выбран `go-pdfium` в режиме WebAssembly.

- [x] открыть PDF;
- [x] получить число страниц;
- [x] извлечь текст;
- [x] найти `/StructTreeRoot`;
- [x] пройти по дереву и достать `/Alt`;
- [x] отрендерить страницу в PNG;
- [ ] прогнать PNG через `tesseract page.png stdout -l rus+eng`.

### Дальше

- [x] реализовать `Analyzer`;
- [x] таймаут 30 секунд;
- [x] лимит 50 страниц и 20 МБ;
- [ ] отдельный процесс/контейнер, чтобы падение не роняло API;
- [x] `pdfcpu` для метаданных и fallback-признака тегов.

## Docker

В режиме WebAssembly `libpdfium.so`, `CGO_ENABLED=1` и `LD_LIBRARY_PATH` не
нужны: PDFium встроен в бинарник.

Для OCR в образе нужны:

- `tesseract-ocr`;
- `tesseract-ocr-rus`;
- `ca-certificates`.

## Готово, если

- [x] Скан без текстового слоя помечен critical.
- [x] PDF без тегов помечен warning.
- [x] `/Alt` считается.
- [x] Битый PDF не роняет сервис.
- [ ] Таймаут срабатывает (проверяется общим контекстом скана).

## Если PDFium падает на конкретном файле

Не роняем API. Оставляем fallback:

- `pdfcpu` для текстового слоя (эвристика по операторам), тегов, `Title`, `Lang`;
- Tesseract для сканов;
- глубину PDF/UA — на бонус.

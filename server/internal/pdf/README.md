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

- `github.com/klippa-app/go-pdfium` — PDFium, лицензия BSD.
- `github.com/pdfcpu/pdfcpu` — быстрый fallback, Apache 2.0.
- `tesseract` CLI — OCR, Apache 2.0.

## План

### День 1. Спайк

- [ ] открыть реальный PDF;
- [ ] получить число страниц;
- [ ] извлечь текст;
- [ ] найти `/StructTreeRoot`;
- [ ] пройти по дереву и достать `/Alt`;
- [ ] отрендерить страницу в PNG;
- [ ] прогнать PNG через `tesseract page.png stdout -l rus+eng`.

### Дальше

- реализовать `Analyzer`;
- таймаут 30 секунд;
- лимит 50 страниц и 20 МБ;
- отдельный процесс/контейнер, чтобы падение не роняло API;
- при ошибке PDFium — fallback на `pdfcpu`.

## Docker

В образе нужны:

- `libpdfium.so`;
- `tesseract-ocr`;
- `tesseract-ocr-rus`;
- `ca-certificates`.

Сборка: `CGO_ENABLED=1`, `CGO_LDFLAGS`, `LD_LIBRARY_PATH`.

## Готово, если

- [ ] Скан без текстового слоя помечен critical.
- [ ] PDF без тегов помечен warning.
- [ ] `/Alt` считается.
- [ ] Битый PDF не роняет сервис.
- [ ] Таймаут срабатывает.

## Если PDFium не собирается

Не тратим день. Оставляем:

- `pdfcpu` для текстового слоя и метаданных;
- Tesseract для сканов;
- глубину PDF/UA — на бонус.

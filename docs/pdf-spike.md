# PDF-спайк (T14)

Дата: 22.11.2026. Задача: понять, на чём делать PDF-проверки —
`go-pdfium` или `pdfcpu`.

## Короткий вывод

**Основной инструмент — `go-pdfium` в режиме WebAssembly.**

Он работает на Windows **без cgo, без MinGW и без `pdfium.dll`**: PDFium
скомпилирован в WebAssembly и встроен в Go-библиотеку. Этого хватает для всего,
что нужно сервису:

- текст;
- теги структуры и `/Alt`;
- рендер страницы в PNG для OCR и ИИ.

`pdfcpu` оставляем как **вспомогательный fallback**: он проще и быстрее, умеет
читать `Title`, `Lang` и признак тегов, но не умеет извлекать текст и рендерить
страницы.

Docker и установка C-компилятора для T14 **не нужны**.

## Что проверяли

Версии:

| Библиотека | Версия | Режим |
| --- | --- | --- |
| `github.com/klippa-app/go-pdfium` | v1.21.5 | `webassembly` (Wazero) |
| `github.com/pdfcpu/pdfcpu` | v0.16.4 | чистый Go |

Тестовые PDF лежат в `server/internal/pdf/testdata/`:

| Файл | Что внутри |
| --- | --- |
| `text.pdf` | текст, `Title`, `Lang`, без тегов |
| `scan.pdf` | только изображение, без текстового слоя |
| `tagged.pdf` | теги структуры, `/Alt`, `Title`, `Lang` |

Все три — синтетические, сгенерированы для спайка. Они маленькие, их можно
коммитить.

## Результаты

| Проверка | pdfcpu v0.16.4 | go-pdfium v1.21.5 (wasm) |
| --- | --- | --- |
| Количество страниц | ✅ | ✅ |
| Текстовый слой | ⚠️ только эвристика по операторам `Tj`/`TJ`, текст не извлекается | ✅ реальные символы через `FPDFText_*` |
| Признак тегов | ✅ `ctx.Tagged` | ✅ структура доступна через `FPDF_StructTree_*` |
| `/Alt` | ✅ обход `StructTreeRoot` | ✅ `FPDF_StructElement_GetAltText` |
| `Title` | ✅ | ✅ `FPDF_GetMetaText` |
| `Lang` | ✅ из каталога | ❌ отдельного API нет, брать из pdfcpu или каталога |
| Рендер в PNG | ❌ | ✅ `RenderPageInDPI` |
| Сборка на Windows | ✅ | ✅ |
| CGO / C-компилятор | не нужен | не нужен в режиме `webassembly` |
| Внешние файлы (`pdfium.dll`) | не нужны | не нужны, wasm встроен |
| Размер бинарника | небольшой | +~6 МБ (встроенный `pdfium.wasm`) |
| Скорость | быстрее | ~2× медленнее нативного cgo (по документации) |

Фактический прогон:

```
file: text.pdf
  pdfcpu: pages=1 tagged=false title="Sample text PDF" lang=ru-RU
  pdfium: pages=1 textChars=21 altTexts=0
  pdfium: render=625x417 saved=pdfspike-out/text.png

file: scan.pdf
  pdfcpu: pages=1 tagged=false title="" lang=<nil>
  pdfium: pages=1 textChars=0 altTexts=0
  pdfium: render=625x251 saved=pdfspike-out/scan.png

file: tagged.pdf
  pdfcpu: pages=1 tagged=true title="Tagged sample PDF" lang=ru-RU
  pdfium: pages=1 textChars=21 altTexts=1
  pdfium: render=625x417 saved=pdfspike-out/tagged.png
```

Скан без текста определяется как `textChars=0`, PDF с `/Alt` — как
`altTexts=1`. Это именно то, что нужно проверкам T15.

## Как воспроизвести

Из папки `server`:

```bash
go run ./cmd/pdfspike
```

Или по конкретному файлу:

```bash
go run ./cmd/pdfspike internal/pdf/testdata/tagged.pdf
```

PNG-файлы появятся в `pdfspike-out/`.

Команда `cmd/pdfspike` — временная, для спайка. После T15 её можно удалить.

## Ограничения и риски

1. **Размер.** WebAssembly-модуль встроен в бинарник и добавляет ~6 МБ.
2. **Скорость.** Режим wasm примерно в 2 раза медленнее нативного cgo.
   Для проверки PDF это приемлемо, но стоит замерить на реальном документе.
3. **Память.** Wazero не возвращает память системе, пока экземпляр не закрыт.
   Нужен пул с ограничением и закрытием экземпляров после скана.
4. **Лимит рендера.** Один рендер ограничен ~536 мегапикселями. Для A4 это
   огромный запас; при необходимости можно рендерить по частям через `Crop`.
5. **Пути.** В wasm-режиме пути должны быть POSIX и абсолютные. Проще передавать
   PDF в память (`File: &data`), как в спайке.
6. **`Lang`.** PDFium не отдаёт язык отдельным API. Берём из pdfcpu или читаем
   `/Lang` из каталога.
7. **Битые PDF.** Нужно проверить отдельно в T15: ошибка PDFium не должна
   ронять API.

## Что делать в T15

1. Реализовать `pdf.Analyzer` из `server/internal/pdf/types.go` на `go-pdfium`
   (webassembly).
2. `Lang` и `Title` брать через pdfcpu или из каталога; `Tagged` — через
   структуру PDFium.
3. Ограничения: 30 секунд, 50 страниц, 20 МБ.
4. Рендерить только те страницы, где нужен OCR или ИИ.
5. Tesseract для сканов подключать отдельным шагом, после базовых проверок.
6. Добавить тесты на `text.pdf`, `scan.pdf`, `tagged.pdf` и на битый файл.

## Ссылки

- `github.com/klippa-app/go-pdfium` — WebAssembly-режим и примеры.
- `NATIVE_VS_WEBASSEMBLY.md` в репозитории go-pdfium.
- `github.com/pdfcpu/pdfcpu` — `pkg/api`, `pkg/pdfcpu/model`.

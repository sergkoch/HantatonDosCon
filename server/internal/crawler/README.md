# Краулер

## Задача

Обойти сайт в пределах домена и собрать страницы, изображения, PDF и ссылки.

## Инструменты

- `github.com/gocolly/colly/v2`
- `github.com/PuerkitoBio/goquery`

## Что сделать

1. Принять стартовый URL и глубину.
2. Ограничить домен.
3. Уважать `robots.txt`.
4. Поставить лимиты: 50 страниц, 4 параллельно, задержка 200 мс.
5. Собрать:
   - `title` страницы;
   - `img` (`src`, `alt`);
   - `a[href]` (текст, ссылка);
   - ссылки на `.pdf`.
6. Вернуть `[]Page`.

## Подсказка

```go
collector := colly.NewCollector(
    colly.AllowedDomains(host),
    colly.MaxDepth(depth),
    colly.Async(true),
)

collector.Limit(&colly.LimitRule{
    DomainGlob:  "*",
    Parallelism: 4,
    Delay:       200 * time.Millisecond,
})

collector.OnHTML("img", func(e *colly.HTMLElement) {
    // e.Attr("src"), e.Attr("alt")
})

collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
    // e.Attr("href"), e.Text
})
```

## Готово, если

- [ ] Глубина 1 и 2 работают.
- [ ] Чужие домены не обходятся.
- [ ] `robots.txt` учитывается.
- [ ] Таймаут не вешает сервис.
- [ ] На демо-сайте находится 10+ элементов.

## Осторожно

- Не сканировать чужие сайты без разрешения.
- Не ставить параллельность больше 4.
- Не ходить на приватные IP и `localhost` из продакшена.

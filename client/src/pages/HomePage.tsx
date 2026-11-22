import { type FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createScan } from "../api";
import { Button } from "../components/Button";
import { Card } from "../components/Card";

export function HomePage() {
  const [url, setUrl] = useState("");
  const [depth, setDepth] = useState(1);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const navigate = useNavigate();

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setBusy(true);

    try {
      const scan = await createScan(url, depth);
      navigate(`/scan/${scan.id}`);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Не удалось запустить проверку",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="container">
      <section className="hero">
        <p className="hero__eyebrow">Проверка доступности</p>
        <h1 className="hero__title">Сайт, который понимают все</h1>
        <p className="hero__lead">
          Проверьте изображения, PDF и ссылки. Найдите не только формальные
          ошибки, но и смысловые: описание не совпадает с картинкой, ссылка
          ведёт не туда.
        </p>

        <form className="form" onSubmit={onSubmit} aria-describedby="scan-hint">
          <div className="field">
            <label className="field__label" htmlFor="url">
              Адрес сайта
            </label>
            <input
              className="input"
              id="url"
              name="url"
              type="url"
              inputMode="url"
              placeholder="https://example.ru"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              required
            />
            <p className="field__hint" id="scan-hint">
              Только http или https. Начните с главной страницы.
            </p>
          </div>

          <div className="field field--small">
            <label className="field__label" htmlFor="depth">
              Глубина
            </label>
            <select
              className="select"
              id="depth"
              name="depth"
              value={depth}
              onChange={(event) => setDepth(Number(event.target.value))}
            >
              <option value={1}>1 — только страница</option>
              <option value={2}>2 — страница и ссылки</option>
              <option value={3}>3 — глубже</option>
            </select>
          </div>

          <Button type="submit" disabled={busy}>
            {busy ? "Запускаем…" : "Проверить"}
          </Button>
        </form>

        {error ? (
          <p className="field__error" role="alert">
            {error}
          </p>
        ) : null}
      </section>

      <section className="section" id="how" aria-labelledby="how-title">
        <h2 className="section__title" id="how-title">
          Что проверяем
        </h2>
        <div className="grid">
          <Card title="Изображения">
            Есть ли <code className="code">alt</code>, осмыслен ли он и
            совпадает ли с картинкой.
          </Card>
          <Card title="PDF">
            Есть ли текстовый слой, теги структуры и альтернативная версия.
          </Card>
          <Card title="Ссылки">Работают ли они и понятно ли, куда ведут.</Card>
        </div>
      </section>
    </div>
  );
}

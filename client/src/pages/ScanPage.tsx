import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getScan } from "../api";
import { scoreLabel } from "../lib/score";
import type { Scan } from "../types";

export function ScanPage() {
  const { id } = useParams<{ id: string }>();
  const [scan, setScan] = useState<Scan | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) {
      return;
    }

    const scanId = id;
    let active = true;
    let timer: number | undefined;

    async function load() {
      try {
        const next = await getScan(scanId);
        if (!active) {
          return;
        }

        setScan(next);

        if (next.status !== "done" && next.status !== "failed") {
          timer = window.setTimeout(load, 1500);
        }
      } catch (cause) {
        if (!active) {
          return;
        }

        setError(
          cause instanceof Error ? cause.message : "Не удалось получить скан",
        );
      }
    }

    load();

    return () => {
      active = false;
      if (timer) {
        window.clearTimeout(timer);
      }
    };
  }, [id]);

  if (error) {
    return (
      <div className="container">
        <p className="field__error" role="alert">
          {error}
        </p>
      </div>
    );
  }

  if (!scan) {
    return (
      <div className="container">
        <p className="status" aria-live="polite">
          Загружаем скан…
        </p>
      </div>
    );
  }

  const done = scan.status === "done";
  const failed = scan.status === "failed";
  const statusText = done
    ? "Проверка завершена"
    : failed
      ? "Проверка не удалась"
      : "Проверяем страницы и медиа…";

  return (
    <div className="container">
      <section className="section" aria-labelledby="scan-title">
        <p className="hero__eyebrow">Проверка</p>
        <h1 className="section__title" id="scan-title">
          {scan.url}
        </h1>

        <p className="status" aria-live="polite">
          {statusText}
        </p>

        <div className="summary">
          <div className="summary__item">
            <span className="summary__value">{scan.score ?? "—"}</span>
            <span className="summary__label">Оценка</span>
          </div>
          <div className="summary__item">
            <span className="summary__value">{scan.stats.critical}</span>
            <span className="summary__label">Критично</span>
          </div>
          <div className="summary__item">
            <span className="summary__value">{scan.stats.warning}</span>
            <span className="summary__label">Предупреждения</span>
          </div>
          <div className="summary__item">
            <span className="summary__value">{scan.stats.ok}</span>
            <span className="summary__label">В порядке</span>
          </div>
        </div>

        <p className="status">{scoreLabel(scan.score)}</p>

        {done ? (
          <Link className="button button--primary" to={`/report/${scan.id}`}>
            Открыть отчёт
          </Link>
        ) : null}
      </section>
    </div>
  );
}

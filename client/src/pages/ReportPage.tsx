import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { getScan } from "../api";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { kindLabel } from "../lib/score";
import type { Scan, Severity } from "../types";

const filters: Array<{ value: "all" | Severity; label: string }> = [
  { value: "all", label: "Все" },
  { value: "critical", label: "Критично" },
  { value: "warning", label: "Предупреждения" },
  { value: "ok", label: "В порядке" },
];

export function ReportPage() {
  const { id } = useParams<{ id: string }>();
  const [scan, setScan] = useState<Scan | null>(null);
  const [error, setError] = useState("");
  const [filter, setFilter] = useState<"all" | Severity>("all");

  useEffect(() => {
    if (!id) {
      return;
    }

    let active = true;

    getScan(id)
      .then((next) => {
        if (active) {
          setScan(next);
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(
            cause instanceof Error
              ? cause.message
              : "Не удалось получить отчёт",
          );
        }
      });

    return () => {
      active = false;
    };
  }, [id]);

  const findings = useMemo(() => {
    if (!scan) {
      return [];
    }

    return filter === "all"
      ? scan.findings
      : scan.findings.filter((finding) => finding.severity === filter);
  }, [scan, filter]);

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
          Загружаем отчёт…
        </p>
      </div>
    );
  }

  return (
    <div className="container">
      <section className="section" aria-labelledby="report-title">
        <p className="hero__eyebrow">Отчёт</p>
        <h1 className="section__title" id="report-title">
          {scan.url}
        </h1>
        <p className="status">
          Оценка: {scan.score ?? "—"} · находок: {scan.findings.length}
        </p>

        <fieldset className="filters">
          <legend className="visually-hidden">Фильтр по критичности</legend>
          {filters.map((item) => (
            <button
              key={item.value}
              type="button"
              className={`filter ${filter === item.value ? "filter--active" : ""}`.trim()}
              onClick={() => setFilter(item.value)}
              aria-pressed={filter === item.value}
            >
              {item.label}
            </button>
          ))}
        </fieldset>

        <div className="table-wrap">
          <table className="table">
            <caption className="visually-hidden">Находки проверки</caption>
            <thead>
              <tr>
                <th scope="col">Критичность</th>
                <th scope="col">Тип</th>
                <th scope="col">Что не так</th>
                <th scope="col">Рекомендация</th>
              </tr>
            </thead>
            <tbody>
              {findings.map((finding) => (
                <tr key={finding.id}>
                  <td>
                    <Badge severity={finding.severity} />
                  </td>
                  <td>{kindLabel(finding.kind)}</td>
                  <td>
                    <strong>{finding.title}</strong>
                    <code className="code code--block">{finding.element}</code>
                    <span className="muted">WCAG {finding.wcag}</span>
                  </td>
                  <td>{finding.recommendation}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {findings.length === 0 ? (
          <p className="empty">По этому фильтру находок нет.</p>
        ) : null}

        <div className="actions">
          <Button type="button" onClick={() => downloadCsv(scan)}>
            Скачать CSV
          </Button>
        </div>
      </section>
    </div>
  );
}

function downloadCsv(scan: Scan) {
  const rows = [
    ["severity", "kind", "title", "element", "recommendation", "wcag"],
    ...scan.findings.map((finding) => [
      finding.severity,
      finding.kind,
      finding.title,
      finding.element,
      finding.recommendation,
      finding.wcag,
    ]),
  ];

  const csv = rows
    .map((row) =>
      row.map((cell) => `"${cell.replaceAll('"', '""')}"`).join(","),
    )
    .join("\n");
  const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");

  link.href = url;
  link.download = `report-${scan.id}.csv`;
  link.click();
  URL.revokeObjectURL(url);
}

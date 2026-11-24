<script lang="ts">
  import { getScan } from "../api";
  import Badge from "../components/Badge.svelte";
  import Button from "../components/Button.svelte";
  import { navigate } from "../lib/router.svelte";
  import { kindLabel, scoreLabel } from "../lib/score";
  import type { Scan, Severity } from "../types";

  let { id }: { id: string } = $props();

  let scan = $state<Scan | null>(null);
  let error = $state("");
  let filter = $state<"all" | Severity>("all");

  const circumference = 2 * Math.PI * 68;

  $effect(() => {
    let active = true;

    getScan(id)
      .then((next) => {
        if (active) scan = next;
      })
      .catch((cause: unknown) => {
        if (active) {
          error =
            cause instanceof Error ? cause.message : "Не удалось получить отчёт";
        }
      });

    return () => {
      active = false;
    };
  });

  const filters: Array<{ value: "all" | Severity; label: string }> = [
    { value: "all", label: "Все" },
    { value: "critical", label: "Критично" },
    { value: "warning", label: "Предупреждения" },
    { value: "ok", label: "В порядке" },
  ];

  const findings = $derived(
    filter === "all"
      ? (scan?.findings ?? [])
      : (scan?.findings.filter((finding) => finding.severity === filter) ?? []),
  );

  function downloadCsv(current: Scan) {
    const rows = [
      ["severity", "kind", "title", "element", "recommendation", "wcag"],
      ...current.findings.map((finding) => [
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
    link.download = `report-${current.id}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  }
</script>

{#if error}
  <div class="container">
    <p class="field__error" role="alert">{error}</p>
  </div>
{:else if scan}
  {@const current = scan}
  <div class="container">
    <section class="section" aria-labelledby="report-title">
      <div class="report-head">
        <div class="report-head__main">
          <p class="hero__eyebrow">Отчёт · проверка v{current.version}</p>
          <h1 class="report-head__title" id="report-title">{current.url}</h1>
          <p class="status">
            Находок: {current.findings.length} · {scoreLabel(current.score)}
            {#if current.previousId}
              ·
              <a
                href={`/report/${current.previousId}`}
                onclick={(event) => {
                  event.preventDefault();
                  navigate(`/report/${current.previousId}`);
                }}
              >
                предыдущая проверка
              </a>
            {/if}
          </p>
          <div class="report-head__stats">
            <span class="badge badge--critical">
              {current.stats.critical} критично
            </span>
            <span class="badge badge--warning">
              {current.stats.warning} предупр.
            </span>
            <span class="badge badge--ok">{current.stats.ok} в порядке</span>
          </div>
          <p class="report-coverage">
            Проверено страниц: {current.pages ?? 0}.
            {#if current.checkedLinks}
              Ссылок проверено: {current.checkedLinks}, не удалось проверить:
              {current.unverifiedLinks ?? 0}.
            {/if}
            Оценка учитывает alt, текст и доступность ссылок, подписи полей,
            заголовки, таблицы и PDF (текст, теги, /Alt, Title, Lang).
            Контраст, клавиатура и ARIA не проверяются.
          </p>
        </div>

        {#if current.score !== null}
          <div
            class="score-ring"
            role="img"
            aria-label={`Оценка ${current.score} из 100`}
          >
            <svg viewBox="0 0 160 160" width="160" height="160" aria-hidden="true">
              <circle
                class="score-ring__track"
                cx="80"
                cy="80"
                r="68"
                fill="none"
                stroke-width="12"
              />
              <circle
                class={`score-ring__value score-ring__value--${
                  current.score >= 80
                    ? "ok"
                    : current.score >= 50
                      ? "warning"
                      : "critical"
                }`}
                cx="80"
                cy="80"
                r="68"
                fill="none"
                stroke-width="12"
                stroke-linecap="round"
                stroke-dasharray={circumference}
                stroke-dashoffset={circumference * (1 - current.score / 100)}
              />
            </svg>
            <div class="score-ring__center">
              <span class="score-ring__num">{current.score}</span>
              <span class="score-ring__label">из 100</span>
            </div>
          </div>
        {/if}
      </div>

      <section class="section" aria-labelledby="changes-title">
        <h2 class="section__title" id="changes-title">
          Что изменилось с прошлой проверки
        </h2>

        {#if current.changes}
          <div class="report-head__stats">
            <span
              class={`badge ${current.changes.added.length > 0 ? "badge--critical" : "badge--ok"}`}
            >
              {current.changes.added.length} новых
            </span>
            <span
              class={`badge ${current.changes.fixed.length > 0 ? "badge--ok" : ""}`.trim()}
            >
              {current.changes.fixed.length} исправлено
            </span>
            <span class="badge">
              {current.changes.remaining} осталось
            </span>
            {#if current.changes.scoreDelta !== undefined}
              <span class="badge">
                {current.changes.scoreDelta >= 0 ? "+" : ""}{current.changes.scoreDelta}
                к оценке
              </span>
            {/if}
          </div>

          {#if current.changes.added.length > 0}
            <h3>Новые проблемы</h3>
            <ul class="links">
              {#each current.changes.added as finding (finding.id)}
                <li>
                  <Badge severity={finding.severity} />
                  {finding.title}
                </li>
              {/each}
            </ul>
          {/if}

          {#if current.changes.fixed.length > 0}
            <h3>Исправлено</h3>
            <ul class="links">
              {#each current.changes.fixed as finding (finding.id)}
                <li>
                  <Badge severity={finding.severity} />
                  {finding.title}
                </li>
              {/each}
            </ul>
          {/if}

          {#if current.changes.added.length === 0 && current.changes.fixed.length === 0}
            <p class="status">
              Новых проблем нет, исправлений тоже нет — состав находок не изменился.
            </p>
          {/if}
        {:else}
          <p class="status">
            Это первая проверка этого адреса. Следующая проверка покажет, что
            изменилось.
          </p>
        {/if}
      </section>

      <fieldset class="filters">
        <legend class="visually-hidden">Фильтр по критичности</legend>
        {#each filters as item (item.value)}
          <button
            type="button"
            class={`filter ${filter === item.value ? "filter--active" : ""}`.trim()}
            onclick={() => (filter = item.value)}
            aria-pressed={filter === item.value}
          >
            {item.label}
          </button>
        {/each}
      </fieldset>

      <div class="table-wrap">
        <table class="table">
          <caption class="visually-hidden">Находки проверки</caption>
          <thead>
            <tr>
              <th scope="col">Критичность</th>
              <th scope="col">Тип</th>
              <th scope="col">Что не так</th>
              <th scope="col">Рекомендация</th>
            </tr>
          </thead>
          <tbody>
            {#each findings as finding (finding.id)}
              <tr>
                <td>
                  <Badge severity={finding.severity} />
                </td>
                <td>{kindLabel(finding.kind)}</td>
                <td>
                  <strong>{finding.title}</strong>
                  <code class="code code--block">{finding.element}</code>
                  <span class="muted">WCAG {finding.wcag}</span>
                </td>
                <td>{finding.recommendation}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if findings.length === 0}
        <p class="empty">По этому фильтру находок нет.</p>
      {/if}

      <div class="actions">
        <a
          class="button button--primary"
          href={`/api/scans/${current.id}/report.pdf`}
          download
        >
          Скачать PDF
        </a>
        <Button variant="ghost" type="button" onclick={() => downloadCsv(current)}>
          Скачать CSV
        </Button>
      </div>
    </section>
  </div>
{:else}
  <div class="container">
    <div class="scan-visual" aria-hidden="true">
      <span class="scan-visual__pulse"></span>
      <span class="scan-visual__pulse"></span>
      <span class="scan-visual__pulse"></span>
    </div>
    <p class="status" aria-live="polite">Загружаем отчёт…</p>
  </div>
{/if}

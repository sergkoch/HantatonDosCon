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
  let displayScore = $state(0);

  const circumference = 2 * Math.PI * 68;

  // Опрашиваем скан, пока он не завершён (раньше это делала ScanPage).
  $effect(() => {
    scan = null;
    error = "";
    displayScore = 0;

    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;

    async function load() {
      try {
        const next = await getScan(id);
        if (!active) return;

        scan = next;

        if (next.status !== "done" && next.status !== "failed") {
          timer = setTimeout(load, 1500);
        }
      } catch (cause) {
        if (!active) return;
        error =
          cause instanceof Error ? cause.message : "Не удалось получить отчёт";
      }
    }

    load();

    return () => {
      active = false;
      if (timer) clearTimeout(timer);
    };
  });

  // Анимация «набора» оценки от 0 до итоговой.
  $effect(() => {
    const current = scan;
    if (!current || current.status !== "done" || current.score === null) {
      return;
    }

    const target = current.score;
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduce) {
      displayScore = target;
      return;
    }

    displayScore = 0;
    const startedAt = performance.now();
    const duration = 1100;
    let frame = 0;

    const tick = (now: number) => {
      const progress = Math.min(1, (now - startedAt) / duration);
      const eased = 1 - Math.pow(1 - progress, 3);
      displayScore = Math.round(target * eased);
      if (progress < 1) {
        frame = requestAnimationFrame(tick);
      }
    };

    // Страховка: если rAF приостановлен (вкладка скрыта), всё равно доходим до цели.
    const settle = setTimeout(() => {
      displayScore = target;
    }, duration + 250);

    frame = requestAnimationFrame(tick);
    return () => {
      cancelAnimationFrame(frame);
      clearTimeout(settle);
    };
  });

  const filters: Array<{ value: "all" | Severity; label: string }> = [
    { value: "all", label: "Все" },
    { value: "critical", label: "Критично" },
    { value: "warning", label: "Предупреждения" },
  ];

  // В списке — только проблемы (без «в порядке»).
  const problems = $derived(
    (scan?.findings ?? []).filter((finding) => finding.severity !== "ok"),
  );

  const findings = $derived(
    filter === "all"
      ? problems
      : problems.filter((finding) => finding.severity === filter),
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
{:else if scan && scan.status === "failed"}
  <div class="container">
    <section class="section" aria-labelledby="failed-title">
      <p class="hero__eyebrow">Проверка</p>
      <h1 class="section__title" id="failed-title">{scan.url}</h1>
      <p class="field__error" role="alert">
        {scan.error ?? "Проверка не удалась"}
      </p>
      <div class="actions">
        <a
          class="button button--primary"
          href="/"
          onclick={(event) => {
            event.preventDefault();
            navigate("/");
          }}
        >
          Попробовать снова
        </a>
      </div>
    </section>
  </div>
{:else if scan && scan.status === "done"}
  {@const current = scan}
  <div class="container">
    <section class="section" aria-labelledby="report-title">
      <div class="report-head">
        <div class="report-head__main">
          <p class="hero__eyebrow">Отчёт · проверка v{current.version}</p>
          <h1 class="report-head__title" id="report-title">{current.url}</h1>
          <p class="status">
            Находок: {problems.length} · {scoreLabel(current.score)}
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
                stroke-dashoffset={circumference * (1 - displayScore / 100)}
              />
            </svg>
            <div class="score-ring__center">
              <span class="score-ring__num">{displayScore}</span>
              <span class="score-ring__label">из 100</span>
            </div>
          </div>
        {/if}
      </div>

      {#if current.changes}
        <section class="section" aria-labelledby="changes-title">
          <h2 class="section__title" id="changes-title">
            Что изменилось с прошлой проверки
          </h2>

          <div class="report-head__stats">
            <span
              class={`badge ${current.changes.added.length > 0 ? "badge--critical" : ""}`.trim()}
            >
              {current.changes.added.length} новых
            </span>
            <span class="badge">
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
        </section>
      {/if}

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
            {#each findings as finding, index (finding.id)}
              <tr
                class="report-row"
                style={`animation-delay: ${Math.min(index * 45, 500)}ms`}
              >
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
        <Button type="button" onclick={() => downloadCsv(current)}>
          Скачать CSV
        </Button>
        <a
          class="button button--primary"
          href={`/api/scans/${current.id}/pdf`}
          download
        >
          Скачать PDF
        </a>
      </div>
    </section>
  </div>
{:else}
  <div class="container">
    <section class="section" aria-labelledby="scanning-title">
      <p class="hero__eyebrow">Проверка</p>
      <h1 class="section__title" id="scanning-title">Проверяем сайт…</h1>
      <div class="scan-visual" aria-hidden="true">
        <span class="scan-visual__pulse"></span>
        <span class="scan-visual__pulse"></span>
        <span class="scan-visual__pulse"></span>
        <span class="scan-visual__core">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
            <path
              d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linejoin="round"
            />
            <circle
              cx="12"
              cy="12"
              r="3"
              stroke="currentColor"
              stroke-width="1.6"
            />
          </svg>
        </span>
      </div>
      <div class="progress" aria-hidden="true">
        <div class="progress__bar"></div>
      </div>
      <p class="status" aria-live="polite">Собираем страницы, ссылки и медиа…</p>
    </section>
  </div>
{/if}

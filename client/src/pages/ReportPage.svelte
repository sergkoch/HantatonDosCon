<script lang="ts">
  import { tick } from "svelte";
  import { getScan, getScanPdf } from "../api";
  import Badge from "../components/Badge.svelte";
  import Button from "../components/Button.svelte";
  import { isPlainClick, navigate } from "../lib/router.svelte";
  import { getLocale, t } from "../lib/i18n.svelte";
  import { kindLabel, scoreLabel } from "../lib/score";
  import type { Scan, Severity } from "../types";

  let { id }: { id: string } = $props();

  let scan = $state<Scan | null>(null);
  let error = $state("");
  let pdfError = $state("");
  let pdfOpen = $state(false);
  let pdfButton: HTMLButtonElement | undefined = $state();
  let pdfMenu: HTMLDivElement | undefined = $state();
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
          cause instanceof Error ? cause.message : t("report.error");
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

  const filters = $derived<Array<{ value: "all" | Severity; label: string }>>([
    { value: "all", label: t("report.filterAll") },
    { value: "critical", label: t("report.filterCritical") },
    { value: "warning", label: t("report.filterWarning") },
  ]);

  // В списке — только проблемы (без «в порядке»).
  const problems = $derived(
    (scan?.findings ?? []).filter((finding) => finding.severity !== "ok"),
  );

  const findings = $derived(
    filter === "all"
      ? problems
      : problems.filter((finding) => finding.severity === filter),
  );

  function safeCell(value: string): string {
    const text = value ?? "";
    return /^[=+\-@]/.test(text) ? `'${text}` : text;
  }

  const severityKeys = {
    critical: "severity.critical",
    warning: "severity.warning",
    ok: "severity.ok",
  } as const;

  async function togglePdfMenu() {
    pdfOpen = !pdfOpen;
    if (pdfOpen) {
      await tick();
      pdfMenu?.querySelector<HTMLButtonElement>("button")?.focus();
    }
  }

  function onWindowClick(event: MouseEvent) {
    if (!pdfOpen) return;
    const target = event.target as HTMLElement | null;
    if (!target?.closest(".pdf-download")) {
      pdfOpen = false;
    }
  }

  function onWindowKeydown(event: KeyboardEvent) {
    if (pdfOpen && event.key === "Escape") {
      pdfOpen = false;
      pdfButton?.focus();
    }
  }

  async function downloadPdf(current: Scan, lang: "ru" | "en") {
    pdfError = "";
    pdfOpen = false;
    pdfButton?.focus();

    try {
      const blob = await getScanPdf(current.id, lang);
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");

      link.href = url;
      link.download = `report-${current.id}-${lang}.pdf`;
      link.click();
      URL.revokeObjectURL(url);
    } catch (cause) {
      pdfError = cause instanceof Error ? cause.message : t("report.pdfError");
    }
  }

  function downloadCsv(current: Scan) {
    const rows = [
      [
        t("csv.severity"),
        t("csv.kind"),
        t("csv.title"),
        t("csv.element"),
        t("csv.recommendation"),
        t("csv.wcag"),
      ],
      ...current.findings.map((finding) => [
        t(severityKeys[finding.severity]),
        kindLabel(finding.kind, getLocale()),
        finding.title,
        finding.element,
        finding.recommendation,
        finding.wcag,
      ]),
    ];

    const csv = rows
      .map((row) =>
        row
          .map((cell) => `"${safeCell(cell).replaceAll('"', '""')}"`)
          .join(","),
      )
      .join("\n");

    const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");

    link.href = url;
    link.download = `report-${current.id}.csv`;
    document.body.append(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 100);
  }
</script>

<svelte:window onclick={onWindowClick} onkeydown={onWindowKeydown} />

{#if error}
  <div class="container">
    <p class="field__error" role="alert">{error}</p>
  </div>
{:else if scan && scan.status !== "done" && scan.status !== "failed"}
  <div class="container">
    <p class="status" role="status">{t("report.running")}</p>
  </div>
{:else if scan && scan.status === "failed"}
  <div class="container">
    <section class="section" aria-labelledby="failed-title">
      <p class="hero__eyebrow">{t("report.scanEyebrow")}</p>
      <h1 class="section__title" id="failed-title">{scan.url}</h1>
      <p class="field__error" role="alert">
        {scan.error ?? t("report.failedFallback")}
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
          {t("report.tryAgain")}
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
          <p class="hero__eyebrow">
            {t("report.eyebrow", { version: current.version })}
          </p>
          <h1 class="report-head__title" id="report-title">{current.url}</h1>
          <p class="status">
            {t("report.findingsCount", { count: problems.length })} · {scoreLabel(
              current.score,
              getLocale(),
            )}
            {#if current.previousId}
              ·
              <a
                href={`/report/${current.previousId}`}
                onclick={(event) => {
                  if (!isPlainClick(event)) return;
                  event.preventDefault();
                  navigate(`/report/${current.previousId}`);
                }}
              >
                {t("report.previous")}
              </a>
            {/if}
          </p>
          <div class="report-head__stats">
            <span class="badge badge--critical">
              {t("report.critical", { count: current.stats.critical })}
            </span>
            <span class="badge badge--warning">
              {t(
                current.stats.warning === 1
                  ? "report.warningOne"
                  : "report.warning",
                { count: current.stats.warning },
              )}
            </span>
          </div>
          <p class="report-coverage">
            {t("report.coveragePages", { count: current.pages ?? 0 })}
            {#if current.checkedLinks}
              {t("report.coverageLinks", {
                checked: current.checkedLinks,
                unverified: current.unverifiedLinks ?? 0,
              })}
            {/if}
            {t("report.coverageNote")}
          </p>
        </div>

        {#if current.score !== null}
          <div
            class="score-ring"
            role="img"
            aria-label={t("report.scoreAria", { score: current.score })}
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
              <span class="score-ring__label">{t("report.scoreOf")}</span>
            </div>
          </div>
        {/if}
      </div>

      {#if current.changes}
        <section class="section" aria-labelledby="changes-title">
          <h2 class="section__title" id="changes-title">
            {t("report.changesTitle")}
          </h2>

          <div class="report-head__stats">
            <span
              class={`badge ${current.changes.added.length > 0 ? "badge--critical" : ""}`.trim()}
            >
              {t("report.added", { count: current.changes.added.length })}
            </span>
            <span class="badge">
              {t("report.fixed", { count: current.changes.fixed.length })}
            </span>
            <span class="badge">
              {t("report.remaining", { count: current.changes.remaining })}
            </span>
            {#if current.changes.scoreDelta !== undefined}
              <span class="badge">
                {t("report.delta", {
                  delta: `${current.changes.scoreDelta >= 0 ? "+" : ""}${current.changes.scoreDelta}`,
                })}
              </span>
            {/if}
          </div>

          {#if current.changes.added.length > 0}
            <h3>{t("report.newProblems")}</h3>
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
            <h3>{t("report.fixedHeading")}</h3>
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
              {t("report.noChanges")}
            </p>
          {/if}
        </section>
      {/if}

      <fieldset class="filters">
        <legend class="visually-hidden">{t("report.filterLegend")}</legend>
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
          <caption class="visually-hidden">{t("report.tableCaption")}</caption>
          <thead>
            <tr>
              <th scope="col">{t("report.thSeverity")}</th>
              <th scope="col">{t("report.thKind")}</th>
              <th scope="col">{t("report.thProblem")}</th>
              <th scope="col">{t("report.thRecommendation")}</th>
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
                <td>{kindLabel(finding.kind, getLocale())}</td>
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
        <p class="empty">{t("report.emptyFilter")}</p>
      {/if}

      <div class="actions">
        <Button type="button" onclick={() => downloadCsv(current)}>
          {t("report.downloadCsv")}
        </Button>

        <div class="pdf-download">
          <button
            bind:this={pdfButton}
            class="button button--primary pdf-download__button"
            type="button"
            aria-haspopup="menu"
            aria-expanded={pdfOpen}
            aria-controls="pdf-menu"
            onclick={togglePdfMenu}
          >
            {t("report.downloadPdf")}
            <svg
              class={"pdf-download__chevron" + (pdfOpen ? " is-open" : "")}
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              aria-hidden="true"
            >
              <path
                d="m6 9 6 6 6-6"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>

          {#if pdfOpen}
            <div
              bind:this={pdfMenu}
              class="lang-menu"
              id="pdf-menu"
              role="menu"
              tabindex="-1"
              aria-label={t("report.pdfLanguage")}
            >
              <button
                class="lang-menu__item"
                type="button"
                role="menuitemradio"
                aria-checked={getLocale() === "ru"}
                onclick={() => downloadPdf(current, "ru")}
              >
                <span lang="ru">Русский</span>
              </button>
              <button
                class="lang-menu__item"
                type="button"
                role="menuitemradio"
                aria-checked={getLocale() === "en"}
                onclick={() => downloadPdf(current, "en")}
              >
                <span lang="en">English</span>
              </button>
            </div>
          {/if}
        </div>
      </div>

      {#if pdfError}
        <p class="field__error" role="alert">{pdfError}</p>
      {/if}
    </section>
  </div>
{:else}
  <div class="container">
    <section class="section" aria-labelledby="scanning-title">
      <p class="hero__eyebrow">{t("report.scanEyebrow")}</p>
      <h1 class="section__title" id="scanning-title">
        {t("report.scanningTitle")}
      </h1>
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
      <p class="status" aria-live="polite">{t("report.scanningStatus")}</p>
    </section>
  </div>
{/if}

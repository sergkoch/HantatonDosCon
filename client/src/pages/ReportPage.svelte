<script lang="ts">
  import { getScan } from "../api";
  import Badge from "../components/Badge.svelte";
  import Button from "../components/Button.svelte";
  import { kindLabel } from "../lib/score";
  import type { Scan, Severity } from "../types";

  let { id }: { id: string } = $props();

  let scan = $state<Scan | null>(null);
  let error = $state("");
  let filter = $state<"all" | Severity>("all");

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
      <p class="hero__eyebrow">Отчёт</p>
      <h1 class="section__title" id="report-title">{scan.url}</h1>
      <p class="status">
        Оценка: {scan.score ?? "—"} · находок: {scan.findings.length}
      </p>

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
        <Button type="button" onclick={() => downloadCsv(current)}>
          Скачать CSV
        </Button>
      </div>
    </section>
  </div>
{:else}
  <div class="container">
    <p class="status" aria-live="polite">Загружаем отчёт…</p>
  </div>
{/if}

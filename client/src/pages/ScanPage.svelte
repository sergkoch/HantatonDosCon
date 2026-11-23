<script lang="ts">
  import { getScan } from "../api";
  import { navigate } from "../lib/router.svelte";
  import { scoreLabel } from "../lib/score";
  import type { Scan } from "../types";

  let { id }: { id: string } = $props();

  let scan = $state<Scan | null>(null);
  let error = $state("");

  $effect(() => {
    scan = null;
    error = "";

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
          cause instanceof Error ? cause.message : "Не удалось получить скан";
      }
    }

    load();

    return () => {
      active = false;
      if (timer) clearTimeout(timer);
    };
  });

  const done = $derived(scan?.status === "done");
  const failed = $derived(scan?.status === "failed");
  const statusText = $derived(
    done
      ? "Проверка завершена"
      : failed
        ? "Проверка не удалась"
        : "Проверяем страницы и медиа…",
  );
</script>

{#if error}
  <div class="container">
    <p class="field__error" role="alert">{error}</p>
  </div>
{:else if scan}
  {@const current = scan}
  <div class="container">
    <section class="section" aria-labelledby="scan-title">
      <p class="hero__eyebrow">Проверка</p>
      <h1 class="section__title" id="scan-title">{scan.url}</h1>

      <p class="status" aria-live="polite">{statusText}</p>

      <div class="summary">
        <div class="summary__item">
          <span class="summary__value">{scan.score ?? "—"}</span>
          <span class="summary__label">Оценка</span>
        </div>
        <div class="summary__item">
          <span class="summary__value">{scan.stats.critical}</span>
          <span class="summary__label">Критично</span>
        </div>
        <div class="summary__item">
          <span class="summary__value">{scan.stats.warning}</span>
          <span class="summary__label">Предупреждения</span>
        </div>
        <div class="summary__item">
          <span class="summary__value">{scan.stats.ok}</span>
          <span class="summary__label">В порядке</span>
        </div>
      </div>

      <p class="status">{scoreLabel(scan.score)}</p>

      {#if done}
        <a
          class="button button--primary"
          href={`/report/${current.id}`}
          onclick={(event) => {
            event.preventDefault();
            navigate(`/report/${current.id}`);
          }}
        >
          Открыть отчёт
        </a>
      {/if}
    </section>
  </div>
{:else}
  <div class="container">
    <p class="status" aria-live="polite">Загружаем скан…</p>
  </div>
{/if}

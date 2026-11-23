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
{:else}
  <div class="container">
    <section class="section" aria-labelledby="scan-title">
      <p class="hero__eyebrow">Проверка</p>
      <h1 class="section__title" id="scan-title">
        {scan?.url ?? "Проверяем сайт…"}
      </h1>

      {#if !done}
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
      {/if}

      <p class="status" aria-live="polite">{statusText}</p>

      {#if scan}
        {@const current = scan}
        <div class="summary">
          <div class="summary__item summary__item--score">
            <span class="summary__value">{current.score ?? "—"}</span>
            <span class="summary__label">Оценка</span>
          </div>
          <div class="summary__item summary__item--critical">
            <span class="summary__value">{current.stats.critical}</span>
            <span class="summary__label">Критично</span>
          </div>
          <div class="summary__item summary__item--warning">
            <span class="summary__value">{current.stats.warning}</span>
            <span class="summary__label">Предупреждения</span>
          </div>
          <div class="summary__item summary__item--ok">
            <span class="summary__value">{current.stats.ok}</span>
            <span class="summary__label">В порядке</span>
          </div>
        </div>

        <p class="status">{scoreLabel(current.score)}</p>

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
      {/if}
    </section>
  </div>
{/if}

<script lang="ts">
  import { onMount } from "svelte";
  import { clearHistory, createScan, listScans } from "../api";
  import { isPlainClick, navigate } from "../lib/router.svelte";
  import { getLocale, t } from "../lib/i18n.svelte";
  import { summarizeScans } from "../lib/insights";
  import type { Scan } from "../types";

  let scans = $state<Scan[]>([]);
  let query = $state("");
  let error = $state("");
  let loading = $state(true);
  let clearOpen = $state(false);
  let clearing = $state(false);
  let cleared = $state(0);

  async function clearAll() {
    clearing = true;
    error = "";

    try {
      cleared = await clearHistory();
      clearOpen = false;
      await load();
    } catch (cause) {
      error = cause instanceof Error ? cause.message : t("history.clearError");
      clearOpen = false;
    } finally {
      clearing = false;
    }
  }

  let repeatId = $state<string | null>(null);
  let repeatError = $state("");

  const summary = $derived(summarizeScans(scans));

  async function repeatScan(scan: Scan) {
    if (repeatId !== null) return;

    repeatId = scan.id;
    repeatError = "";

    try {
      const next = await createScan(scan.url, scan.depth);
      navigate(`/report/${next.id}`);
    } catch (cause) {
      repeatError =
        cause instanceof Error ? cause.message : t("history.repeatError");
    } finally {
      repeatId = null;
    }
  }

  function summaryDate(raw: string | null): string {
    return raw === null ? "—" : formatDate(raw);
  }

  async function load(url = "") {
    loading = true;
    error = "";

    try {
      scans = await listScans(url, 50);
    } catch (cause) {
      error = cause instanceof Error ? cause.message : t("history.error");
    } finally {
      loading = false;
    }
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    void load(query);
  }

  function openReport(id: string) {
    navigate(`/report/${id}`);
  }

  function formatDate(raw: string): string {
    return new Date(raw).toLocaleString(
      getLocale() === "ru" ? "ru-RU" : "en-GB",
      {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      },
    );
  }

  function statusLabel(status: Scan["status"]): string {
    switch (status) {
      case "queued":
        return t("history.queued");
      case "running":
        return t("history.running");
      case "done":
        return t("history.done");
      case "failed":
        return t("history.failed");
    }
  }

  onMount(() => {
    void load();
  });
</script>

<div class="container">
  <section class="section" aria-labelledby="history-title">
    <p class="hero__eyebrow">{t("history.eyebrow")}</p>
    <h1 class="section__title" id="history-title">{t("history.title")}</h1>
    <p class="status">
      {t("history.note")}
    </p>

    <form
      class="form form--filter"
      onsubmit={submit}
      aria-label={t("history.filterLabel")}
      {...{
        toolname: "filter_scan_history",
        tooldescription: "Фильтрует журнал проверок по точному адресу сайта",
      }}
    >
      <div class="field">
        <label class="field__label" for="history-url"
          >{t("history.urlLabel")}</label
        >
        <input
          class="input"
          id="history-url"
          name="url"
          type="url"
          bind:value={query}
          placeholder="https://example.ru"
          inputmode="url"
          {...{
            toolparamdescription: "Адрес сайта, историю которого нужно показать",
          }}
        />
      </div>
      <div class="actions">
        <button class="button button--primary" type="submit"
          >{t("history.show")}</button
        >
        <button
          class="button"
          type="button"
          onclick={() => {
            query = "";
            void load();
          }}
        >
          {t("history.reset")}
        </button>
      </div>
    </form>

    {#if error}
      <p class="field__error" role="alert">{error}</p>
    {:else if loading}
      <p class="status" aria-live="polite">{t("history.loading")}</p>
    {:else if scans.length === 0}
      <p class="empty">{t("history.empty")}</p>
    {:else}
      <section class="history-summary" aria-labelledby="history-summary-title">
        <h2 class="history-summary__title" id="history-summary-title">
          {t("history.summaryTitle")}
        </h2>
        <dl class="history-summary__grid">
          <div class="history-summary__item">
            <dt>{t("history.summaryTotal")}</dt>
            <dd>{summary.total}</dd>
          </div>
          <div class="history-summary__item">
            <dt>{t("history.summaryAverage")}</dt>
            <dd>{summary.average ?? "—"}</dd>
          </div>
          <div class="history-summary__item">
            <dt>{t("history.summaryCritical")}</dt>
            <dd>{summary.critical}</dd>
          </div>
          <div class="history-summary__item">
            <dt>{t("history.summaryLast")}</dt>
            <dd>{summaryDate(summary.lastAt)}</dd>
          </div>
        </dl>
      </section>

      <div class="table-wrap">
        <table class="table">
          <caption class="visually-hidden">{t("history.caption")}</caption>
          <thead>
            <tr>
              <th scope="col">{t("history.version")}</th>
              <th scope="col">{t("history.address")}</th>
              <th scope="col">{t("history.date")}</th>
              <th scope="col">{t("history.status")}</th>
              <th scope="col">{t("history.findings")}</th>
              <th scope="col">{t("history.score")}</th>
              <th scope="col">{t("history.actions")}</th>
            </tr>
          </thead>
          <tbody>
            {#each scans as scan (scan.id)}
              <tr>
                <td>v{scan.version}</td>
                <td>{scan.url}</td>
                <td>{formatDate(scan.createdAt)}</td>
                <td>{statusLabel(scan.status)}</td>
                <td>
                  {scan.stats.critical + scan.stats.warning}
                  <span class="muted">
                    ({t("history.criticalCount", {
                      count: scan.stats.critical,
                    })})
                  </span>
                </td>
                <td>{scan.score ?? "—"}</td>
                <td class="row-actions">
                  <a
                    href={`/report/${scan.id}`}
                    onclick={(event) => {
                      if (!isPlainClick(event)) return;
                      event.preventDefault();
                      openReport(scan.id);
                    }}
                  >
                    {t("history.open")}
                  </a>
                  <button
                    class="link-button"
                    type="button"
                    disabled={repeatId !== null}
                    onclick={() => repeatScan(scan)}
                  >
                    {repeatId === scan.id
                      ? t("history.repeatBusy")
                      : t("history.repeat")}
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}

    {#if repeatError}
      <p class="field__error" role="alert">{repeatError}</p>
    {/if}

    {#if cleared > 0}
      <p class="status" aria-live="polite">
        {t("history.cleared", { count: cleared })}
      </p>
    {/if}

    <div class="actions actions--between">
      <a
        href="/"
        onclick={(event) => {
          if (!isPlainClick(event)) return;
          event.preventDefault();
          navigate("/");
        }}
      >
        {t("history.newCheck")}
      </a>

      {#if scans.length > 0}
        {#if clearOpen}
          <div
            class="history-clear"
            role="group"
            aria-label={t("history.clearConfirm")}
          >
            <span class="history-clear__text">{t("history.clearConfirm")}</span>
            <button
              class="button button--danger"
              type="button"
              disabled={clearing}
              onclick={clearAll}
            >
              {clearing ? t("history.clearing") : t("history.clearYes")}
            </button>
            <button
              class="button"
              type="button"
              disabled={clearing}
              onclick={() => (clearOpen = false)}
            >
              {t("history.clearNo")}
            </button>
          </div>
        {:else}
          <button
            class="button button--danger"
            type="button"
            onclick={() => (clearOpen = true)}
          >
            {t("history.clear")}
          </button>
        {/if}
      {/if}
    </div>
  </section>
</div>

<script lang="ts">
  import { onMount } from "svelte";
  import { listScans } from "../api";
  import { navigate } from "../lib/router.svelte";
  import type { Scan } from "../types";

  let scans = $state<Scan[]>([]);
  let query = $state("");
  let error = $state("");
  let loading = $state(true);

  async function load(url = "") {
    loading = true;
    error = "";

    try {
      scans = await listScans(url, 50);
    } catch (cause) {
      error =
        cause instanceof Error ? cause.message : "Не удалось получить историю";
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
    return new Date(raw).toLocaleString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function statusLabel(status: Scan["status"]): string {
    switch (status) {
      case "queued":
        return "В очереди";
      case "running":
        return "Идёт проверка";
      case "done":
        return "Готово";
      case "failed":
        return "Ошибка";
    }
  }

  onMount(() => {
    void load();
  });
</script>

<div class="container">
  <section class="section" aria-labelledby="history-title">
    <p class="hero__eyebrow">Журнал проверок</p>
    <h1 class="section__title" id="history-title">История проверок</h1>
    <p class="status">
      Каждая повторная проверка одного адреса получает новый номер версии.
      История хранится только в этом браузере и не смешивается с проверками
      других посетителей.
    </p>

    <form
      class="form form--filter"
      onsubmit={submit}
      aria-label="Фильтр истории по адресу"
    >
      <div class="field">
        <label class="field__label" for="history-url">Адрес сайта</label>
        <input
          class="input"
          id="history-url"
          name="url"
          type="url"
          bind:value={query}
          placeholder="https://example.ru"
          inputmode="url"
        />
      </div>
      <div class="actions">
        <button class="button button--primary" type="submit">Показать</button>
        <button
          class="button"
          type="button"
          onclick={() => {
            query = "";
            void load();
          }}
        >
          Сбросить
        </button>
      </div>
    </form>

    {#if error}
      <p class="field__error" role="alert">{error}</p>
    {:else if loading}
      <p class="status" aria-live="polite">Загружаем историю…</p>
    {:else if scans.length === 0}
      <p class="empty">Проверок пока нет. Запустите первую проверку на главной.</p>
    {:else}
      <div class="table-wrap">
        <table class="table">
          <caption class="visually-hidden">Журнал проверок</caption>
          <thead>
            <tr>
              <th scope="col">Версия</th>
              <th scope="col">Адрес</th>
              <th scope="col">Дата</th>
              <th scope="col">Статус</th>
              <th scope="col">Находки</th>
              <th scope="col">Оценка</th>
              <th scope="col">Отчёт</th>
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
                    ({scan.stats.critical} критично)
                  </span>
                </td>
                <td>{scan.score ?? "—"}</td>
                <td>
                  <a
                    href={`/report/${scan.id}`}
                    onclick={(event) => {
                      event.preventDefault();
                      openReport(scan.id);
                    }}
                  >
                    Открыть
                  </a>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}

    <div class="actions">
      <a
        href="/"
        onclick={(event) => {
          event.preventDefault();
          navigate("/");
        }}
      >
        Новая проверка
      </a>
    </div>
  </section>
</div>

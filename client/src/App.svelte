<script lang="ts">
  import Layout from "./components/Layout.svelte";
  import HomePage from "./pages/HomePage.svelte";
  import NotFoundPage from "./pages/NotFoundPage.svelte";
  import { getPath } from "./lib/router.svelte";

  const path = $derived(getPath());
  const reportId = $derived(path.match(/^\/report\/([^/]+)$/)?.[1] ?? null);

  const loadHistory = () =>
    import("./pages/HistoryPage.svelte").then((module) => module.default);
  const loadReport = () =>
    import("./pages/ReportPage.svelte").then((module) => module.default);
</script>

<Layout>
  {#if path === "/"}
    <HomePage />
  {:else if path === "/history"}
    {#await loadHistory() then HistoryPage}
      <HistoryPage />
    {:catch}
      <div class="container">
        <p class="field__error" role="alert">Не удалось загрузить историю.</p>
      </div>
    {/await}
  {:else if reportId !== null}
    {#await loadReport() then ReportPage}
      <ReportPage id={reportId} />
    {:catch}
      <div class="container">
        <p class="field__error" role="alert">Не удалось загрузить отчёт.</p>
      </div>
    {/await}
  {:else}
    <NotFoundPage />
  {/if}
</Layout>

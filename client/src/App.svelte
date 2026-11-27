<script lang="ts">
  import Layout from "./components/Layout.svelte";
  import RouteSkeleton from "./components/RouteSkeleton.svelte";
  import HomePage from "./pages/HomePage.svelte";
  import NotFoundPage from "./pages/NotFoundPage.svelte";
  import { getPath } from "./lib/router.svelte";
  import { loadHistory, loadReport } from "./lib/routes";

  const path = $derived(getPath());
  const reportId = $derived(path.match(/^\/report\/([^/]+)$/)?.[1] ?? null);
</script>

<Layout>
  {#if path === "/"}
    <HomePage />
  {:else if path === "/history"}
    {#await loadHistory()}
      <RouteSkeleton />
    {:then module}
      {@const HistoryPage = module.default}
      <HistoryPage />
    {:catch}
      <div class="container">
        <p class="field__error" role="alert">Не удалось загрузить историю.</p>
      </div>
    {/await}
  {:else if reportId !== null}
    {#await loadReport()}
      <RouteSkeleton />
    {:then module}
      {@const ReportPage = module.default}
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

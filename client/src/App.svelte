<script lang="ts">
  import Layout from "./components/Layout.svelte";
  import RouteSkeleton from "./components/RouteSkeleton.svelte";
  import HomePage from "./pages/HomePage.svelte";
  import NotFoundPage from "./pages/NotFoundPage.svelte";
  import { t } from "./lib/i18n.svelte";
  import { getPath } from "./lib/router.svelte";
  import { loadHistory, loadReport, matchRoute } from "./lib/routes";

  const path = $derived(getPath());
  const route = $derived(matchRoute(path));
</script>

<Layout>
  {#if route.kind === "home"}
    <HomePage />
  {:else if route.kind === "history"}
    {#await loadHistory()}
      <RouteSkeleton />
    {:then HistoryPage}
      <HistoryPage />
    {:catch}
      <div class="container">
        <p class="field__error" role="alert">{t("history.error")}</p>
      </div>
    {/await}
  {:else if route.kind === "report" && route.reportId !== null}
    {#await loadReport()}
      <RouteSkeleton />
    {:then ReportPage}
      <ReportPage id={route.reportId} />
    {:catch}
      <div class="container">
        <p class="field__error" role="alert">{t("report.error")}</p>
      </div>
    {/await}
  {:else}
    <NotFoundPage />
  {/if}
</Layout>

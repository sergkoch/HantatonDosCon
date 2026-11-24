<script lang="ts">
  import Layout from "./components/Layout.svelte";
  import HistoryPage from "./pages/HistoryPage.svelte";
  import HomePage from "./pages/HomePage.svelte";
  import NotFoundPage from "./pages/NotFoundPage.svelte";
  import ReportPage from "./pages/ReportPage.svelte";
  import ScanPage from "./pages/ScanPage.svelte";
  import { getPath } from "./lib/router.svelte";

  const path = $derived(getPath());
  const scanId = $derived(path.match(/^\/scan\/([^/]+)$/)?.[1] ?? null);
  const reportId = $derived(path.match(/^\/report\/([^/]+)$/)?.[1] ?? null);
</script>

<Layout>
  {#if path === "/"}
    <HomePage />
  {:else if path === "/history"}
    <HistoryPage />
  {:else if scanId !== null}
    <ScanPage id={scanId} />
  {:else if reportId !== null}
    <ReportPage id={reportId} />
  {:else}
    <NotFoundPage />
  {/if}
</Layout>

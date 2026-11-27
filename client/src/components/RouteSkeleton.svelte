<script lang="ts">
  import { t } from "../lib/i18n.svelte";

  let { label }: { label?: string } = $props();
  const text = $derived(label ?? t("route.loading"));
</script>

<div
  class="container route-skeleton"
  role="status"
  aria-live="polite"
  aria-busy="true"
>
  <span class="visually-hidden">{text}</span>
  <div class="route-skeleton__head" aria-hidden="true"></div>
  <div class="route-skeleton__line route-skeleton__line--short" aria-hidden="true"></div>
  {#each [0, 1, 2, 3, 4] as row (row)}
    <div class="route-skeleton__row" aria-hidden="true"></div>
  {/each}
</div>

<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    title,
    icon,
    children,
  }: { title: string; icon?: Snippet; children: Snippet } = $props();
</script>

<article class="card">
  {#if icon}
    <span class="card__icon" aria-hidden="true">{@render icon()}</span>
  {/if}
  <h3 class="card__title">{title}</h3>
  <div class="card__body">{@render children()}</div>
</article>

<script lang="ts">
  import type { Snippet } from "svelte";

  let { title, children }: { title: string; children: Snippet } = $props();
</script>

<article class="card">
  <h3 class="card__title">{title}</h3>
  <div class="card__body">{@render children()}</div>
</article>

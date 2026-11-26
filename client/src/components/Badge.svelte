<script lang="ts">
  import type { Severity } from "../types";
  import { t } from "../lib/i18n.svelte";

  let { severity }: { severity: Severity } = $props();

  const keys = {
    critical: "severity.critical",
    warning: "severity.warning",
    ok: "severity.ok",
  } as const;
</script>

<span class={`badge badge--${severity}`}>{t(keys[severity])}</span>

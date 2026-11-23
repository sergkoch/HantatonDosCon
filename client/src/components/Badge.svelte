<script lang="ts">
  import type { Severity } from "../types";

  let { severity }: { severity: Severity } = $props();

  const labels: Record<Severity, string> = {
    critical: "Критично",
    warning: "Предупреждение",
    ok: "В порядке",
  };
</script>

<span class={`badge badge--${severity}`}>{labels[severity]}</span>

import type { Severity } from "../types";

const labels: Record<Severity, string> = {
  critical: "Критично",
  warning: "Предупреждение",
  ok: "В порядке",
};

export function Badge({ severity }: { severity: Severity }) {
  return <span className={`badge badge--${severity}`}>{labels[severity]}</span>;
}

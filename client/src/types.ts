export type Severity = "critical" | "warning" | "ok";

export type FindingKind = "image" | "pdf" | "link";

export type Finding = {
  id: string;
  kind: FindingKind;
  severity: Severity;
  title: string;
  element: string;
  recommendation: string;
  wcag: string;
};

export type ScanStats = {
  critical: number;
  warning: number;
  ok: number;
};

export type ScanStatus = "queued" | "running" | "done" | "failed";

export type Scan = {
  id: string;
  url: string;
  depth: number;
  status: ScanStatus;
  score: number | null;
  stats: ScanStats;
  findings: Finding[];
  createdAt: string;
};

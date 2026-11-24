export type Severity = "critical" | "warning" | "ok";

export type FindingKind =
  | "image"
  | "pdf"
  | "link"
  | "page"
  | "form"
  | "table";

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
  pages: number;
  checkedLinks?: number;
  unverifiedLinks?: number;
  error?: string;
  stats: ScanStats;
  findings: Finding[];
  createdAt: string;
};

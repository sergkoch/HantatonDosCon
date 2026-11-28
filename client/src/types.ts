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
  code?: string;
  kind: FindingKind;
  severity: Severity;
  title: string;
  element: string;
  recommendation: string;
  wcag: string;
  /** Уверенность ИИ для смысловых находок, 0..1. */
  confidence?: number;
  /** Те же тексты на втором языке, если скан сохранён с переводом. */
  titleAlt?: string;
  recommendationAlt?: string;
};

export type ScanStats = {
  critical: number;
  warning: number;
  ok: number;
};

export type ScanStatus = "queued" | "running" | "done" | "failed";

export type Changes = {
  added: Finding[];
  fixed: Finding[];
  remaining: number;
  scoreDelta?: number;
};

export type AICoverage = {
  enabled: boolean;
  imagesChecked: number;
  linksChecked: number;
  skipped: number;
};

export type Scan = {
  id: string;
  url: string;
  depth: number;
  status: ScanStatus;
  version: number;
  previousId?: string;
  score: number | null;
  pages: number;
  checkedLinks?: number;
  unverifiedLinks?: number;
  error?: string;
  stats: ScanStats;
  ai?: AICoverage;
  changes?: Changes;
  findings: Finding[];
  createdAt: string;
  /** Язык, на котором запускалась проверка. */
  lang?: "ru" | "en";
};

import type { Finding, Scan } from "../types";

export type InsightRow = {
  key: string;
  count: number;
  /** Ширина полосы в процентах от самой частой группы. */
  percent: number;
};

function breakdown(
  findings: Finding[],
  keyOf: (finding: Finding) => string,
): InsightRow[] {
  const counts = new Map<string, number>();

  for (const finding of findings) {
    if (finding.severity === "ok") {
      continue;
    }

    const key = keyOf(finding).trim() || "unknown";
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }

  const rows = [...counts.entries()]
    .map(([key, count]) => ({ key, count }))
    .sort((a, b) => b.count - a.count || a.key.localeCompare(b.key));

  const max = rows[0]?.count ?? 0;

  return rows.map((row) => ({
    ...row,
    percent: max === 0 ? 0 : Math.round((row.count / max) * 100),
  }));
}

export function kindBreakdown(findings: Finding[]): InsightRow[] {
  return breakdown(findings, (finding) => finding.kind);
}

export function wcagBreakdown(findings: Finding[]): InsightRow[] {
  return breakdown(findings, (finding) => finding.wcag || "—");
}

export function problemsCount(findings: Finding[]): number {
  return findings.filter((finding) => finding.severity !== "ok").length;
}

export type HistorySummary = {
  total: number;
  /** Средняя только по сканам с числовой оценкой. */
  average: number | null;
  critical: number;
  /** ISO-дата самой свежей проверки. */
  lastAt: string | null;
};

export function summarizeScans(scans: Scan[]): HistorySummary {
  const scores = scans
    .map((scan) => scan.score)
    .filter((score): score is number => typeof score === "number");

  const average =
    scores.length === 0
      ? null
      : Math.round(scores.reduce((sum, score) => sum + score, 0) / scores.length);

  const critical = scans.reduce(
    (sum, scan) => sum + (scan.stats?.critical ?? 0),
    0,
  );

  let lastAt: string | null = null;
  for (const scan of scans) {
    if (lastAt === null || scan.createdAt > lastAt) {
      lastAt = scan.createdAt;
    }
  }

  return { total: scans.length, average, critical, lastAt };
}

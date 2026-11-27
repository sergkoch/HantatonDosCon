import type {
  Finding,
  FindingKind,
  Scan,
} from "../types";

export type InsightRow<K extends string = string> = {
  key: K;
  count: number;
  /** Ширина полосы в процентах от самой частой группы. */
  percent: number;
};

export type HistorySummary = {
  total: number;
  averageScore: number | null;
  critical: number;
  lastDate: string | null;
};

/** Проблемы — всё, кроме находок «в порядке». */
export function problemFindings(findings: Finding[]): Finding[] {
  return findings.filter((finding) => finding.severity !== "ok");
}

function toRows<K extends string>(counts: Map<K, number>): InsightRow<K>[] {
  const max = Math.max(0, ...counts.values());

  return Array.from(counts, ([key, count]) => ({
    key,
    count,
    percent: max === 0 ? 0 : Math.round((count / max) * 100),
  })).sort(
    (a, b) =>
      b.count - a.count || a.key.localeCompare(b.key, "ru", { numeric: true }),
  );
}

/** Проблемы по типам контента, от частых к редким. */
export function kindBreakdown(findings: Finding[]): InsightRow<FindingKind>[] {
  const counts = new Map<FindingKind, number>();

  for (const finding of problemFindings(findings)) {
    counts.set(finding.kind, (counts.get(finding.kind) ?? 0) + 1);
  }

  return toRows(counts);
}

/** Проблемы по критериям WCAG, от частых к редким. */
export function wcagBreakdown(findings: Finding[]): InsightRow[] {
  const counts = new Map<string, number>();

  for (const finding of problemFindings(findings)) {
    const key = finding.wcag.trim() === "" ? "—" : finding.wcag;
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }

  return toRows(counts);
}

/**
 * Сводка по журналу проверок. Средняя оценка считается только по сканам
 * с числовым score: queued, running и failed её не ломают.
 */
export function historySummary(scans: Scan[]): HistorySummary {
  const scores = scans
    .map((scan) => scan.score)
    .filter((score): score is number => typeof score === "number");

  const averageScore =
    scores.length === 0
      ? null
      : Math.round(scores.reduce((sum, score) => sum + score, 0) / scores.length);

  const critical = scans.reduce(
    (sum, scan) => sum + (scan.stats?.critical ?? 0),
    0,
  );

  const lastDate = scans.reduce<string | null>((latest, scan) => {
    if (latest === null) return scan.createdAt;
    return scan.createdAt > latest ? scan.createdAt : latest;
  }, null);

  return {
    total: scans.length,
    averageScore,
    critical,
    lastDate,
  };
}

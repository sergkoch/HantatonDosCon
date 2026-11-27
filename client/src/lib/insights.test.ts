import { describe, expect, it } from "vitest";
import {
  historySummary,
  kindBreakdown,
  problemFindings,
  wcagBreakdown,
} from "./insights";
import type { Finding, Scan } from "../types";

function finding(
  overrides: Partial<Finding> & Pick<Finding, "kind" | "severity" | "wcag">,
): Finding {
  return {
    id: "f1",
    code: "CODE",
    title: "Проблема",
    element: "<img>",
    recommendation: "Исправьте",
    ...overrides,
  };
}

function scan(overrides: Partial<Scan> = {}): Scan {
  return {
    id: "scan-1",
    url: "https://example.ru",
    depth: 1,
    status: "done",
    version: 1,
    score: 100,
    pages: 1,
    stats: { critical: 0, warning: 0, ok: 0 },
    findings: [],
    createdAt: "2026-11-27T10:00:00Z",
    ...overrides,
  };
}

describe("problemFindings", () => {
  it("ignores ok findings", () => {
    const findings = [
      finding({ id: "a", kind: "image", severity: "critical", wcag: "1.1.1" }),
      finding({ id: "b", kind: "image", severity: "ok", wcag: "1.1.1" }),
      finding({ id: "c", kind: "link", severity: "warning", wcag: "2.4.4" }),
    ];

    expect(problemFindings(findings).map((item) => item.id)).toEqual([
      "a",
      "c",
    ]);
  });
});

describe("kindBreakdown", () => {
  it("returns an empty list for an empty scan", () => {
    expect(kindBreakdown([])).toEqual([]);
  });

  it("sorts by frequency and scales bars from the maximum", () => {
    const findings = [
      finding({ id: "1", kind: "link", severity: "critical", wcag: "2.4.4" }),
      finding({ id: "2", kind: "image", severity: "critical", wcag: "1.1.1" }),
      finding({ id: "3", kind: "image", severity: "warning", wcag: "1.1.1" }),
      finding({ id: "4", kind: "pdf", severity: "ok", wcag: "1.3.1" }),
    ];

    expect(kindBreakdown(findings)).toEqual([
      { key: "image", count: 2, percent: 100 },
      { key: "link", count: 1, percent: 50 },
    ]);
  });

  it("counts every problem exactly once", () => {
    const findings = [
      finding({ id: "1", kind: "image", severity: "critical", wcag: "1.1.1" }),
      finding({ id: "2", kind: "pdf", severity: "warning", wcag: "1.3.1" }),
      finding({ id: "3", kind: "table", severity: "warning", wcag: "1.3.1" }),
      finding({ id: "4", kind: "form", severity: "ok", wcag: "3.3.2" }),
    ];

    const rows = kindBreakdown(findings);
    const total = rows.reduce((sum, row) => sum + row.count, 0);

    expect(total).toBe(problemFindings(findings).length);
  });
});

describe("wcagBreakdown", () => {
  it("groups by criterion and puts frequent ones first", () => {
    const findings = [
      finding({ id: "1", kind: "link", severity: "critical", wcag: "2.4.4" }),
      finding({ id: "2", kind: "link", severity: "warning", wcag: "2.4.4" }),
      finding({ id: "3", kind: "image", severity: "critical", wcag: "1.1.1" }),
      finding({ id: "4", kind: "image", severity: "ok", wcag: "1.1.1" }),
    ];

    expect(wcagBreakdown(findings)).toEqual([
      { key: "2.4.4", count: 2, percent: 100 },
      { key: "1.1.1", count: 1, percent: 50 },
    ]);
  });
});

describe("historySummary", () => {
  it("returns an empty summary for an empty history", () => {
    expect(historySummary([])).toEqual({
      total: 0,
      averageScore: null,
      critical: 0,
      lastDate: null,
    });
  });

  it("averages only numeric scores", () => {
    const summary = historySummary([
      scan({
        id: "1",
        score: 100,
        stats: { critical: 1, warning: 2, ok: 3 },
        createdAt: "2026-11-25T10:00:00Z",
      }),
      scan({
        id: "2",
        score: 50,
        stats: { critical: 2, warning: 0, ok: 1 },
        createdAt: "2026-11-27T10:00:00Z",
      }),
      scan({
        id: "3",
        status: "running",
        score: null,
        stats: { critical: 0, warning: 0, ok: 0 },
        createdAt: "2026-11-26T10:00:00Z",
      }),
    ]);

    expect(summary.total).toBe(3);
    expect(summary.averageScore).toBe(75);
    expect(summary.critical).toBe(3);
    expect(summary.lastDate).toBe("2026-11-27T10:00:00Z");
  });

  it("does not break on queued scans without a score", () => {
    const summary = historySummary([
      scan({ id: "1", status: "queued", score: null, createdAt: "2026-11-24T10:00:00Z" }),
      scan({ id: "2", status: "done", score: 80, createdAt: "2026-11-23T10:00:00Z" }),
    ]);

    expect(summary.averageScore).toBe(80);
    expect(summary.lastDate).toBe("2026-11-24T10:00:00Z");
  });
});

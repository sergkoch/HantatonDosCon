import { describe, expect, it } from "vitest";
import type { Finding, Scan } from "../types";
import {
  kindBreakdown,
  problemsCount,
  summarizeScans,
  wcagBreakdown,
} from "./insights";

function finding(partial: Partial<Finding>): Finding {
  return {
    id: "f1",
    kind: "image",
    severity: "critical",
    title: "t",
    element: "e",
    recommendation: "r",
    wcag: "1.1.1",
    ...partial,
  };
}

function scan(partial: Partial<Scan>): Scan {
  return {
    id: "s1",
    url: "https://example.ru",
    depth: 1,
    status: "done",
    version: 1,
    score: 100,
    pages: 1,
    stats: { critical: 0, warning: 0, ok: 0 },
    findings: [],
    createdAt: "2026-11-27T10:00:00Z",
    ...partial,
  };
}

describe("kindBreakdown", () => {
  it("returns nothing for an empty list", () => {
    expect(kindBreakdown([])).toEqual([]);
  });

  it("ignores findings in order", () => {
    const rows = kindBreakdown([
      finding({ id: "a", kind: "link", severity: "warning" }),
      finding({ id: "b", kind: "image" }),
      finding({ id: "c", kind: "image" }),
      finding({ id: "d", kind: "image", severity: "ok" }),
    ]);

    expect(rows).toEqual([
      { key: "image", count: 2, percent: 100 },
      { key: "link", count: 1, percent: 50 },
    ]);
  });

  it("counts problems across all types", () => {
    const findings = [
      finding({ id: "a" }),
      finding({ id: "b", kind: "pdf" }),
      finding({ id: "c", severity: "ok" }),
    ];

    expect(problemsCount(findings)).toBe(2);
    const sum = kindBreakdown(findings).reduce((acc, row) => acc + row.count, 0);
    expect(sum).toBe(2);
  });
});

describe("wcagBreakdown", () => {
  it("groups by WCAG criterion and sorts by frequency", () => {
    const rows = wcagBreakdown([
      finding({ id: "a", wcag: "1.1.1" }),
      finding({ id: "b", wcag: "2.4.4" }),
      finding({ id: "c", wcag: "1.1.1" }),
    ]);

    expect(rows.map((row) => row.key)).toEqual(["1.1.1", "2.4.4"]);
    expect(rows[0].percent).toBe(100);
    expect(rows[1].percent).toBe(50);
  });
});

describe("summarizeScans", () => {
  it("handles empty history", () => {
    expect(summarizeScans([])).toEqual({
      total: 0,
      average: null,
      critical: 0,
      lastAt: null,
    });
  });

  it("averages only numeric scores", () => {
    const summary = summarizeScans([
      scan({ id: "a", score: 100 }),
      scan({ id: "b", score: 50 }),
      scan({ id: "c", score: null, status: "running" }),
    ]);

    expect(summary.total).toBe(3);
    expect(summary.average).toBe(75);
  });

  it("sums critical findings and finds the latest scan", () => {
    const summary = summarizeScans([
      scan({
        id: "a",
        createdAt: "2026-11-25T10:00:00Z",
        stats: { critical: 2, warning: 1, ok: 0 },
      }),
      scan({
        id: "b",
        createdAt: "2026-11-27T12:00:00Z",
        stats: { critical: 1, warning: 0, ok: 0 },
      }),
    ]);

    expect(summary.critical).toBe(3);
    expect(summary.lastAt).toBe("2026-11-27T12:00:00Z");
  });
});

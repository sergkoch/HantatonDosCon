import type { Component } from "svelte";

export type RouteKind = "home" | "history" | "report" | "notFound";

export type Route = {
  kind: RouteKind;
  reportId: string | null;
};

export function matchRoute(path: string): Route {
  if (path === "/") {
    return { kind: "home", reportId: null };
  }

  if (path === "/history") {
    return { kind: "history", reportId: null };
  }

  const report = path.match(/^\/report\/([^/]+)$/);
  if (report) {
    return { kind: "report", reportId: report[1] };
  }

  return { kind: "notFound", reportId: null };
}

type HistoryModule = typeof import("../pages/HistoryPage.svelte");
type ReportModule = typeof import("../pages/ReportPage.svelte");

let historyPromise: Promise<HistoryModule["default"]> | null = null;
let reportPromise: Promise<ReportModule["default"]> | null = null;

/**
 * Промис кешируется: повторный переход в историю не ждёт загрузку чанка
 * и не мигает скелетоном.
 */
export function loadHistory(): Promise<HistoryModule["default"]> {
  if (!historyPromise) {
    historyPromise = import("../pages/HistoryPage.svelte").then(
      (module) => module.default,
    );
  }

  return historyPromise;
}

export function loadReport(): Promise<ReportModule["default"]> {
  if (!reportPromise) {
    reportPromise = import("../pages/ReportPage.svelte").then(
      (module) => module.default,
    );
  }

  return reportPromise;
}

/** Заранее подгружает чанк истории по наведению или фокусу на ссылке. */
export function prefetchHistory(): void {
  void loadHistory();
}

export type LoadedRoute = {
  component: Component<Record<string, never>> | Component<{ id: string }>;
  kind: RouteKind;
};

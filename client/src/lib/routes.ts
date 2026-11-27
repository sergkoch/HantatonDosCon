/**
 * Ленивые страницы кешируются: повторный переход не запускает новый импорт,
 * а предзагрузка по наведению делает переход мгновенным.
 */

function importHistory() {
  return import("../pages/HistoryPage.svelte");
}

function importReport() {
  return import("../pages/ReportPage.svelte");
}

let historyPromise: ReturnType<typeof importHistory> | null = null;
let reportPromise: ReturnType<typeof importReport> | null = null;

export function loadHistory() {
  if (!historyPromise) {
    historyPromise = importHistory().catch((error: unknown) => {
      historyPromise = null;
      throw error;
    });
  }
  return historyPromise;
}

export function loadReport() {
  if (!reportPromise) {
    reportPromise = importReport().catch((error: unknown) => {
      reportPromise = null;
      throw error;
    });
  }
  return reportPromise;
}

/** Загружает чанк истории заранее. */
export function prefetchHistory(): void {
  void loadHistory().catch(() => {
    // Ошибку покажет сама страница при переходе.
  });
}

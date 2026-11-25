let path = $state(window.location.pathname);

function titleFor(next: string): string {
  if (next === "/") return "Доступный контент — проверка сайта";
  if (next === "/history") return "История проверок — Доступный контент";
  if (next.startsWith("/scan/")) return "Проверка идёт — Доступный контент";
  if (next.startsWith("/report/")) return "Отчёт — Доступный контент";
  return "Страница не найдена — Доступный контент";
}

function applyPath(next: string): void {
  path = next;
  document.title = titleFor(next);
  window.scrollTo({ top: 0, behavior: "auto" });
  requestAnimationFrame(() => {
    document.getElementById("main")?.focus();
  });
}

export function navigate(to: string): void {
  window.history.pushState(null, "", to);
  applyPath(to);
}

export function getPath(): string {
  return path;
}

export function isPlainClick(event: MouseEvent): boolean {
  return (
    event.button === 0 &&
    !event.metaKey &&
    !event.ctrlKey &&
    !event.shiftKey &&
    !event.altKey
  );
}

window.addEventListener("popstate", () => {
  applyPath(window.location.pathname);
});

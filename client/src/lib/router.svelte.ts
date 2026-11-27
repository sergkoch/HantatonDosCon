let path = $state(window.location.pathname);

function applyPath(next: string): void {
  const changed = path !== next;
  path = next;

  if (changed) {
    window.scrollTo({ top: 0, behavior: "auto" });
  }

  // preventScroll важен: иначе фокус на main прокручивает страницу второй раз.
  requestAnimationFrame(() => {
    document.getElementById("main")?.focus({ preventScroll: true });
  });
}

export function navigate(to: string): void {
  const url = new URL(to, window.location.origin);
  const target = url.pathname + url.search;
  const current = window.location.pathname + window.location.search;

  // Повторный клик по текущему пункту меню не должен сбрасывать прокрутку.
  if (target === current) {
    return;
  }

  window.history.pushState(null, "", target + url.hash);
  applyPath(url.pathname);
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

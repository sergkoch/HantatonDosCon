let path = $state(window.location.pathname);

function applyPath(next: string): void {
  path = next;
  window.scrollTo({ top: 0, behavior: "auto" });
  requestAnimationFrame(() => {
    document.getElementById("main")?.focus({ preventScroll: true });
  });
}

export function navigate(to: string): void {
  const current = window.location.pathname + window.location.search;
  if (current === to) {
    return;
  }

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

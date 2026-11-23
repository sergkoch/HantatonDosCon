let path = $state(window.location.pathname);

export function navigate(to: string): void {
  window.history.pushState(null, "", to);
  path = to;
}

export function getPath(): string {
  return path;
}

window.addEventListener("popstate", () => {
  path = window.location.pathname;
});

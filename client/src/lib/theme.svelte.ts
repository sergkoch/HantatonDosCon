import { tick } from "svelte";

export type Theme = "light" | "dark";

type ViewTransitionDoc = Document & {
  startViewTransition?: (
    callback: () => void | Promise<void>,
  ) => { ready: Promise<void> };
};

function readInitial(): Theme {
  if (typeof document !== "undefined") {
    const attr = document.documentElement.dataset.theme;
    if (attr === "light" || attr === "dark") {
      return attr;
    }
  }
  return "dark";
}

function prefersReducedMotion(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

let current = $state<Theme>(readInitial());

export function getTheme(): Theme {
  return current;
}

export function setTheme(next: Theme): void {
  current = next;
  document.documentElement.dataset.theme = next;
  try {
    localStorage.setItem("theme", next);
  } catch {
    // приватный режим/недоступное хранилище — просто игнорируем
  }
}

export function toggleTheme(event?: MouseEvent): void {
  const next = current === "dark" ? "light" : "dark";

  const startViewTransition = (
    document as ViewTransitionDoc
  ).startViewTransition?.bind(document);

  if (!startViewTransition || prefersReducedMotion()) {
    setTheme(next);
    return;
  }

  let applied = false;
  const apply = () => {
    if (applied) return;
    applied = true;
    setTheme(next);
  };

  let transition: { ready: Promise<void>; skipTransition?: () => void };

  try {
    transition = startViewTransition(async () => {
      apply();
      await tick();
    });
  } catch {
    apply();
    return;
  }

  // Если переход не запустился (фоновая вкладка, троттлинг), кнопка всё равно
  // должна работать: применяем тему принудительно и снимаем зависший переход.
  const fallback = window.setTimeout(() => {
    apply();
    try {
      transition.skipTransition?.();
    } catch {
      // переход уже завершён
    }
  }, 300);

  const x = event?.clientX ?? window.innerWidth - 48;
  const y = event?.clientY ?? 48;

  transition.ready
    .then(() => {
      window.clearTimeout(fallback);

      const radius = Math.hypot(
        Math.max(x, window.innerWidth - x),
        Math.max(y, window.innerHeight - y),
      );

      document.documentElement.animate(
        {
          clipPath: [
            `circle(0px at ${x}px ${y}px)`,
            `circle(${radius}px at ${x}px ${y}px)`,
          ],
        },
        {
          duration: 550,
          easing: "cubic-bezier(0.4, 0, 0.2, 1)",
          pseudoElement: "::view-transition-new(root)",
        },
      );
    })
    .catch(() => {
      window.clearTimeout(fallback);
      apply();
    });
}

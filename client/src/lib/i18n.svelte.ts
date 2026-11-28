import {
  type Locale,
  type MessageKey,
  translate,
} from "./messages";

const STORAGE_KEY = "locale";

function readInitial(): Locale {
  if (typeof window === "undefined") {
    return "ru";
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "ru" || stored === "en") {
      return stored;
    }
  } catch {
    // приватный режим — смотрим на язык браузера
  }

  return navigator.language?.toLowerCase().startsWith("en") ? "en" : "ru";
}

const initialLocale = readInitial();
let current = $state<Locale>(initialLocale);

function applyLocale(locale: Locale): void {
  if (typeof document !== "undefined") {
    document.documentElement.lang = locale;
  }
}

applyLocale(initialLocale);

export function getLocale(): Locale {
  return current;
}

export function t(
  key: MessageKey,
  params?: Record<string, string | number>,
): string {
  return translate(current, key, params);
}

export function setLocale(next: Locale): void {
  current = next;
  applyLocale(next);

  try {
    localStorage.setItem(STORAGE_KEY, next);
  } catch {
    // приватный режим — язык живёт до перезагрузки
  }
}

export function toggleLocale(): void {
  setLocale(current === "ru" ? "en" : "ru");
}

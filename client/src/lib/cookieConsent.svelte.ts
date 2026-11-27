/**
 * Общее состояние баннера согласия на cookie.
 *
 * Согласие хранится в localStorage под ключом `cookie-consent`. Модуль
 * используется и баннером, и кнопкой в футере: кнопка открывает карточку
 * заново, даже если согласие уже принято.
 */

const STORAGE_KEY = "cookie-consent";

let visible = $state(false);
let accepted = $state(false);
let loaded = $state(false);

function read(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === "accepted";
  } catch {
    // Приватный режим или отключённое хранилище: согласие живёт до перезагрузки.
    return false;
  }
}

function ensureLoaded(): void {
  if (loaded) return;
  accepted = read();
  visible = !accepted;
  loaded = true;
}

/** Инициализация при монтировании приложения. */
export function initCookieConsent(): void {
  ensureLoaded();
}

/** Открыт ли баннер прямо сейчас. */
export function isCookieBannerOpen(): boolean {
  return visible;
}

/** Принято ли согласие в этом браузере. */
export function isCookieAccepted(): boolean {
  return accepted;
}

/** Открыть карточку согласия из футера. */
export function openCookieConsent(): void {
  ensureLoaded();
  visible = true;
}

/** Закрыть карточку без сохранения согласия. */
export function closeCookieConsent(): void {
  visible = false;
}

/** Принять согласие и запомнить его между перезагрузками. */
export function acceptCookieConsent(): void {
  accepted = true;
  visible = false;

  try {
    localStorage.setItem(STORAGE_KEY, "accepted");
  } catch {
    // localStorage недоступен — скрываем баннер до конца сессии.
  }
}

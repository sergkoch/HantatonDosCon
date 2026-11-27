const STORAGE_KEY = "cookie-consent";

function readAccepted(): boolean {
  if (typeof localStorage === "undefined") {
    return false;
  }

  try {
    return localStorage.getItem(STORAGE_KEY) === "accepted";
  } catch {
    return false;
  }
}

let accepted = $state(readAccepted());
let open = $state(false);
let started = false;

export function consentAccepted(): boolean {
  return accepted;
}

export function consentOpen(): boolean {
  return open;
}

/** Открывает карточку согласия (в том числе повторно из футера). */
export function openConsent(): void {
  open = true;
}

/** Просто закрывает карточку, не записывая согласие. */
export function closeConsent(): void {
  open = false;
}

export function acceptConsent(): void {
  accepted = true;
  open = false;

  try {
    localStorage.setItem(STORAGE_KEY, "accepted");
  } catch {
    // приватный режим: согласие живёт до перезагрузки
  }
}

/** Показывает карточку при первом визите, если согласия ещё нет. */
export function initConsent(): void {
  if (started) {
    return;
  }
  started = true;

  if (!accepted) {
    open = true;
  }
}

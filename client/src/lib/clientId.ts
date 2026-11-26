const STORAGE_KEY = "ac.client-id";

let memoryID = "";

function randomID(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join(
    "",
  );
}

/**
 * Идентификатор браузера: история проверок хранится только для него.
 * Лежит в localStorage; при недоступном хранилище живёт до перезагрузки.
 */
export function clientID(): string {
  if (memoryID !== "") {
    return memoryID;
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && /^[a-z0-9]{8,64}$/i.test(stored)) {
      memoryID = stored;
      return stored;
    }

    const created = randomID();
    localStorage.setItem(STORAGE_KEY, created);
    memoryID = created;
    return created;
  } catch {
    // Приватный режим или отключённое хранилище: история живёт до перезагрузки.
    memoryID = randomID();
    return memoryID;
  }
}

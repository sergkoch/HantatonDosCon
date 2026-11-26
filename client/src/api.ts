import { clientID } from "./lib/clientId";
import { getLocale, t } from "./lib/i18n.svelte";
import type { MessageKey } from "./lib/messages";
import type { Scan } from "./types";

// Коды ошибок API → ключи сообщений. Текст берём из словаря текущего языка.
const errorKeys: Record<string, MessageKey> = {
  not_found: "error.notFound",
  invalid_input: "error.invalidInput",
  rate_limited: "error.rateLimited",
  internal_error: "error.internal",
};

function requestHeaders(extra: Record<string, string> = {}): Record<string, string> {
  return { ...extra, "X-Client-ID": clientID(), "X-Lang": getLocale() };
}

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { detail?: string; error?: string };
    const key = body.error ? errorKeys[body.error] : undefined;
    if (key) {
      return t(key);
    }
    return body.detail ?? t("error.http", { status: response.status });
  } catch {
    return t("error.http", { status: response.status });
  }
}

export async function createScan(url: string, depth: number): Promise<Scan> {
  const response = await fetch("/api/scans", {
    method: "POST",
    headers: requestHeaders({ "Content-Type": "application/json" }),
    body: JSON.stringify({ url, depth }),
  });

  if (!response.ok) {
    throw new Error(await readError(response));
  }

  return (await response.json()) as Scan;
}

export async function getScan(id: string): Promise<Scan> {
  const response = await fetch(`/api/scans/${id}`, {
    headers: requestHeaders(),
  });

  if (!response.ok) {
    throw new Error(await readError(response));
  }

  return (await response.json()) as Scan;
}

export async function listScans(url = "", limit = 20): Promise<Scan[]> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (url.trim() !== "") {
    params.set("url", url.trim());
  }

  const response = await fetch(`/api/scans?${params.toString()}`, {
    headers: requestHeaders(),
  });

  if (!response.ok) {
    throw new Error(await readError(response));
  }

  return (await response.json()) as Scan[];
}

/**
 * PDF-отчёт отдаётся только владельцу скана, поэтому качаем через fetch:
 * обычный переход по ссылке не отправит X-Client-ID.
 */
export async function getScanPdf(
  id: string,
  lang: "ru" | "en",
): Promise<Blob> {
  const response = await fetch(`/api/scans/${id}/pdf?lang=${lang}`, {
    headers: requestHeaders(),
  });

  if (!response.ok) {
    throw new Error(await readError(response));
  }

  return await response.blob();
}

/** Удаляет все проверки текущего браузера. */
export async function clearHistory(): Promise<number> {
  const response = await fetch("/api/scans", {
    method: "DELETE",
    headers: requestHeaders(),
  });

  if (!response.ok) {
    throw new Error(await readError(response));
  }

  const body = (await response.json()) as { deleted?: number };
  return body.deleted ?? 0;
}

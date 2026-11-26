import { clientID } from "./lib/clientId";
import type { Scan } from "./types";

function requestHeaders(extra: Record<string, string> = {}): Record<string, string> {
  return { ...extra, "X-Client-ID": clientID() };
}

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { detail?: string; error?: string };
    return body.detail ?? body.error ?? `Ошибка ${response.status}`;
  } catch {
    return `Ошибка ${response.status}`;
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

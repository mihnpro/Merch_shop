import { getAccessToken } from "../lib/tokens";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
const BASE_URL = `${API_BASE}/api/v1`;

export async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");

  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  return parseResponse<T>(response);
}

export async function uploadFile<T>(
  path: string,
  file: File,
  field = "file",
): Promise<T> {
  const form = new FormData();
  form.append(field, file);

  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    body: form,
    headers,
  });

  return parseResponse<T>(response);
}

async function parseResponse<T>(response: Response): Promise<T> {
  const text = await response.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!response.ok) {
    const message = data?.message ?? `Ошибка запроса (${response.status})`;
    throw new Error(message);
  }

  return data as T;
}

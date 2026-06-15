
const API_BASE = import.meta.env.VITE_API_URL ?? "";
const BASE_URL = `${API_BASE}/api/v1`;

export async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  return parseResponse<T>(response);
}

export async function uploadFile<T>(
  path: string,
  file: File,
  field = "file",
): Promise<T> {
  const form = new FormData();
  form.append(field, file);

  const response = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    body: form,
    credentials: "include",
  });

  return parseResponse<T>(response);
}

async function parseResponse<T>(response: Response): Promise<T> {
  const text = await response.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!response.ok) {
    if (response.status === 403 && data?.code === "USER_BLOCKED") {
      throw new Error("Ваш аккаунт заблокирован");
    }
    const message = data?.message ?? `Ошибка запроса (${response.status})`;
    throw new Error(message);
  }

  return data as T;
}

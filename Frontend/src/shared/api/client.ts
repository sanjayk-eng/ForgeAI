const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://127.0.0.1:8080";

export type ApiErrorPayload = {
  code: string;
  message: string;
};

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code = "UNKNOWN_ERROR") {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

type ApiResponse<T> = {
  data?: T;
  error?: Partial<ApiErrorPayload>;
};

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {}),
    },
  });

  const payload = (await response.json().catch(() => null)) as ApiResponse<T> | null;

  if (!response.ok) {
    throw new ApiError(
      payload?.error?.message ?? "Something went wrong",
      response.status,
      payload?.error?.code,
    );
  }

  return payload?.data as T;
}

export function apiUrl(path: string): string {
  return `${API_BASE_URL}${path}`;
}

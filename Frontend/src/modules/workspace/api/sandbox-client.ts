import { request } from "../../../shared/api/client";

export function authenticatedRequest<T>(
  accessToken: string,
  path: string,
  options: RequestInit = {},
) {
  return request<T>(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(options.headers ?? {}),
    },
  });
}
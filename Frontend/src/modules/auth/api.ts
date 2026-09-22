import { apiUrl, request } from "../../shared/api/client";
import type { AuthTokens, Credentials, Registration, UserProfile } from "../../shared/auth/types";

export function login(input: Credentials): Promise<AuthTokens> {
  return request<AuthTokens>("/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function register(input: Registration): Promise<AuthTokens> {
  return request<AuthTokens>("/auth/register", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function refresh(refreshToken: string): Promise<AuthTokens> {
  return request<AuthTokens>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

export function logout(): Promise<void> {
  return request<void>("/auth/logout", { method: "POST" });
}

export function getCurrentUser(accessToken: string): Promise<UserProfile> {
  return request<UserProfile>("/auth/me", {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function authenticateOAuth(provider: string, code: string): Promise<AuthTokens> {
  const query = `type=${encodeURIComponent(provider)}&code=${encodeURIComponent(code)}`;
  return request<AuthTokens>(`/auth/callback?${query}`);
}

export function oauthUrl(provider: "google" | "github"): string {
  const configuredUrl = provider === "google"
    ? import.meta.env.VITE_GOOGLE_OAUTH_URL
    : import.meta.env.VITE_GITHUB_OAUTH_URL;

  return configuredUrl || apiUrl(`/auth/callback?type=${provider}`);
}

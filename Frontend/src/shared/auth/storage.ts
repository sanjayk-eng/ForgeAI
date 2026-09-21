import type { AuthTokens } from "./types";

const TOKEN_KEY = "forgeai.auth.tokens";

export function readTokens(): AuthTokens | null {
  const value = localStorage.getItem(TOKEN_KEY);
  if (!value) return null;

  try {
    return JSON.parse(value) as AuthTokens;
  } catch {
    localStorage.removeItem(TOKEN_KEY);
    return null;
  }
}

export function writeTokens(tokens: AuthTokens): void {
  localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens));
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
}

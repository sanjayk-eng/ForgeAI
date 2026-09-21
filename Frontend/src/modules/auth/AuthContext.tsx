import { useCallback, useEffect, useState, type ReactNode } from "react";
import { getCurrentUser, login, refresh, register } from "./api";
import { clearTokens, readTokens, writeTokens } from "../../shared/auth/storage";
import type { AuthTokens, Credentials, Registration, UserProfile } from "../../shared/auth/types";
import { AuthContext } from "./authContextStore";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [tokens, setTokens] = useState<AuthTokens | null>(() => readTokens());
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    const initialTokens = readTokens();
    async function restoreSession() {
      if (!initialTokens) {
        setLoading(false);
        return;
      }

      try {
        const currentUser = await getCurrentUser(initialTokens.access_token);
        if (active) setUser(currentUser);
      } catch {
        try {
          const nextTokens = await refresh(initialTokens.refresh_token);
          writeTokens(nextTokens);
          const currentUser = await getCurrentUser(nextTokens.access_token);
          if (active) {
            setTokens(nextTokens);
            setUser(currentUser);
          }
        } catch {
          clearTokens();
          if (active) {
            setTokens(null);
            setUser(null);
          }
        }
      } finally {
        if (active) setLoading(false);
      }
    }

    void restoreSession();
    return () => {
      active = false;
    };
  }, []);

  const acceptTokens = useCallback(async (nextTokens: AuthTokens) => {
    writeTokens(nextTokens);
    setTokens(nextTokens);
    try {
      const currentUser = await getCurrentUser(nextTokens.access_token);
      setUser(currentUser);
    } catch {
      setUser(null);
    }
  }, []);

  const signIn = useCallback(
    async (input: Credentials) => {
      await acceptTokens(await login(input));
    },
    [acceptTokens],
  );

  const signUp = useCallback(
    async (input: Registration) => {
      await acceptTokens(await register(input));
    },
    [acceptTokens],
  );

  const signOut = useCallback(() => {
    clearTokens();
    setTokens(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, tokens, loading, signIn, signUp, acceptTokens, signOut }}
    >
      {children}
    </AuthContext.Provider>
  );
}

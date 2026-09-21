import { createContext } from "react";
import type { AuthTokens, Credentials, Registration, UserProfile } from "../../shared/auth/types";

export type AuthContextValue = {
  user: UserProfile | null;
  tokens: AuthTokens | null;
  loading: boolean;
  signIn(input: Credentials): Promise<void>;
  signUp(input: Registration): Promise<void>;
  acceptTokens(tokens: AuthTokens): Promise<void>;
  signOut(): void;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

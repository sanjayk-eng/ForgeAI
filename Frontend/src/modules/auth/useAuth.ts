import { useContext } from "react";
import { AuthContext } from "./authContextStore";
import type { AuthContextValue } from "./authContextStore";

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used inside AuthProvider");
  return value;
}

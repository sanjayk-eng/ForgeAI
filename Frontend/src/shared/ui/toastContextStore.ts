import { createContext } from "react";

export type ToastContextValue = {
  pushError(message: string): void;
  pushSuccess(message: string): void;
};

export const ToastContext = createContext<ToastContextValue | null>(null);

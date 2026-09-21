import { useState, type ReactNode } from "react";
import { CircleAlert, X } from "lucide-react";
import { ToastContext } from "./toastContextStore";

type Toast = {
  id: number;
  message: string;
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  function pushError(message: string) {
    const id = Date.now();
    setToasts((current) => [...current, { id, message }]);
    window.setTimeout(
      () => setToasts((current) => current.filter((toast) => toast.id !== id)),
      4500,
    );
  }

  function dismiss(id: number) {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }

  return (
    <ToastContext.Provider value={{ pushError }}>
      {children}
      <div
        className="pointer-events-none fixed right-5 top-5 z-50 grid w-[min(360px,calc(100vw-40px))] gap-3"
        aria-live="polite"
      >
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className="pointer-events-auto flex items-start gap-3 border border-red-200 bg-white p-4 text-sm text-red-900 shadow-xl"
            role="alert"
          >
            <CircleAlert className="mt-0.5 shrink-0 text-red-500" size={18} />
            <span className="flex-1 leading-5">{toast.message}</span>
            <button
              className="text-red-400 hover:text-red-700"
              aria-label="Dismiss error"
              onClick={() => dismiss(toast.id)}
            >
              <X size={16} />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

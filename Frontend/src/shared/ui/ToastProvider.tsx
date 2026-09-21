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
        aria-live="assertive"
      >
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className="animate-toast-enter pointer-events-auto relative flex items-start gap-3 overflow-hidden rounded-xl border border-red-300/30 bg-forge-panel/95 p-4 text-sm text-[#f1f3f5] shadow-[0_18px_50px_rgba(0,0,0,0.4)] backdrop-blur-md"
            role="alert"
          >
            <span className="mt-0.5 grid size-7 shrink-0 place-items-center rounded-lg border border-forge-signal/30 bg-forge-signal/10 text-forge-signal">
              <CircleAlert size={16} />
            </span>
            <span className="flex-1 leading-5 text-forge-soft">{toast.message}</span>
            <button
              className="grid size-7 place-items-center rounded-lg border-0 bg-transparent text-forge-muted transition hover:bg-white/10 hover:text-[#f1f3f5]"
              aria-label="Dismiss error"
              onClick={() => dismiss(toast.id)}
            >
              <X size={16} />
            </button>
            <span
              className="animate-toast-timer absolute inset-x-0 bottom-0 h-0.5 bg-forge-signal"
              aria-hidden="true"
            />
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

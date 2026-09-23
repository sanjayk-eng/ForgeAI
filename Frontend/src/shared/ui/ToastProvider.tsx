import { useState, type ReactNode } from "react";
import { CircleAlert, CheckCircle2, X } from "lucide-react";
import { ToastContext } from "./toastContextStore";

type Toast = {
  id: number;
  message: string;
  type: "error" | "success";
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  function pushError(message: string) {
    const id = Date.now();
    setToasts((current) => [...current, { id, message, type: "error" }]);
    window.setTimeout(
      () => setToasts((current) => current.filter((toast) => toast.id !== id)),
      4500,
    );
  }

  function pushSuccess(message: string) {
    const id = Date.now();
    setToasts((current) => [...current, { id, message, type: "success" }]);
    window.setTimeout(
      () => setToasts((current) => current.filter((toast) => toast.id !== id)),
      4500,
    );
  }

  function dismiss(id: number) {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }

  return (
    <ToastContext.Provider value={{ pushError, pushSuccess }}>
      {children}
      <div
        className="pointer-events-none fixed right-5 top-5 z-50 grid w-[min(360px,calc(100vw-40px))] gap-3"
        aria-live="assertive"
      >
        {toasts.map((toast) => {
          const isError = toast.type === "error";
          return (
            <div
              key={toast.id}
              className={`animate-toast-enter pointer-events-auto relative flex items-start gap-3 overflow-hidden rounded-xl border bg-forge-panel/95 p-4 text-sm text-[#f1f3f5] shadow-[0_18px_50px_rgba(0,0,0,0.4)] backdrop-blur-md ${
                isError ? "border-red-300/30" : "border-green-300/30"
              }`}
              role="alert"
            >
              <span
                className={`mt-0.5 grid size-7 shrink-0 place-items-center rounded-lg border ${
                  isError
                    ? "border-forge-signal/30 bg-forge-signal/10 text-forge-signal"
                    : "border-forge-accent/30 bg-forge-accent/10 text-forge-accent"
                }`}
              >
                {isError ? <CircleAlert size={16} /> : <CheckCircle2 size={16} />}
              </span>
              <span className="flex-1 leading-5 text-forge-soft">
                {toast.message}
              </span>
              <button
                className="grid size-7 place-items-center rounded-lg border-0 bg-transparent text-forge-muted transition hover:bg-white/10 hover:text-[#f1f3f5]"
                aria-label={`Dismiss ${toast.type}`}
                onClick={() => dismiss(toast.id)}
              >
                <X size={16} />
              </button>
              <span
                className={`animate-toast-timer absolute inset-x-0 bottom-0 h-0.5 ${
                  isError ? "bg-forge-signal" : "bg-forge-accent"
                }`}
                aria-hidden="true"
              />
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}

import { LoaderCircle } from "lucide-react";

export function LoadingOverlay({ message }: { message: string }) {
  return (
    <div className="absolute inset-0 z-10 grid place-items-center bg-[var(--overlay)]/75 p-6 backdrop-blur-[2px]" role="status" aria-live="polite">
      <div className="flex items-center gap-3 border border-forge-accent/25 bg-forge-card px-5 py-4 text-sm font-bold text-forge-text shadow-2xl">
        <LoaderCircle size={18} className="animate-spin text-forge-accent" />
        {message}
      </div>
    </div>
  );
}
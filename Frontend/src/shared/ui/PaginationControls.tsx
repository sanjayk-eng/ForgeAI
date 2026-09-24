import { ChevronLeft, ChevronRight } from "lucide-react";

type PaginationControlsProps = {
  page: number;
  totalPages: number;
  total: number;
  onPageChange: (page: number) => void;
};

export function PaginationControls({ page, totalPages, total, onPageChange }: PaginationControlsProps) {
  if (totalPages <= 1) return null;
  return (
    <div className="flex items-center justify-between gap-4 border-t border-white/[0.08] pt-4 text-xs text-forge-muted">
      <span>Showing page {page} of {totalPages} · {total} total</span>
      <div className="flex gap-2">
        <button type="button" className="grid size-9 place-items-center border border-white/[0.1] transition hover:border-forge-accent/40 hover:text-forge-text disabled:opacity-40" onClick={() => onPageChange(page - 1)} disabled={page <= 1} aria-label="Previous page"><ChevronLeft size={15} /></button>
        <button type="button" className="grid size-9 place-items-center border border-white/[0.1] transition hover:border-forge-accent/40 hover:text-forge-text disabled:opacity-40" onClick={() => onPageChange(page + 1)} disabled={page >= totalPages} aria-label="Next page"><ChevronRight size={15} /></button>
      </div>
    </div>
  );
}

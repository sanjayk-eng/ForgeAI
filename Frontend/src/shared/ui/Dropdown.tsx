import { useEffect, type ReactNode } from "react";

type DropdownProps = {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
  className?: string;
  align?: "left" | "right";
};

export function Dropdown({ 
  isOpen, 
  onClose, 
  children, 
  className = "",
  align = "right" 
}: DropdownProps) {
  useEffect(() => {
    if (!isOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 z-40"
        onClick={onClose}
        aria-hidden="true"
      />
      
      {/* Dropdown content */}
      <div
        className={`absolute ${align === "right" ? "right-0" : "left-0"} top-full z-50 mt-2 rounded-lg border border-[var(--border)] bg-forge-card shadow-[0_8px_30px_rgba(0,0,0,0.24)] ${className}`}
        role="menu"
        aria-label="Dropdown menu"
      >
        {children}
      </div>
    </>
  );
}

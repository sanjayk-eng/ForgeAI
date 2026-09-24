import { useState, type FormEvent, type ReactNode } from "react";
import { Eye, EyeOff, LoaderCircle } from "lucide-react";

type AuthFormProps = {
  children: ReactNode;
  onSubmit: () => Promise<void>;
  submitLabel: string;
  error: string;
};

export function AuthForm({ children, onSubmit, submitLabel, error }: AuthFormProps) {
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit}>
      {children}
      {error && (
        <p className="m-0 text-xs leading-[1.4] text-red-300" role="alert">
          {error}
        </p>
      )}
      <button
        className="mt-1 flex h-12 items-center justify-center gap-2 rounded-[10px] border-0 bg-forge-accent font-extrabold text-[var(--primary-foreground)] shadow-[0_8px_24px_rgba(198,243,106,0.14)] transition hover:-translate-y-px hover:bg-forge-accent-strong hover:shadow-[0_12px_30px_rgba(198,243,106,0.2)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-forge-accent disabled:cursor-wait disabled:opacity-70 disabled:hover:translate-y-0"
        type="submit"
        disabled={submitting}
      >
        {submitting && <LoaderCircle className="animate-spin" size={17} />}
        {submitting ? "Working..." : submitLabel}
      </button>
    </form>
  );
}

type PasswordFieldProps = {
  value: string;
  onChange: (value: string) => void;
};

export function PasswordField({ value, onChange }: PasswordFieldProps) {
  const [visible, setVisible] = useState(false);

  return (
    <label className="grid gap-2 text-xs font-bold text-forge-soft">
      <span>Password</span>
      <span className="relative">
        <input
          className="h-12 w-full rounded-[10px] border border-[var(--border)] bg-[var(--input)] px-3.5 pr-12 text-sm text-forge-text outline-none transition placeholder:text-forge-muted focus:border-forge-accent focus:ring-4 focus:ring-forge-accent/10"
          required
          minLength={8}
          type={visible ? "text" : "password"}
          value={value}
          onChange={(event) => onChange(event.target.value)}
        />
        <button
          type="button"
          className="absolute right-2 top-0 grid h-12 place-items-center border-0 bg-transparent text-forge-soft transition hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
          aria-label={visible ? "Hide password" : "Show password"}
          onClick={() => setVisible(!visible)}
        >
          {visible ? <EyeOff size={17} /> : <Eye size={17} />}
        </button>
      </span>
    </label>
  );
}

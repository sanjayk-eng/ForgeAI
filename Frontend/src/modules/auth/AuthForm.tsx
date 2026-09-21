import { useState, type FormEvent, type ReactNode } from "react";
import { Eye, EyeOff, LoaderCircle } from "lucide-react";

export function AuthForm({ children, onSubmit, submitLabel, error }: { children: ReactNode; onSubmit: () => Promise<void>; submitLabel: string; error: string }) {
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

  return <form className="auth-form" onSubmit={handleSubmit}>
    {children}
    {error && <p className="form-error" role="alert">{error}</p>}
    <button className="primary-button" type="submit" disabled={submitting}>
      {submitting ? <LoaderCircle className="spin" size={17} /> : null}
      {submitting ? "Working..." : submitLabel}
    </button>
  </form>;
}

export function PasswordField({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  const [visible, setVisible] = useState(false);
  return <label className="field">
    <span>Password</span>
    <span className="password-input">
      <input required minLength={8} type={visible ? "text" : "password"} value={value} onChange={(event) => onChange(event.target.value)} />
      <button type="button" className="icon-button" aria-label={visible ? "Hide password" : "Show password"} onClick={() => setVisible(!visible)}>
        {visible ? <EyeOff size={17} /> : <Eye size={17} />}
      </button>
    </span>
  </label>;
}

import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { apiUrl } from "../../shared/api/client";

export function VerifyEmailPage() {
  const location = useLocation();
  const [message, setMessage] = useState("Verifying your email...");
  const [verified, setVerified] = useState(false);

  useEffect(() => {
    const token = new URLSearchParams(location.search).get("token");
    if (!token) {
      setMessage("This verification link is missing its token.");
      return;
    }

    fetch(apiUrl(`/auth/verify-email?token=${encodeURIComponent(token)}`))
      .then(async (response) => {
        const payload = await response.json();
        if (!response.ok) throw new Error(payload?.error?.message ?? "Email verification failed");
        setVerified(true);
        setMessage("Your email is verified. You can now sign in.");
      })
      .catch((reason) => {
        setMessage(reason instanceof Error ? reason.message : "Email verification failed");
      });
  }, [location.search]);

  return (
    <div className="mx-auto w-full max-w-[420px] text-center">
      <p className="m-0 font-mono text-[11px] uppercase tracking-[0.08em] text-forge-accent">
        Email verification
      </p>
      <h2 className="my-2 text-4xl font-extrabold tracking-[-0.05em]">
        {verified ? "You are verified." : "Check your link."}
      </h2>
      <p className="m-0 text-sm leading-[1.5] text-forge-soft" role="status">
        {message}
      </p>
      {verified && (
        <Link className="mt-5 inline-flex font-bold text-forge-accent no-underline" to="/login">
          Return to sign in
        </Link>
      )}
    </div>
  );
}

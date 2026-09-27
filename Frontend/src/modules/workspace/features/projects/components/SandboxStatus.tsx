import { Loader2, CheckCircle2, AlertCircle, XCircle, Container } from "lucide-react";
import type { SandboxStatus as Status } from "../../../api/sandbox.api";

interface SandboxStatusProps {
  status: Status;
  showLabel?: boolean;
  className?: string;
}

export function SandboxStatus({ status, showLabel = true, className = "" }: SandboxStatusProps) {
  const config = getStatusConfig(status);

  return (
    <div className={`inline-flex items-center gap-2 ${className}`}>
      {config.icon}
      {showLabel && (
        <span className={`text-xs font-medium ${config.color}`}>
          {config.label}
        </span>
      )}
    </div>
  );
}

function getStatusConfig(status: Status) {
  switch (status) {
    case "CREATING":
    case "STARTING":
    case "RESTARTING":
      return {
        icon: <Loader2 size={14} className="animate-spin text-sky-400" />,
        label: status === "CREATING" ? "Setting up..." : status === "STARTING" ? "Starting..." : "Restarting...",
        color: "text-sky-400",
      };
    case "RUNNING":
      return {
        icon: <CheckCircle2 size={14} className="text-emerald-400" />,
        label: "Ready",
        color: "text-emerald-400",
      };
    case "CREATED":
    case "STOPPED":
      return {
        icon: <Container size={14} className="text-gray-400" />,
        label: status === "CREATED" ? "Created" : "Stopped",
        color: "text-gray-400",
      };
    case "STOPPING":
    case "DESTROYING":
      return {
        icon: <Loader2 size={14} className="animate-spin text-amber-400" />,
        label: status === "STOPPING" ? "Stopping..." : "Cleaning up...",
        color: "text-amber-400",
      };
    case "FAILED":
      return {
        icon: <AlertCircle size={14} className="text-red-400" />,
        label: "Failed",
        color: "text-red-400",
      };
    case "DESTROYED":
      return {
        icon: <XCircle size={14} className="text-gray-500" />,
        label: "Destroyed",
        color: "text-gray-500",
      };
    default:
      return {
        icon: <Container size={14} className="text-gray-400" />,
        label: status,
        color: "text-gray-400",
      };
  }
}

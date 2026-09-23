type NotificationBadgeProps = {
  count: number;
};

export function NotificationBadge({ count }: NotificationBadgeProps) {
  if (count === 0) return null;

  return (
    <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-forge-signal px-1 font-mono text-[9px] font-bold text-forge-bg">
      {count > 99 ? "99+" : count}
    </span>
  );
}

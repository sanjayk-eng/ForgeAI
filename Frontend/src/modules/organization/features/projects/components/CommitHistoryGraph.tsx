import type { GitCommit } from "../../../api/git.api";

interface CommitHistoryGraphProps {
  commits: GitCommit[];
  isLoading: boolean;
  isFetching: boolean;
  error: Error | null;
  onRefresh: () => void;
}

const laneColors = ["#c084fc", "#38bdf8", "#fbbf24", "#34d399", "#fb7185", "#818cf8"];
const rowHeight = 44;
const laneWidth = 18;

export function CommitHistoryGraph({ commits, isLoading, isFetching, error, onRefresh }: CommitHistoryGraphProps) {
  const layout = layoutCommits(commits);
  const graphWidth = Math.max(24, (layout.laneCount - 1) * laneWidth + 24);

  return (
    <section className="shrink-0 border border-[var(--border)] bg-forge-bg">
      <div className="flex h-9 items-center justify-between border-b border-[var(--border)] px-3 text-[10px] font-semibold uppercase tracking-[0.12em] text-forge-muted">
        <span>Commit history · {commits.length}</span>
        <button type="button" onClick={onRefresh} disabled={isFetching} className="normal-case tracking-normal hover:text-forge-text disabled:opacity-50">
          Refresh
        </button>
      </div>
      {isLoading ? (
        <p className="px-3 py-3 text-xs text-forge-muted">Loading commit history</p>
      ) : error ? (
        <p className="px-3 py-3 text-xs text-forge-signal">{error.message || "Could not load commit history."}</p>
      ) : commits.length === 0 ? (
        <p className="px-3 py-3 text-xs text-forge-muted">No commits yet. Create a commit to start project history.</p>
      ) : (
        <div className="max-h-44 overflow-auto">
          <div className="relative min-w-max" style={{ minHeight: commits.length * rowHeight }}>
            <svg
              aria-hidden="true"
              className="pointer-events-none absolute left-0 top-0"
              width={graphWidth}
              height={commits.length * rowHeight}
            >
              {layout.rows.flatMap((row) => row.parents.map((parent) => {
                const target = layout.rows[parent.index];
                const startX = laneX(row.lane);
                const endX = laneX(target.lane);
                const startY = row.index * rowHeight + rowHeight / 2;
                const endY = target.index * rowHeight + rowHeight / 2;
                return (
                  <path
                    key={`${row.commit.hash}-${parent.hash}`}
                    d={`M ${startX} ${startY} C ${startX} ${startY + 18}, ${endX} ${endY - 18}, ${endX} ${endY}`}
                    fill="none"
                    stroke={laneColors[row.lane % laneColors.length]}
                    strokeWidth="2"
                  />
                );
              }))}
              {layout.rows.map((row) => (
                <circle
                  key={row.commit.hash}
                  cx={laneX(row.lane)}
                  cy={row.index * rowHeight + rowHeight / 2}
                  r="4"
                  fill="var(--forge-bg)"
                  stroke={laneColors[row.lane % laneColors.length]}
                  strokeWidth="2"
                />
              ))}
            </svg>
            <ol className="m-0 list-none p-0">
              {layout.rows.map(({ commit, lane }) => (
                <li
                  key={commit.hash}
                  className="flex h-11 items-center gap-3 border-b border-[var(--border)]/60 pr-3 text-xs"
                  style={{ paddingLeft: graphWidth }}
                >
                  <span className="min-w-0 flex-1 truncate text-forge-text" title={commit.subject}>
                    {commit.subject}
                  </span>
                  {(commit.refs ?? []).slice(0, 2).map((ref) => (
                    <span key={ref} className="max-w-28 truncate rounded border border-forge-accent/30 px-1.5 py-0.5 font-mono text-[9px] text-forge-accent" title={ref}>
                      {ref}
                    </span>
                  ))}
                  <span className="max-w-24 truncate text-forge-muted" title={commit.author}>{commit.author}</span>
                  <time className="shrink-0 text-forge-muted" dateTime={commit.time}>
                    {formatCommitDate(commit.time)}
                  </time>
                  <code className="shrink-0 font-mono text-forge-accent">{(commit.hash ?? "").slice(0, 7)}</code>
                  <span className="sr-only" style={{ color: laneColors[lane % laneColors.length] }}>
                    Lane {lane + 1}
                  </span>
                </li>
              ))}
            </ol>
          </div>
        </div>
      )}
    </section>
  );
}

function laneX(lane: number) {
  return 12 + lane * laneWidth;
}

function layoutCommits(commits: GitCommit[]) {
  const lanes: (string | null)[] = [];
  const rows = commits.map((commit, index) => {
    let lane = lanes.indexOf(commit.hash);
    if (lane < 0) {
      lane = lanes.indexOf(null);
      if (lane < 0) lane = lanes.length;
    }
    const parentHashes = commit.parents ?? [];
    lanes[lane] = parentHashes[0] ?? null;
    const parents = parentHashes.map((hash, parentIndex) => {
      let parentLane = parentIndex === 0 ? lane : lanes.indexOf(hash);
      if (parentLane < 0) {
        parentLane = lanes.indexOf(null);
        if (parentLane < 0) parentLane = lanes.length;
        lanes[parentLane] = hash;
      }
      return { hash, lane: parentLane };
    });
    return { commit, index, lane, parents };
  });
  const rowByHash = new Map(rows.map((row) => [row.commit.hash, row]));

  return {
    laneCount: Math.max(1, lanes.length),
    rows: rows.map((row) => ({
      ...row,
      parents: row.parents.flatMap((parent) => {
        const target = rowByHash.get(parent.hash);
        return target && target.index > row.index ? [{ hash: parent.hash, index: target.index }] : [];
      }),
    })),
  };
}

function formatCommitDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

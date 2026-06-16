import { GitMerge, GitPullRequest, GitPullRequestClosed, FileText, ExternalLink } from 'lucide-react';
import type { PullRequestInfo } from '../../../stores/taskStore';

// PullRequestSummary renders an at-a-glance overview of the pull request that
// Octra opened for a GitHub issue/PR task, so the user does not have to leave the
// app and open GitHub to see what happened (issue #44, part 2). It intentionally
// only reads data the backend already returns — no extra GitHub round trips.

interface PullRequestSummaryProps {
  pr: PullRequestInfo;
}

function formatRange(pr: PullRequestInfo): string {
  const base = pr.baseBranch || 'main';
  if (pr.headBranch) {
    return `${pr.headBranch} → ${base}`;
  }
  return base;
}

function StatBlock({ label, value, tone }: { label: string; value: number | undefined; tone?: 'add' | 'del' }) {
  const display = typeof value === 'number' ? value.toLocaleString() : '—';
  const valueClass =
    tone === 'add'
      ? 'text-emerald-500'
      : tone === 'del'
        ? 'text-rose-500'
        : 'text-[var(--text)]';
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2">
      <span className={`text-lg font-semibold tabular-nums ${valueClass}`}>
        {tone === 'add' && typeof value === 'number' ? '+' : ''}
        {tone === 'del' && typeof value === 'number' ? '−' : ''}
        {display}
      </span>
      <span className="mt-0.5 text-[10px] font-medium uppercase tracking-wide text-[var(--text-muted)]">
        {label}
      </span>
    </div>
  );
}

export function PullRequestSummary({ pr }: PullRequestSummaryProps) {
  const isClosed = pr.state === 'closed';
  const isMerged = pr.state === 'merged';
  const statusLabel = isMerged ? 'Merged' : isClosed ? 'Closed' : 'Open';
  const statusClass = isMerged
    ? 'bg-violet-500/15 text-violet-500'
    : isClosed
      ? 'bg-rose-500/15 text-rose-500'
      : 'bg-emerald-500/15 text-emerald-500';
  const StatusIcon = isClosed ? GitPullRequestClosed : GitPullRequest;

  const title = pr.title || (pr.number ? `Pull request #${pr.number}` : 'Pull request');
  const files = pr.changedFiles ?? [];

  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 text-[var(--text)] shadow-sm">
      {/* Status pill + PR number */}
      <div className="flex items-center gap-2">
        <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold ${statusClass}`}>
          <StatusIcon size={13} />
          {statusLabel}
        </span>
        {pr.number ? (
          <span className="text-sm font-medium text-[var(--text-muted)]">#{pr.number}</span>
        ) : null}
        {pr.repository ? (
          <span className="truncate text-xs text-[var(--text-muted)]">{pr.repository}</span>
        ) : null}
      </div>

      {/* Title */}
      <h3 className="mt-2 text-base font-semibold leading-snug">{title}</h3>

      {/* head → base */}
      <p className="mt-1 text-xs text-[var(--text-muted)]">
        Wants to merge{' '}
        <span className="font-medium text-[var(--text)]">{formatRange(pr)}</span>
      </p>

      {/* Stats */}
      <div className="mt-3 grid grid-cols-3 gap-2">
        <StatBlock label="Commits" value={pr.commits} />
        <StatBlock label="Additions" value={pr.additions} tone="add" />
        <StatBlock label="Deletions" value={pr.deletions} tone="del" />
      </div>

      {/* Files changed */}
      {files.length > 0 && (
        <div className="mt-4">
          <div className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-[var(--text-muted)]">
            Files changed ({files.length})
          </div>
          <ul className="max-h-40 space-y-1 overflow-auto">
            {files.map((file) => (
              <li
                key={file}
                className="flex items-center gap-2 rounded-md border border-[var(--border)] bg-[var(--background)] px-2 py-1 text-xs"
              >
                <FileText size={13} className="shrink-0 text-[var(--text-muted)]" />
                <span className="truncate" title={file}>{file}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Actions */}
      <div className="mt-4 flex items-center gap-2">
        <a
          href={pr.url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-[var(--border)] px-3 py-2 text-sm font-medium text-[var(--text)] transition-colors hover:bg-[var(--background)]"
        >
          <ExternalLink size={15} />
          Review
        </a>
        <a
          href={pr.url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-[var(--accent)] px-3 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90"
        >
          <GitMerge size={15} />
          Merge Pull Request
        </a>
      </div>
      {pr.issueUrl ? (
        <a
          href={pr.issueUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-2 inline-block text-xs text-[var(--text-muted)] underline-offset-2 hover:underline"
        >
          View linked issue
        </a>
      ) : null}
    </div>
  );
}

import type { PullRequestInfo } from '../stores/taskStore';

// Pure helpers that turn the backend "success" payload into the pull request
// overview shown in the Solution pane (issue #44, part 2). Kept in lib/ — free
// of React/store side effects — so it can be unit-tested directly.

export function parseIntOrUndefined(value: unknown): number | undefined {
  const n = Number.parseInt(String(value ?? ''), 10);
  return Number.isNaN(n) ? undefined : n;
}

// Builds the PR overview from the success payload. Returns null unless the
// backend reported a pull request URL, so chat / non-GitHub tasks show no card.
export function parsePullRequestInfo(data?: Record<string, any>): PullRequestInfo | null {
  if (!data) return null;
  const url = typeof data.pullRequestUrl === 'string' ? data.pullRequestUrl.trim() : '';
  if (!url) return null;

  let changedFiles: string[] | undefined;
  const rawFiles = data.pullRequestFiles;
  if (Array.isArray(rawFiles)) {
    changedFiles = rawFiles.filter((f: unknown): f is string => typeof f === 'string');
  } else if (typeof rawFiles === 'string' && rawFiles.trim()) {
    try {
      const parsed = JSON.parse(rawFiles);
      if (Array.isArray(parsed)) {
        changedFiles = parsed.filter((f: unknown): f is string => typeof f === 'string');
      }
    } catch (err) {
      console.warn('[WS] Failed to parse pullRequestFiles payload:', err);
    }
  }

  const stateRaw = typeof data.pullRequestState === 'string' ? data.pullRequestState.trim().toLowerCase() : '';
  const state = stateRaw === 'closed' || stateRaw === 'merged' || stateRaw === 'open' ? stateRaw : 'open';

  return {
    url,
    number: parseIntOrUndefined(data.pullRequestNumber),
    title: typeof data.pullRequestTitle === 'string' ? data.pullRequestTitle : undefined,
    state,
    repository: typeof data.pullRequestRepo === 'string' ? data.pullRequestRepo : undefined,
    baseBranch: typeof data.pullRequestBase === 'string' ? data.pullRequestBase : undefined,
    headBranch: typeof data.pullRequestHead === 'string' ? data.pullRequestHead : undefined,
    issueUrl: typeof data.githubIssueUrl === 'string' ? data.githubIssueUrl : undefined,
    commits: parseIntOrUndefined(data.pullRequestCommits),
    additions: parseIntOrUndefined(data.pullRequestAdditions),
    deletions: parseIntOrUndefined(data.pullRequestDeletions),
    changedFiles,
  };
}

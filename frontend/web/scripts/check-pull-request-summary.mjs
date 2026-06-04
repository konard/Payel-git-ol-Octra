import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { createServer } from 'vite';

// Regression coverage for issue #44, part 2: when Octra finishes a GitHub
// issue/PR task it now shows a pull request overview in the Solution pane so the
// user does not have to leave the app and open GitHub. This guards both the
// success-payload parser (parsePullRequestInfo) and the wiring that renders the
// PullRequestSummary card.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

// --- 1. Source-level wiring assertions. -----------------------------------
const viewer = read('src/app/components/SolutionViewer.tsx');
assert.match(viewer, /from '\.\/PullRequestSummary'/, 'SolutionViewer must import PullRequestSummary');
assert.match(viewer, /useTaskStore\(\(state\) => state\.pullRequest\)/, 'SolutionViewer must read pullRequest from the store');
assert.match(viewer, /<PullRequestSummary pr=\{pullRequest\} \/>/, 'SolutionViewer must render the PR summary');

const summary = read('src/app/components/PullRequestSummary.tsx');
assert.match(summary, /export function PullRequestSummary\(/, 'PullRequestSummary must be exported');
assert.match(summary, /Merge Pull Request/, 'PR summary must offer a Merge button');
assert.match(summary, /Review/, 'PR summary must offer a Review button');
assert.match(summary, /Commits/, 'PR summary must show a commits stat');
assert.match(summary, /Additions/, 'PR summary must show an additions stat');
assert.match(summary, /Deletions/, 'PR summary must show a deletions stat');
assert.match(summary, /Files changed/, 'PR summary must list changed files');

const store = read('src/stores/taskStore.ts');
assert.match(store, /pullRequest: PullRequestInfo \| null/, 'taskStore must hold a pullRequest');
assert.match(store, /setPullRequest:/, 'taskStore must expose setPullRequest');

// --- 2. Behavioural test of the success-payload parser. -------------------
const server = await createServer({
  root,
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: { middlewareMode: true },
  optimizeDeps: { noDiscovery: true, include: [] },
  appType: 'custom',
});

try {
  const { parsePullRequestInfo } = await server.ssrLoadModule('/src/lib/pullRequest.ts');

  // No URL → no PR overview (chat/non-GitHub tasks must not show a card).
  assert.equal(parsePullRequestInfo(undefined), null, 'missing data yields no PR');
  assert.equal(parsePullRequestInfo({}), null, 'empty data yields no PR');
  assert.equal(parsePullRequestInfo({ pullRequestState: 'open' }), null, 'state without a URL yields no PR');

  // Full backend payload (files arrive as a JSON-encoded string from Go).
  const pr = parsePullRequestInfo({
    pullRequestUrl: 'https://github.com/octra-labs/app/pull/12',
    pullRequestNumber: '12',
    pullRequestTitle: 'Fix #7: crash on paste',
    pullRequestRepo: 'octra-labs/app',
    pullRequestBase: 'main',
    pullRequestHead: 'issue-7-abc',
    pullRequestState: 'open',
    githubIssueUrl: 'https://github.com/octra-labs/app/issues/7',
    pullRequestCommits: '3',
    pullRequestAdditions: '42',
    pullRequestDeletions: '5',
    pullRequestFiles: JSON.stringify(['main.go', 'README.md']),
  });
  assert.ok(pr, 'a payload with a pull request URL produces an overview');
  assert.equal(pr.url, 'https://github.com/octra-labs/app/pull/12');
  assert.equal(pr.number, 12, 'PR number is parsed to an int');
  assert.equal(pr.title, 'Fix #7: crash on paste');
  assert.equal(pr.repository, 'octra-labs/app');
  assert.equal(pr.baseBranch, 'main');
  assert.equal(pr.headBranch, 'issue-7-abc');
  assert.equal(pr.state, 'open');
  assert.equal(pr.issueUrl, 'https://github.com/octra-labs/app/issues/7');
  assert.equal(pr.commits, 3);
  assert.equal(pr.additions, 42);
  assert.equal(pr.deletions, 5);
  assert.deepEqual(pr.changedFiles, ['main.go', 'README.md'], 'changed files decode from the JSON string');

  // Unknown states fall back to "open"; files may also arrive as a real array.
  const fallback = parsePullRequestInfo({
    pullRequestUrl: 'https://github.com/o/r/pull/1',
    pullRequestState: 'weird',
    pullRequestFiles: ['a.ts'],
  });
  assert.equal(fallback.state, 'open', 'unknown state falls back to open');
  assert.deepEqual(fallback.changedFiles, ['a.ts'], 'array files are kept as-is');
  assert.equal(fallback.number, undefined, 'missing number stays undefined');

  console.log('check-pull-request-summary: all assertions passed');
} finally {
  await server.close();
}

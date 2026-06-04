import { StrictMode, createElement } from 'react';
import { createRoot } from 'react-dom/client';
import '../../src/styles/index.css';
import { PullRequestSummary } from '../../src/app/components/PullRequestSummary';

const pr = {
  url: 'https://github.com/octra-labs/app/pull/1248',
  number: 1248,
  title: 'feat: implement kinetic-theming logic for mobile grid systems',
  state: 'open' as const,
  repository: 'octra-labs/app',
  baseBranch: 'main',
  headBranch: 'alex_dev_ops',
  issueUrl: 'https://github.com/octra-labs/app/issues/1240',
  commits: 14,
  additions: 412,
  deletions: 89,
  changedFiles: ['src/styles/tailwind.config.js', 'src/core/grid_system.py'],
};

createRoot(document.getElementById('root')!).render(
  createElement(StrictMode, null, createElement(PullRequestSummary, { pr })),
);

// Deterministic fixtures for the landing-page screenshot harness.
//
// Each surface seeds the real task store (and, for research, returns the chat
// messages) so the harness renders the genuine production components — TopBar,
// SolutionViewer and Chat — exactly as a user sees them. Keeping the data here
// makes the four landing screenshots reproducible: re-run the harness and the
// same UI is captured every time. See scripts/capture-landing.mjs.

import { useTaskStore } from '../stores/taskStore';
import type { ChatMessage } from '../app/components/chat/Chat';

export type Surface = 'code-view' | 'research-progress' | 'document-reader' | 'presentation-deck';

export interface SurfaceConfig {
  mode: 'canvas' | 'chat' | 'solution';
  messages: ChatMessage[];
}

const TASK_BOARD_TSX = `import { useMemo, useState } from 'react';
import { TaskCard } from './TaskCard';
import { fetchTasks, type Task } from '../lib/api';

const COLUMNS: Array<{ id: Task['status']; label: string }> = [
  { id: 'todo', label: 'To do' },
  { id: 'in_progress', label: 'In progress' },
  { id: 'done', label: 'Done' },
];

export function TaskBoard() {
  const [query, setQuery] = useState('');
  const [tasks, setTasks] = useState<Task[]>(() => fetchTasks());

  const visible = useMemo(
    () => tasks.filter((task) => task.title.toLowerCase().includes(query.toLowerCase())),
    [tasks, query],
  );

  const move = (id: string, status: Task['status']) =>
    setTasks((prev) => prev.map((task) => (task.id === id ? { ...task, status } : task)));

  return (
    <div className="board">
      <header className="board__header">
        <h1>Sprint board</h1>
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Filter tasks…"
        />
      </header>
      <div className="board__columns">
        {COLUMNS.map((column) => (
          <section key={column.id} className="board__column">
            <h2>{column.label}</h2>
            {visible
              .filter((task) => task.status === column.id)
              .map((task) => (
                <TaskCard key={task.id} task={task} onMove={move} />
              ))}
          </section>
        ))}
      </div>
    </div>
  );
}
`;

const TASK_CARD_TSX = `import type { Task } from '../lib/api';

interface TaskCardProps {
  task: Task;
  onMove: (id: string, status: Task['status']) => void;
}

export function TaskCard({ task, onMove }: TaskCardProps) {
  return (
    <article className="card" data-status={task.status}>
      <h3>{task.title}</h3>
      <p>{task.summary}</p>
      <footer>
        <span className="card__owner">{task.owner}</span>
        <button onClick={() => onMove(task.id, 'done')}>Mark done</button>
      </footer>
    </article>
  );
}
`;

const API_TS = `export interface Task {
  id: string;
  title: string;
  summary: string;
  owner: string;
  status: 'todo' | 'in_progress' | 'done';
}

const SEED: Task[] = [
  { id: 't-1', title: 'Design sprint board', summary: 'Columns and cards', owner: 'Mia', status: 'done' },
  { id: 't-2', title: 'Wire up filtering', summary: 'Search by title', owner: 'Leo', status: 'in_progress' },
  { id: 't-3', title: 'Persist task moves', summary: 'Save column changes', owner: 'Ada', status: 'todo' },
];

export function fetchTasks(): Task[] {
  return SEED.map((task) => ({ ...task }));
}
`;

const MAIN_TSX = `import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { TaskBoard } from './components/TaskBoard';
import './styles.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <TaskBoard />
  </StrictMode>,
);
`;

const PACKAGE_JSON = `{
  "name": "sprint-board",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  }
}
`;

const QUARTERLY_REPORT_MD = `# Q2 2026 Business Review

A concise summary of performance, key wins, and focus areas for the next quarter.

## Highlights

- Revenue grew **18%** quarter over quarter, led by self-serve signups.
- Net retention reached **121%**, the highest in four quarters.
- Three enterprise pilots converted to annual contracts.

## Key metrics

| Metric | Q1 2026 | Q2 2026 | Change |
| --- | --- | --- | --- |
| Active workspaces | 4,210 | 5,640 | +34% |
| Paid seats | 1,180 | 1,520 | +29% |
| Monthly recurring revenue | $96k | $113k | +18% |
| Avg. tasks per workspace | 12 | 17 | +42% |

## What worked

1. The refreshed onboarding cut time-to-first-task to under five minutes.
2. Research and document tasks broadened usage beyond engineering teams.
3. Faster streaming made long-running runs feel responsive.

## Focus for next quarter

- Ship shared workspaces for cross-functional teams.
- Expand the template library for research and reporting.
- Continue reducing cold-start latency on large projects.

> Prepared by the Octra analytics workspace — figures are illustrative.
`;

const DECK_MD = `# Octra: Agents That Ship Real Work

## Why teams choose Octra
- One workspace for code, research, documents, and decks
- A Boss agent plans, Managers coordinate, Workers execute
- Every run is traceable from the prompt to the finished files
Visual: Split screen — task prompt on the left, finished files on the right
Source: Octra product overview, 2026
> Opening slide for the investor and onboarding decks.

## How a task flows
- Describe the goal in plain language
- Agents break it into manager and worker roles
- Results stream into the viewer as they are produced
Visual: Horizontal flow diagram of Boss to Managers to Workers
Source: Octra architecture notes

## What you get
- Review-ready code with a GitHub handoff
- Sourced research summaries
- Reports and slide decks from the same trace
Source: Octra capabilities matrix
`;

// A tiny but valid base64 payload so the PPTX preview shows its Download button
// without shipping a real binary into the repository.
const PPTX_BASE64 = 'UEsDBBQABgAIAAAAIQ4=';

function seedCodeView(): void {
  const store = useTaskStore.getState();
  store.clearCodeFiles();
  store.upsertCodeFiles([
    { path: 'package.json', name: 'package.json', language: 'json', content: PACKAGE_JSON, status: 'ready', workerRole: 'Frontend Engineer' },
    { path: 'src/main.tsx', name: 'main.tsx', language: 'typescript', content: MAIN_TSX, status: 'ready', workerRole: 'Frontend Engineer' },
    { path: 'src/lib/api.ts', name: 'api.ts', language: 'typescript', content: API_TS, status: 'ready', workerRole: 'Frontend Engineer' },
    { path: 'src/components/TaskCard.tsx', name: 'TaskCard.tsx', language: 'typescript', content: TASK_CARD_TSX, status: 'ready', workerRole: 'Frontend Engineer' },
    // Listed last so it becomes the latest (and therefore the active) file.
    { path: 'src/components/TaskBoard.tsx', name: 'TaskBoard.tsx', language: 'typescript', content: TASK_BOARD_TSX, status: 'ready', workerRole: 'Frontend Engineer' },
  ]);
}

function seedDocument(): void {
  const store = useTaskStore.getState();
  store.clearCodeFiles();
  store.upsertCodeFiles([
    { path: 'Quarterly-Business-Report.md', name: 'Quarterly-Business-Report.md', language: 'markdown', content: QUARTERLY_REPORT_MD, status: 'ready', workerRole: 'Research Analyst' },
  ]);
}

function seedPresentation(): void {
  const store = useTaskStore.getState();
  store.clearCodeFiles();
  store.upsertCodeFiles([
    { path: 'octra-deck.md', name: 'octra-deck.md', language: 'markdown', content: DECK_MD, status: 'ready', managerRole: 'Presentation Manager' },
    // Listed last so the .pptx becomes the active file and renders the slide preview.
    { path: 'octra-deck.pptx', name: 'octra-deck.pptx', language: 'plaintext', encoding: 'base64', content: PPTX_BASE64, status: 'ready', managerRole: 'Presentation Manager' },
  ]);
}

function seedResearch(): ChatMessage[] {
  const store = useTaskStore.getState();
  store.clearCodeFiles();
  const steps = [
    'Searching the web for "global AI agent platform market size 2026"',
    'Reading: Gartner — Emerging Tech: AI Agents (2026)',
    'Comparing adoption across enterprise and self-serve segments',
    'Searching the web for "autonomous agent funding rounds Q1 2026"',
    'Extracting key figures from the McKinsey State of AI report',
    'Consolidating findings into a sourced summary',
  ];
  steps.forEach((step) => store.recordSearchStep(step, 'done', steps.length));

  return [
    {
      id: 'msg-research-1',
      sender: 'boss',
      read: true,
      timestamp: new Date('2026-06-10T09:24:00'),
      text:
        'Here is the market scan you asked for. I searched recent sources, compared the figures, and consolidated them into a short brief:\n\n' +
        '• The AI agent platform market is estimated at ~$8.5B in 2026, growing roughly 44% year over year.\n' +
        '• Enterprise adoption is led by engineering and research teams, with self-serve signups growing fastest.\n' +
        '• Funding stayed strong through Q1 2026, concentrated in orchestration and tooling startups.\n\n' +
        'Sources are listed in the steps above. Want me to turn this into a one-page report or a slide deck?',
    },
  ];
}

export function seedSurface(surface: Surface): SurfaceConfig {
  switch (surface) {
    case 'research-progress':
      return { mode: 'chat', messages: seedResearch() };
    case 'document-reader':
      seedDocument();
      return { mode: 'solution', messages: [] };
    case 'presentation-deck':
      seedPresentation();
      return { mode: 'solution', messages: [] };
    case 'code-view':
    default:
      seedCodeView();
      return { mode: 'solution', messages: [] };
  }
}

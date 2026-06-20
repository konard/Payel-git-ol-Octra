import { create } from 'zustand';

export type AgentNodeType = 'boss' | 'manager' | 'worker' | 'universal' | 'github';
export type AgentNodeStatus = 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
export type TaskStatus = 'idle' | 'creating' | 'planning' | 'executing' | 'done' | 'error' | 'cancelled';

export interface AgentNode {
  id: string;
  type: AgentNodeType;
  role: string;
  status: AgentNodeStatus;
  progress: number;
  filesCount?: number;
  workerCount?: number;
  techStack?: string[];
  position?: { x: number; y: number };
  // GitHub specific fields
  repoUrl?: string;
  prUrl?: string;
  commitCount?: number;
  // Human-readable reason shown when a node fails (issue #75 п.3: GitHub publish errors)
  errorMessage?: string;
  // N8n automation
  n8nTrigger?: 'start' | 'end' | 'middle' | 'custom';
  n8nPercentage?: number;
  n8nWorkflowId?: string;
  n8nWebhookUrl?: string;
  // Node scale
  scale?: number;
}

export interface WorkflowConfig {
  useAiPlanning: boolean;
  managers: Array<{
    role: string;
    description: string;
    priority: number;
    workers: Array<{ role: string; description: string }>;
  }>;
  architecture: string;
  techStack: string[];
}

export interface Edge {
  from: string;
  to: string;
}

export type CodeFileStatus = 'streaming' | 'ready';

export interface CodeFile {
  path: string;
  name: string;
  language: string;
  encoding?: string;
  content: string;
  status: CodeFileStatus;
  workerRole?: string;
  managerRole?: string;
  updatedAt: number;
}

export interface LogEntry {
  id: string;
  timestamp: Date;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
}

// PullRequestInfo — обзор созданного pull request для панели Solution, чтобы
// пользователю не приходилось возвращаться на GitHub (issue #44).
export interface PullRequestInfo {
  url: string;
  number?: number;
  title?: string;
  state?: 'open' | 'closed' | 'merged';
  repository?: string;
  baseBranch?: string;
  headBranch?: string;
  issueUrl?: string;
  commits?: number;
  additions?: number;
  deletions?: number;
  changedFiles?: string[];
}

export type SearchPhase = 'idle' | 'searching' | 'done';

// SearchStep — один пункт в блоке «Searching the web» в чате. text — человекочитаемая
// строка шага (например, «Searching the web for «httpx install python»»).
export interface SearchStep {
  id: string;
  text: string;
  provider?: string;
  model?: string;
  resultCount?: number;
}

interface TaskState {
  taskId: string | null;
  status: TaskStatus;
  nodes: AgentNode[];
  edges: Edge[];
  workflow: WorkflowConfig | null;
  logs: LogEntry[];
  solutionZip: Blob | null;
  zipUrl: string | null;
  codeFiles: CodeFile[];
  latestCodeFilePath: string | null;
  isConnected: boolean;
  tokensUsed: number;
  startTime: number | null;
  searchSteps: SearchStep[];
  searchPhase: SearchPhase;
  searchStepsCount: number;
  pullRequest: PullRequestInfo | null;
  nixStorePath: string | null;
  nixTaskId: string | null;

  // Actions
  setTaskId: (taskId: string) => void;
  setTaskStatus: (status: TaskStatus) => void;
  addNode: (node: Omit<AgentNode, 'progress'>) => void;
  updateNode: (id: string, updates: Partial<AgentNode>) => void;
  removeNode: (id: string) => void;
  addEdge: (edge: Edge) => void;
  addLog: (log: Omit<LogEntry, 'id' | 'timestamp'>) => void;
  clearLogs: () => void;
  setSolutionZip: (blob: Blob) => void;
  setZipUrl: (url: string) => void;
  upsertCodeFiles: (files: Array<Partial<CodeFile> & { path: string; content: string }>) => void;
  updateCodeFileContent: (path: string, content: string) => void;
  completeCodeStreaming: () => void;
  clearCodeFiles: () => void;
  setConnectionStatus: (connected: boolean) => void;
  recordSearchStep: (step: string, phase: SearchPhase, count: number, provider?: string, model?: string, resultCount?: number) => void;
  clearSearchSteps: () => void;
  setPullRequest: (pr: PullRequestInfo | null) => void;
  setNixStorePath: (path: string | null, taskId: string | null) => void;
  loadProjectFiles: (chatId: string) => Promise<void>;
  setTokensUsed: (tokens: number) => void;
  setStartTime: (time: number) => void;
  getWorkflow: () => WorkflowConfig | null;
  setWorkflow: (workflow: WorkflowConfig | null) => void;
  setGraph: (nodes: Array<Omit<AgentNode, 'progress'> & { progress?: number }>, edges: Edge[]) => void;
  resetTask: () => void;
  resetTaskExecution: () => void;
}

let nodeIdCounter = 0;

const getFileName = (path: string): string => path.split('/').filter(Boolean).at(-1) || path || 'Untitled';

export const isGeneratedAgentNodeId = (id: string): boolean =>
  id.startsWith('boss-') ||
  id.startsWith('manager-') ||
  id.startsWith('worker-') ||
  id.startsWith('universal-') ||
  id === 'github-archive' ||
  id === 'zip-archive';

const normalizeCodeFile = (
  file: Partial<CodeFile> & { path: string; content: string },
  existing?: CodeFile,
): CodeFile => ({
  path: file.path,
  name: file.name || existing?.name || getFileName(file.path),
  language: file.language || existing?.language || 'plaintext',
  encoding: file.encoding || existing?.encoding,
  content: file.content,
  status: file.status || existing?.status || 'streaming',
  workerRole: file.workerRole || existing?.workerRole,
  managerRole: file.managerRole || existing?.managerRole,
  updatedAt: file.updatedAt || Date.now(),
});

// Validation functions
const validateNodeUpdate = (updates: Partial<AgentNode>): Partial<AgentNode> => {
  const validatedUpdates = { ...updates };

  // Validate n8nPercentage is within 0-100 range
  if (updates.n8nPercentage !== undefined) {
    validatedUpdates.n8nPercentage = Math.max(0, Math.min(100, updates.n8nPercentage));
  }

  // Validate n8nTrigger is one of the allowed values
  if (updates.n8nTrigger !== undefined && !['start', 'end', 'middle', 'custom'].includes(updates.n8nTrigger)) {
    delete validatedUpdates.n8nTrigger;
  }

  return validatedUpdates;
};

export const useTaskStore = create<TaskState>((set, get) => ({
  taskId: null,
  status: 'idle',
  nodes: [],
  edges: [],
  workflow: null,
  logs: [],
  solutionZip: null,
  zipUrl: null,
  codeFiles: [],
  latestCodeFilePath: null,
  isConnected: false,
  tokensUsed: 0,
  startTime: null,
  searchSteps: [],
  searchPhase: 'idle',
  searchStepsCount: 0,
  pullRequest: null,
  nixStorePath: null,
  nixTaskId: null,

  setTaskId: (taskId) => set({ taskId }),
  
  setTaskStatus: (status) => set({ status }),
  
  addNode: (node) => set((state) => {
    const newNode: AgentNode = {
      ...node,
      progress: 0,
      id: node.id || `node-${++nodeIdCounter}-${Date.now()}`,
    };
    return { nodes: [...state.nodes, newNode] };
  }),
  
  updateNode: (id, updates) => set((state) => ({
    nodes: state.nodes.map((n) => (n.id === id ? { ...n, ...validateNodeUpdate(updates) } : n)),
  })),
  
  removeNode: (id) => set((state) => ({
    nodes: state.nodes.filter((n) => n.id !== id),
    edges: state.edges.filter((e) => e.from !== id && e.to !== id),
  })),
  
  addEdge: (edge) => set((state) => ({
    edges: [...state.edges, edge],
  })),
  
  addLog: (log) => set((state) => ({
    logs: [...state.logs, { ...log, id: `log-${Date.now()}`, timestamp: new Date() }],
  })),
  
  clearLogs: () => set({ logs: [] }),
  
  setSolutionZip: (blob) => set({ solutionZip: blob }),
  
  setZipUrl: (url) => set({ zipUrl: url }),

  upsertCodeFiles: (files) => set((state) => {
    if (files.length === 0) {
      return state;
    }

    const byPath = new Map(state.codeFiles.map((file) => [file.path, file]));
    files.forEach((file) => {
      byPath.set(file.path, normalizeCodeFile(file, byPath.get(file.path)));
    });

    const codeFiles = Array.from(byPath.values()).sort((a, b) => a.path.localeCompare(b.path));
    const latestCodeFilePath = files[files.length - 1]?.path ?? state.latestCodeFilePath;

    return { codeFiles, latestCodeFilePath };
  }),

  updateCodeFileContent: (path, content) => set((state) => ({
    codeFiles: state.codeFiles.map((file) => (
      file.path === path
        ? { ...file, content, status: 'ready', updatedAt: Date.now() }
        : file
    )),
    latestCodeFilePath: path,
  })),

  completeCodeStreaming: () => set((state) => ({
    codeFiles: state.codeFiles.map((file) => (
      file.status === 'streaming' ? { ...file, status: 'ready' } : file
    )),
  })),

  clearCodeFiles: () => set({
    codeFiles: [],
    latestCodeFilePath: null,
    searchSteps: [],
    searchPhase: 'idle',
    searchStepsCount: 0,
    pullRequest: null,
  }),
  
  setConnectionStatus: (connected) => set({ isConnected: connected }),

  // recordSearchStep — накапливает шаги веб-поиска для блока «Searching the web».
  // Воркеры присылают шаги по мере поиска (phase='searching') и финальное событие
  // завершения (phase='done', count = число выполненных шагов). Несколько воркеров
  // могут искать последовательно: новый шаг после 'done' снова переводит блок в
  // активное состояние, поэтому в конце пользователь видит «Completed N steps».
  recordSearchStep: (step, phase, count, provider?, model?, resultCount?) => set((state) => {
    let steps = state.searchSteps;
    if (step && !steps.some((s) => s.text === step)) {
      steps = [...steps, {
        id: `search-${steps.length}-${Date.now()}`,
        text: step,
        provider,
        model,
        resultCount,
      }];
    }
    const searchPhase: SearchPhase = phase === 'done' ? 'done' : 'searching';
    const searchStepsCount = count > 0 ? Math.max(state.searchStepsCount, count) : state.searchStepsCount;
    return { searchSteps: steps, searchPhase, searchStepsCount };
  }),

  clearSearchSteps: () => set({ searchSteps: [], searchPhase: 'idle', searchStepsCount: 0 }),

  setPullRequest: (pr) => set({ pullRequest: pr }),

  setNixStorePath: (path, taskId) => set({ nixStorePath: path, nixTaskId: taskId }),

  loadProjectFiles: async (chatId: string) => {
    try {
      const authApiUrl = (import.meta as any).env?.VITE_AUTH_URL || '';
      const response = await fetch(`${authApiUrl}/chat/${chatId}/files`, {
        credentials: 'include',
      });

      if (!response.ok) {
        console.warn('[loadProjectFiles] HTTP', response.status);
        return;
      }

      const json = await response.json();
      const files = json?.data?.files;
      if (!Array.isArray(files) || files.length === 0) return;

      const codeFiles = files.map((f: any) => ({
        path: f.path,
        name: f.path.split('/').filter(Boolean).at(-1) || f.path,
        language: f.language || 'plaintext',
        encoding: f.encoding || undefined,
        content: f.content,
        status: 'ready' as const,
        updatedAt: Date.now(),
      }));

      set({
        codeFiles,
        latestCodeFilePath: codeFiles[codeFiles.length - 1]?.path ?? null,
      });
    } catch (err) {
      console.error('Failed to load project files:', err);
    }
  },

  setTokensUsed: (tokens) => set({ tokensUsed: tokens }),

  setStartTime: (time) => set({ startTime: time }),

  getWorkflow: () => get().workflow,

  setWorkflow: (workflow) => set({ workflow }),

  // Fully replace the canvas graph (used when switching between chats so the
  // previous chat's nodes never leak into the next one) and clear any task
  // execution state tied to the old graph.
  setGraph: (nodes, edges) => set({
    nodes: nodes.map((node) => ({ ...node, progress: node.progress ?? 0 })),
    edges,
    taskId: null,
    status: 'idle',
    logs: [],
    solutionZip: null,
    zipUrl: null,
    codeFiles: [],
    latestCodeFilePath: null,
    tokensUsed: 0,
    startTime: null,
    searchSteps: [],
    searchPhase: 'idle',
    searchStepsCount: 0,
    pullRequest: null,
    nixStorePath: null,
    nixTaskId: null,
  }),

  resetTask: () => set((state) => {
    // Keep user-created nodes, not auto-generated runtime nodes.
    const userNodes = state.nodes.filter(node => !isGeneratedAgentNodeId(node.id));
    const userNodeIds = new Set(userNodes.map(node => node.id));

    // Keep only edges fully inside the preserved user workflow
    const userEdges = state.edges.filter(edge =>
      userNodeIds.has(edge.from) && userNodeIds.has(edge.to)
    );

    return {
      taskId: null,
      status: 'idle',
      nodes: userNodes,
      edges: userEdges,
      // Keep workflow to prevent user from losing their work
      workflow: state.workflow,
      logs: [],
      solutionZip: null,
      zipUrl: null,
      codeFiles: [],
      latestCodeFilePath: null,
      tokensUsed: 0,
      startTime: null,
      searchSteps: [],
      searchPhase: 'idle',
      searchStepsCount: 0,
      pullRequest: null,
      nixStorePath: null,
      nixTaskId: null,
    };
  }),

  // Clear execution state for a new run while preserving only user-authored
  // workflow nodes. Generated runtime nodes from the previous task must not
  // leak into the next graph (for example a stale Boss before a direct
  // Universal run).
  resetTaskExecution: () => set((state) => {
    const userNodes = state.nodes.filter(node => !isGeneratedAgentNodeId(node.id));
    const userNodeIds = new Set(userNodes.map(node => node.id));
    const userEdges = state.edges.filter(edge =>
      userNodeIds.has(edge.from) && userNodeIds.has(edge.to)
    );

    return {
      taskId: null,
      status: 'idle',
      nodes: userNodes,
      edges: userEdges,
      logs: [],
      solutionZip: null,
      zipUrl: null,
      codeFiles: [],
      latestCodeFilePath: null,
      tokensUsed: 0,
      startTime: null,
      searchSteps: [],
      searchPhase: 'idle',
      searchStepsCount: 0,
      pullRequest: null,
      nixStorePath: null,
      nixTaskId: null,
    };
  }),
}));

// Dev-only handle to inspect/seed the task store from the browser console
// (e.g. window.octraTaskStore.getState().upsertCodeFiles([...])). Off in prod.
if (import.meta.env.DEV && typeof window !== 'undefined') {
  (window as Window & { octraTaskStore?: typeof useTaskStore }).octraTaskStore = useTaskStore;
}

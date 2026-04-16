import { create } from 'zustand';

export type AgentNodeType = 'boss' | 'manager' | 'worker' | 'zip';
export type AgentNodeStatus = 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
export type TaskStatus = 'idle' | 'creating' | 'planning' | 'executing' | 'done' | 'error';

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
  // ZIP Archive specific fields
  fileName?: string;
  fileSize?: string;
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

export interface LogEntry {
  id: string;
  timestamp: Date;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
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
  isConnected: boolean;
  tokensUsed: number;
  startTime: number | null;
  
  // Actions
  setTaskId: (taskId: string) => void;
  setTaskStatus: (status: TaskStatus) => void;
  addNode: (node: Omit<AgentNode, 'progress'>) => void;
  updateNode: (id: string, updates: Partial<AgentNode>) => void;
  removeNode: (id: string) => void;
  addEdge: (edge: Edge) => void;
  addLog: (log: Omit<LogEntry, 'id' | 'timestamp'>) => void;
  setSolutionZip: (blob: Blob) => void;
  setZipUrl: (url: string) => void;
  setConnectionStatus: (connected: boolean) => void;
  setTokensUsed: (tokens: number) => void;
  setStartTime: (time: number) => void;
  getWorkflow: () => WorkflowConfig | null;
  setWorkflow: (workflow: WorkflowConfig | null) => void;
  resetTask: () => void;
}

let nodeIdCounter = 0;

export const useTaskStore = create<TaskState>((set, get) => ({
  taskId: null,
  status: 'idle',
  nodes: [],
  edges: [],
  workflow: null,
  logs: [],
  solutionZip: null,
  zipUrl: null,
  isConnected: false,
  tokensUsed: 0,
  startTime: null,

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
    nodes: state.nodes.map((n) => (n.id === id ? { ...n, ...updates } : n)),
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
  
  setSolutionZip: (blob) => set({ solutionZip: blob }),
  
  setZipUrl: (url) => set({ zipUrl: url }),
  
  setConnectionStatus: (connected) => set({ isConnected: connected }),
  
  setTokensUsed: (tokens) => set({ tokensUsed: tokens }),

  setStartTime: (time) => set({ startTime: time }),

  getWorkflow: () => get().workflow,

  setWorkflow: (workflow) => set({ workflow }),

  resetTask: () => set((state) => {
    // Keep user-created nodes (not auto-generated ones like boss-1, manager-*, worker-*)
    const userNodes = state.nodes.filter(node =>
      !node.id.startsWith('boss-') &&
      !node.id.startsWith('manager-') &&
      !node.id.startsWith('worker-') &&
      node.id !== 'zip-archive'
    );

    // Keep edges that connect user nodes
    const userEdges = state.edges.filter(edge =>
      userNodes.some(node => node.id === edge.source || node.id === edge.target)
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
      tokensUsed: 0,
      startTime: null,
    };
  }),
}));

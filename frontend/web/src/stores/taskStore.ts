import { create } from 'zustand';

export type AgentNodeType = 'boss' | 'manager' | 'worker' | 'github';
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
  // GitHub specific fields
  repoUrl?: string;
  commitCount?: number;
  // N8n automation
  n8nTrigger?: 'start' | 'end' | 'middle' | 'custom';
  n8nPercentage?: number;
  n8nWorkflowId?: string;
  n8nWebhookUrl?: string;
  // Custom prompt
  customPrompt?: string;
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
      tokensUsed: 0,
      startTime: null,
    };
  }),
}));

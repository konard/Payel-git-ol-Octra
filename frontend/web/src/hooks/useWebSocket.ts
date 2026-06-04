import { useEffect, useRef, useCallback } from 'react';
import { useTaskStore } from '../stores/taskStore';
import { parsePullRequestInfo } from '../lib/pullRequest';
import { t } from './useI18n';

interface WebSocketMessage {
  type: 'connected' | 'reconnected' | 'progress' | 'processing' | 'success' | 'error' | 'chat' | 'clarification_request' | 'pong';
  progress?: number;
  message?: string;
  data?: Record<string, any>;
  task_id?: string;
  status?: string;
  question?: string;
  is_clarification?: boolean;
  is_history?: boolean;
  sender?: 'boss' | 'user';
}

interface WorkflowConfig {
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

interface CreateTaskPayload {
  username: string;
  title: string;
  description: string;
  tokens: Record<string, string>;
  meta: {
    provider: string;
    model: string;
  };
  workflow?: WorkflowConfig;
}

interface StreamedCodeFile {
  path: string;
  content: string;
  language?: string;
  encoding?: string;
  worker_role?: string;
  workerRole?: string;
  manager_role?: string;
  managerRole?: string;
  status?: 'streaming' | 'ready';
  updated_at?: number;
  updatedAt?: number;
}

const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_DELAY = 15000;
const HEARTBEAT_INTERVAL = 30000; // 30 seconds

const isActiveTaskStatus = (status?: string | null) =>
  !!status &&
  ['creating', 'planning', 'executing', 'boss_planning', 'managers_assigned'].includes(status);

function parseCodeFiles(data?: Record<string, any>): StreamedCodeFile[] {
  const raw = data?.code_files;
  if (!raw) return [];

  try {
    if (typeof raw === 'string') {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    }
    if (Array.isArray(raw)) {
      return raw;
    }
  } catch (err) {
    console.warn('[WS] Failed to parse code_files payload:', err);
  }

  return [];
}

function normalizeTaskType(value: unknown): string {
  return typeof value === 'string' ? value.trim().toLowerCase() : '';
}

function normalizeRoleKey(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9а-яё]+/gi, '');
}

function solutionMarkdownFile(data: Record<string, any> | undefined, currentTaskType: string): StreamedCodeFile | null {
  if (!data) return null;

  const taskType = normalizeTaskType(data.taskType || data.task_type) || currentTaskType;
  const content = typeof data.solutionMarkdown === 'string'
    ? data.solutionMarkdown
    : typeof data.solution_markdown === 'string'
      ? data.solution_markdown
      : taskType && taskType !== 'code' && typeof data.response === 'string'
        ? data.response
        : '';

  if (!content.trim()) return null;

  const safeTaskType = taskType && taskType !== 'code' ? taskType : 'result';
  return {
    path: `solution/final-${safeTaskType}.md`,
    content,
    language: 'markdown',
    status: 'ready',
    updated_at: Date.now(),
  };
}

export function useWebSocket(url: string, onChatMessage?: (message: string, sender: 'boss' | 'user', isClarification?: boolean, progress?: number, showProgress?: boolean) => void, onProgressUpdate?: (progress: number, message?: string) => void) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const heartbeatTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttempts = useRef(0);
  const isMounted = useRef(false);
  const isManuallyClosed = useRef(false);
  const lastTaskPayload = useRef<CreateTaskPayload | null>(null);
  const activeTaskId = useRef<string | null>(null);
  const taskStartTime = useRef<number>(0);
  const hasApiConfigError = useRef(false);

  // Track managers across messages
  const managerCounter = useRef(0);
  const managersKnown = useRef<Set<string>>(new Set());
  const zipNodeAdded = useRef(false);
  const workerCounters = useRef<Record<string, number>>({});
  const workersByManager = useRef<Record<string, string[]>>({});
  const workerIdsByManager = useRef<Record<string, string[]>>({});
  const currentManagerRole = useRef<string>('');
  const currentTaskType = useRef<string>('');

  const storeActions = {
    setConnectionStatus: useTaskStore((state) => state.setConnectionStatus),
    setTaskStatus: useTaskStore((state) => state.setTaskStatus),
    setTaskId: useTaskStore((state) => state.setTaskId),
    addNode: useTaskStore((state) => state.addNode),
    updateNode: useTaskStore((state) => state.updateNode),
    addEdge: useTaskStore((state) => state.addEdge),
    addLog: useTaskStore((state) => state.addLog),
    setZipUrl: useTaskStore((state) => state.setZipUrl),
    setTokensUsed: useTaskStore((state) => state.setTokensUsed),
    upsertCodeFiles: useTaskStore((state) => state.upsertCodeFiles),
    completeCodeStreaming: useTaskStore((state) => state.completeCodeStreaming),
    clearCodeFiles: useTaskStore((state) => state.clearCodeFiles),
    recordSearchStep: useTaskStore((state) => state.recordSearchStep),
    setPullRequest: useTaskStore((state) => state.setPullRequest),
    nodes: () => useTaskStore.getState().nodes,
  };

  const isGeneratedNode = (node: { id: string }) =>
    node.id.startsWith('boss-') ||
    node.id.startsWith('manager-') ||
    node.id.startsWith('worker-') ||
    node.id === 'github-archive' ||
    node.id === 'zip-archive';

  const hasUserNodes = () => {
    const nodes = storeActions.nodes();
    return nodes.some(node => !isGeneratedNode(node));
  };

  const findReusableWorkerNode = (role: string) => {
    const roleKey = normalizeRoleKey(role);
    if (!roleKey) return null;
    return storeActions.nodes().find((node) =>
      node.type === 'worker' &&
      !isGeneratedNode(node) &&
      normalizeRoleKey(node.role) === roleKey,
    ) || null;
  };

  const clearReconnectTimer = () => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  };

  const clearHeartbeatTimer = () => {
    if (heartbeatTimerRef.current) {
      clearTimeout(heartbeatTimerRef.current);
      heartbeatTimerRef.current = null;
    }
  };

  const startHeartbeat = () => {
    clearHeartbeatTimer();
    heartbeatTimerRef.current = setTimeout(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        // Send ping frame
        wsRef.current.send(JSON.stringify({ type: 'ping' }));
        startHeartbeat(); // Schedule next heartbeat
      }
    }, HEARTBEAT_INTERVAL);
  };

  const waitForOpen = useCallback(() => {
    return new Promise<void>((resolve, reject) => {
      const ws = wsRef.current;
      if (!ws || ws.readyState === WebSocket.CLOSED) {
        reject(new Error('WebSocket closed'));
        return;
      }

      if (ws.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      const onOpen = () => {
        cleanup();
        resolve();
      };

      const onClose = () => {
        cleanup();
        reject(new Error('WebSocket closed before opening'));
      };

      const onError = () => {
        cleanup();
        reject(new Error('WebSocket connection failed'));
      };

      const timeout = setTimeout(() => {
        cleanup();
        reject(new Error('WebSocket connection timeout'));
      }, 10000);

      const cleanup = () => {
        clearTimeout(timeout);
        ws.removeEventListener('open', onOpen);
        ws.removeEventListener('close', onClose);
        ws.removeEventListener('error', onError);
      };

      ws.addEventListener('open', onOpen);
      ws.addEventListener('close', onClose);
      ws.addEventListener('error', onError);
    });
  }, []);

  const handleWebSocketMessage = (msg: WebSocketMessage) => {
    switch (msg.type) {
      case 'connected':
        storeActions.addLog({
          message: msg.message || 'Connected to task creation service',
          type: 'info',
        });
        // Save the task_id for reconnection
        if (msg.task_id) {
          activeTaskId.current = msg.task_id;
          taskStartTime.current = Date.now();
          console.log('[WS] Task ID saved:', msg.task_id);
        }
        break;

      case 'reconnected':
        storeActions.addLog({
          message: msg.message || `Reconnected (progress: ${msg.progress || 0}%)`,
          type: 'info',
        });
        // Restore task status from backend
        if (msg.status) {
          storeActions.setTaskStatus(msg.status as any);
          // If task is completed, clear active task ID to prevent auto-reconnect
          if (msg.status === 'done' || msg.status === 'error' || msg.status === 'cancelled') {
            activeTaskId.current = null;
          }
        }
        if (msg.task_id) {
          storeActions.setTaskId(msg.task_id);
        }
        break;

      case 'progress':
      case 'processing':
        handleCodeFilesMessage(msg);
        // Skip historical messages to avoid reprocessing
        if (!msg.is_history) {
          handleProgressMessage(msg);
        }
        break;

      case 'success':
        handleCodeFilesMessage(msg);
        storeActions.setTaskStatus('done');
        storeActions.completeCodeStreaming();
        storeActions.addLog({
          message: msg.message || 'Project ready!',
          type: 'success',
        });
        // Non-code tasks (research/document/presentation): Boss posts a short
        // text answer in chat; the full result lives in the Solution tab.
        if (msg.data?.chatSummary && onChatMessage) {
          onChatMessage(msg.data.chatSummary, 'boss', false);
        }
        if (msg.data?.repoUrl) {
          storeActions.setZipUrl(msg.data.repoUrl);
          addGitHubNode(msg.data.repoUrl);
        } else if (msg.data?.zipUrl) {
          storeActions.setZipUrl(msg.data.zipUrl);
        }
        // Pull request summary for the Solution pane so the user does not need to
        // jump back to GitHub (issue #44).
        {
          const pullRequest = parsePullRequestInfo(msg.data);
          if (pullRequest) {
            storeActions.setPullRequest(pullRequest);
          }
        }
        // Update all nodes to done
        finalizeAllNodes('done');
        // Clear stored task payload — task is complete, no need to resend
        lastTaskPayload.current = null;
        activeTaskId.current = null;
        taskStartTime.current = 0;

        // Update ZIP node with file info if available
        if (msg.data?.filesCount || msg.data?.zipSize) {
          storeActions.updateNode('zip-archive', {
            status: 'done',
            filesCount: msg.data.filesCount,
            fileSize: msg.data.zipSize ? `${(msg.data.zipSize / 1024).toFixed(1)} KB` : '',
          });
        }
        break;

      case 'error':
        const errorMessage = msg.message || 'Error occurred';
        storeActions.setTaskStatus('error');
        storeActions.addLog({
          message: errorMessage,
          type: 'error',
        });
        // Update all nodes to error
        finalizeAllNodes('error');

        // Check if this is an API configuration error
        const isApiError = errorMessage.includes('API key') ||
                          errorMessage.includes('not found in tokens') ||
                          errorMessage.includes('Payment Required') ||
                          errorMessage.includes('credits');

        if (isApiError) {
          hasApiConfigError.current = true;
          storeActions.addLog({
            message: 'API configuration error - please check your API keys',
            type: 'error',
          });
        }

        // Clear stored task payload — task failed, don't auto-resend
        lastTaskPayload.current = null;
        activeTaskId.current = null;
        taskStartTime.current = 0;

        // Reset connection state to allow new tasks
        isManuallyClosed.current = false;
        reconnectAttempts.current = 0;
        break;

      case 'chat':
        if (onChatMessage && msg.message && msg.sender) {
          onChatMessage(msg.message, msg.sender as 'boss' | 'user', msg.is_clarification);
        }
        break;

      case 'clarification_request':
        if (onChatMessage && msg.question) {
          onChatMessage(msg.question, 'boss', true);
        }
        break;
    }
  };

  const handleCodeFilesMessage = (msg: WebSocketMessage) => {
    const codeFiles = parseCodeFiles(msg.data);
    const finalSolution = solutionMarkdownFile(msg.data, currentTaskType.current);
    if (finalSolution) {
      codeFiles.push(finalSolution);
    }
    if (codeFiles.length === 0) return;

    storeActions.upsertCodeFiles(codeFiles
      .filter((file) => file.path && typeof file.content === 'string')
      .map((file) => ({
        path: file.path,
        name: file.path.split('/').filter(Boolean).at(-1) || file.path,
        language: file.language || 'plaintext',
        encoding: file.encoding,
        content: file.content,
        status: file.status || 'streaming',
        workerRole: file.workerRole || file.worker_role,
        managerRole: file.managerRole || file.manager_role,
        updatedAt: file.updatedAt || file.updated_at || Date.now(),
      })));
  };

  // Helper: capitalize first letter
  const capitalizeRole = (role: string): string => {
    return role
      .split(/[_\s]+/)
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ');
  };

  // Helper: finalize all nodes (done/error)
  const finalizeAllNodes = (status: 'done' | 'error') => {
    const nodes = storeActions.nodes();
    nodes.forEach((node) => {
      storeActions.updateNode(node.id, { status });
    });
  };

  // Helper: add the publishing sink during code packaging, then attach the
  // repository URL when the backend returns one.
  const addGitHubNode = (repoUrl?: string) => {
    const nodeStatus: 'working' | 'done' = repoUrl ? 'done' : 'working';
    if (zipNodeAdded.current) {
      storeActions.updateNode('github-archive', {
        ...(repoUrl ? { repoUrl } : {}),
        status: nodeStatus,
      });
      return;
    }
    zipNodeAdded.current = true;

    // Find all worker nodes
    const nodes = storeActions.nodes();
    const workerNodes = nodes.filter(n => n.type === 'worker');
    const managerNodes = nodes.filter(n => n.type === 'manager');

    // Calculate center position
    const allX = [...workerNodes, ...managerNodes].map(n => n.position?.x || 0);
    const centerX = allX.length > 0 ? Math.min(...allX) + (Math.max(...allX) - Math.min(...allX)) / 2 : 400;

    storeActions.addNode({
      id: 'github-archive',
      type: 'github',
      role: 'GitHub',
      status: nodeStatus,
      ...(repoUrl ? { repoUrl } : {}),
      position: { x: centerX, y: 520 },
    });

    // Add edges from all workers to ZIP (or managers if no workers)
    const edgeSources = workerNodes.length > 0 ? workerNodes : managerNodes;
    edgeSources.forEach((node) => {
      storeActions.addEdge({ from: node.id, to: 'github-archive' });
    });

    console.log('[WS] => GitHub node added with', edgeSources.length, 'edges');
  };



  const handleProgressMessage = (msg: WebSocketMessage) => {
    const progress = msg.progress || 0;
    const message = msg.message || '';

    console.log('[WS] handleProgressMessage:', message, 'progress:', progress);

    const taskType = normalizeTaskType(msg.data?.taskType || msg.data?.task_type);
    if (taskType) {
      currentTaskType.current = taskType;
    }

    storeActions.addLog({
      message: `${message} (${progress}%)`,
      type: 'info',
    });

    // === WEB SEARCH STEPS — research workers stream their search progress ===
    // The worker sends data.search_phase ('searching' | 'done') with an optional
    // data.search_step line and a final data.search_steps_count. These drive the
    // collapsible "Searching the web" panel in the chat.
    const searchPhase = msg.data?.search_phase;
    if (searchPhase === 'searching' || searchPhase === 'done') {
      const step = typeof msg.data?.search_step === 'string' ? msg.data.search_step : '';
      const count = Number.parseInt(String(msg.data?.search_steps_count ?? ''), 10);
      storeActions.recordSearchStep(step, searchPhase, Number.isNaN(count) ? 0 : count);
    }

    // Update progress in chat if this is a boss progress message
    if (onProgressUpdate && (message.includes('AI is planning') ||
                            message.includes('Creating') ||
                            message.includes('completed') ||
                            message.includes('packaging') ||
                            message.includes('Boss validating'))) {
      onProgressUpdate(progress, message);
    }

    // === BOSS ===
    if (message.includes('Architecture planned')) {
      console.log('[WS] => Boss detected, adding node');
      storeActions.setTaskStatus('planning');
      managerCounter.current = 0;
      managersKnown.current.clear();
      zipNodeAdded.current = false;
      workerCounters.current = {};
      workersByManager.current = {};
      workerIdsByManager.current = {};
      currentTaskType.current = taskType || currentTaskType.current;

      // Only add auto-generated boss if no user nodes exist
      if (!hasUserNodes()) {
        const rawTechStack = msg.data?.techStack;
        const techStack: string[] = Array.isArray(rawTechStack)
          ? rawTechStack
          : typeof rawTechStack === 'string' && rawTechStack !== 'null' && rawTechStack.trim()
            ? rawTechStack.split(',').map((s: string) => s.trim()).filter(Boolean)
            : [];

        storeActions.addNode({
          id: 'boss-1',
          type: 'boss',
          role: 'CEO',
          status: 'working',
          techStack,
          position: { x: 400, y: 50 },
        });
      }
    }

    // === MANAGERS — "Starting manager: Project Lead / Architect" ===
    const startingMatch = message.match(/Starting manager:\s*(.+)/i);
    if (startingMatch) {
      const role = startingMatch[1].trim();
      if (!managersKnown.current.has(role)) {
        managersKnown.current.add(role);
        const idx = managerCounter.current++;
        const managerId = `manager-${idx}`;
        console.log('[WS] => Adding manager:', role, 'at index', idx);
        storeActions.setTaskStatus('executing');

        // Only add auto-generated manager if no user nodes exist
        if (!hasUserNodes()) {
          storeActions.addNode({
            id: managerId,
            type: 'manager',
            role: capitalizeRole(role),
            status: 'working',
            position: {
              x: 200 + (idx * 250),
              y: 250,
            },
          });
          storeActions.addEdge({ from: 'boss-1', to: managerId });
        }
      } else {
        // Already added — just update status
        const idx = [...managersKnown.current].indexOf(role);
        if (idx >= 0) {
          storeActions.updateNode(`manager-${idx}`, { status: 'working' });
        }
      }
    }

    // === MANAGER WORKING — "Manager working: Project Lead / Architect" ===
    const workingMatch = message.match(/Manager working:\s*(.+)/i);
    if (workingMatch) {
      if (msg.data?.current_role) {
        // Use current_role from message data if available
        currentManagerRole.current = msg.data.current_role;
      } else {
        currentManagerRole.current = workingMatch[1].trim();
      }
      const role = currentManagerRole.current;
      console.log('[WS] => Manager working:', role, 'setting currentManagerRole');
      const idx = [...managersKnown.current].indexOf(role);
      if (idx >= 0) {
        storeActions.updateNode(`manager-${idx}`, { status: 'working' });
      }
    }

    // === MANAGER COMPLETED — "Manager completed: Project Lead / Architect" ===
    const completedMatch = message.match(/Manager completed:\s*(.+)/i);
    if (completedMatch) {
      const role = completedMatch[1].trim();
      const idx = [...managersKnown.current].indexOf(role);
      if (idx >= 0) {
        storeActions.updateNode(`manager-${idx}`, { status: 'done' });
      }
    }

    // === WORKER STARTING — "Starting worker 1/7: Go Backend Developer" ===
    const workerStartMatch = message.match(/Starting worker \d+\/\d+:\s*(.+)$/i);
    console.log('[WS] => Worker start regex check:', message, 'match:', !!workerStartMatch);
    if (workerStartMatch) {
      const role = workerStartMatch[1].trim();
      // Use current_role from message data if available (more reliable than currentManagerRole)
      const managerRole = msg.data?.current_role || currentManagerRole.current;
      const managerIdx = [...managersKnown.current].indexOf(managerRole);

      // Initialize worker tracking for this manager
      if (!workersByManager.current[managerRole]) {
        workersByManager.current[managerRole] = [];
      }
      if (!workerIdsByManager.current[managerRole]) {
        workerIdsByManager.current[managerRole] = [];
      }
      if (!workerCounters.current[managerRole]) {
        workerCounters.current[managerRole] = 0;
      }

      // Check if worker already exists (by role uniqueness)
      if (!workersByManager.current[managerRole].includes(role)) {
        const wIdx = workerCounters.current[managerRole]++;
        workersByManager.current[managerRole].push(role);
        const reusableWorker = findReusableWorkerNode(role);
        const workerId = reusableWorker?.id || `worker-${managerRole.replace(/[^a-zA-Z0-9]/g, '')}-${wIdx}`;
        workerIdsByManager.current[managerRole].push(workerId);

        if (reusableWorker) {
          console.log('[WS] => Reusing custom worker:', workerId);
          storeActions.updateNode(workerId, { status: 'working' });
        } else {
          // Add generated workers when the user did not already place a matching one.
          const mgrPosition = managerIdx >= 0 ?
            (storeActions.nodes().find(n => n.id === `manager-${managerIdx}`)?.position?.x || 200) : 200;
          const workerX = mgrPosition + (wIdx - 1) * 120;
          const workerY = 370;

          console.log('[WS] => Adding worker:', workerId, 'at position:', workerX, workerY);
          storeActions.addNode({
            id: workerId,
            type: 'worker',
            role: capitalizeRole(role),
            status: 'working',
            position: { x: workerX, y: workerY },
          });

          // Edge from manager to worker
          if (managerIdx >= 0) {
            storeActions.addEdge({ from: `manager-${managerIdx}`, to: workerId });
          }
        }
      }
    }

    // === WORKER COMPLETED — "Worker 1/7 (Go Backend Developer) completed: 5 files (65%)" ===
    const workerCompleteMatch = message.match(/Worker \d+\/\d+ \((.+?)\) completed:\s*(\d+)\s*files/i);
    if (workerCompleteMatch) {
      const role = workerCompleteMatch[1].trim();
      const filesCount = parseInt(workerCompleteMatch[2], 10);
      // Use current_role from message data if available
      const managerRole = msg.data?.current_role || currentManagerRole.current;

      // Find worker by role and manager
      if (workersByManager.current[managerRole]) {
        const wIdx = workersByManager.current[managerRole].indexOf(role);
        if (wIdx >= 0) {
          const sanitizedManagerRole = managerRole.replace(/[^a-zA-Z0-9]/g, '');
          const workerId = workerIdsByManager.current[managerRole]?.[wIdx] || `worker-${sanitizedManagerRole}-${wIdx}`;

          console.log('[WS] => Worker completed:', workerId, 'files:', filesCount);
          storeActions.updateNode(workerId, {
            status: 'done',
            filesCount,
          });
        }
      }
    }

    // === ALL MANAGERS COMPLETED ===
    if (message.includes('All managers completed')) {
      // Update all managers to done
      managersKnown.current.forEach((role) => {
        const idx = [...managersKnown.current].indexOf(role);
        if (idx >= 0) {
          storeActions.updateNode(`manager-${idx}`, { status: 'done' });
        }
      });

    }

    // === BOSS VALIDATING ===
    if (message.includes('Boss validating')) {
      storeActions.updateNode('boss-1', { status: 'reviewing' });

    }

    // === PACKAGING ===
    if (message.includes('Packaging')) {
      storeActions.updateNode('boss-1', { status: 'done' });

      if (currentTaskType.current === '' || currentTaskType.current === 'code') {
        addGitHubNode();
      }
    }

    // === PROJECT READY (progress 100) ===
    if (progress === 100 && msg.data?.repoUrl) {
      storeActions.setZipUrl(msg.data.repoUrl);
      addGitHubNode(msg.data.repoUrl);
    }
  };

  const connect = useCallback(() => {
    if (wsRef.current) {
      const readyState = wsRef.current.readyState;
      if (readyState === WebSocket.OPEN || readyState === WebSocket.CONNECTING) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
      wsRef.current = null;
    }

    // On reconnect, use /api/task/reconnect with task_id instead of creating a new task
    // Only reconnect if we have an active task that's still running and not too old
    const currentStatus = useTaskStore.getState().status;
    const taskAge = Date.now() - taskStartTime.current;
    const maxTaskAge = 2 * 60 * 60 * 1000; // 2 hours
    const isReconnect = activeTaskId.current !== null &&
      isActiveTaskStatus(currentStatus) &&
      taskAge < maxTaskAge &&
      !hasApiConfigError.current; // Don't reconnect if we had an API config error
    const connectUrl = isReconnect
      ? url.replace('/task/create', '/task/reconnect')
      : url;

    console.log('[WS] 🔌 Connecting to', connectUrl, isReconnect ? '(reconnect)' : '(new)');
    const ws = new WebSocket(connectUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('[WS] Connected to', connectUrl);
      reconnectAttempts.current = 0;
      storeActions.setConnectionStatus(true);
      startHeartbeat(); // Start heartbeat when connected

      // On reconnect, send task_id to restore state from Redis
      if (isReconnect && activeTaskId.current) {
        console.log('[WS] Reconnecting to task:', activeTaskId.current);
        ws.send(JSON.stringify({ task_id: activeTaskId.current }));
        storeActions.addLog({
          message: t('console.reconnecting'),
          type: 'info',
        });
      }
    };

    ws.onclose = (event) => {
      console.log('[WS] Closed, code:', event.code, 'reason:', event.reason);
      wsRef.current = null;
      storeActions.setConnectionStatus(false);

      const currentStatus = useTaskStore.getState().status;
      const shouldReconnect =
        event.code !== 1000 &&
        !isManuallyClosed.current &&
        isMounted.current &&
        activeTaskId.current !== null &&
        isActiveTaskStatus(currentStatus) &&
        !hasApiConfigError.current;

      if (!shouldReconnect) {
        return;
      }

      storeActions.addLog({
        message: t('console.connectionLost'),
        type: 'warning',
      });

      clearReconnectTimer();
      const delay = Math.min(RECONNECT_INTERVAL * Math.pow(1.5, reconnectAttempts.current), MAX_RECONNECT_DELAY);
      reconnectAttempts.current++;
      reconnectTimerRef.current = setTimeout(() => {
        connect();
      }, delay);
    };

    ws.onerror = (event) => {
      console.error('[WS] WebSocket error:', event);
      storeActions.setConnectionStatus(false);
    };

    ws.onmessage = (event) => {
      console.log('[WS] 📨 Raw message:', event.data);
      try {
        const msg: WebSocketMessage = JSON.parse(event.data);
        console.log('[WS] Parsed message type:', msg.type, 'message:', msg.message);

        // Handle pong responses
        if (msg.type === 'pong') {
          console.log('[WS] ❤️ Pong received');
          return;
        }

        handleWebSocketMessage(msg);
      } catch (err) {
        console.error('[WS] Failed to parse message:', err, 'Raw data:', event.data);
      }
    };
  }, [url, storeActions]);

  const send = useCallback((data: CreateTaskPayload) => {
    // Clear any stale task identifiers to prevent blocking new sends
    activeTaskId.current = null;
    taskStartTime.current = 0;
    hasApiConfigError.current = false;
    currentTaskType.current = '';
    zipNodeAdded.current = false;
    workerCounters.current = {};
    workersByManager.current = {};
    workerIdsByManager.current = {};

    const currentStatus = useTaskStore.getState().status;
    // If status is stuck at a non-terminal state (e.g. from a previous failed
    // send that set status to 'creating' before the error), force it back to
    // idle so the user can retry.
    if (currentStatus !== 'idle') {
      if (!['done', 'error', 'cancelled'].includes(currentStatus)) {
        storeActions.setTaskStatus('idle');
      }
    }

    if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED || wsRef.current.readyState === WebSocket.CLOSING) {
      storeActions.addLog({
        message: t('bottomInput.reconnecting'),
        type: 'warning',
      });
      isManuallyClosed.current = false;
      connect();
    }

    waitForOpen().then(() => {
      const freshStatus = useTaskStore.getState().status;
      const isReconnect = activeTaskId.current !== null && isActiveTaskStatus(freshStatus);

      if (!isReconnect) {
        // Reset API config error flag for new tasks
        hasApiConfigError.current = false;
        // Store the payload for potential reconnection
        lastTaskPayload.current = data;
        wsRef.current!.send(JSON.stringify(data));
        storeActions.clearCodeFiles();
        storeActions.setTaskStatus('creating');
        storeActions.setTaskId(`task-${Date.now()}`);
        storeActions.addLog({
          message: `${t('bottomInput.taskCreated')}: ${data.title}`,
          type: 'success',
        });
      }
    }).catch((err) => {
      console.error('[WS] Send error:', err);
      storeActions.addLog({
        message: `${t('bottomInput.sendError')}: ${err.message}`,
        type: 'error',
      });
    });
  }, [connect, storeActions, waitForOpen]);

  const disconnect = useCallback(() => {
    isManuallyClosed.current = true;
    clearReconnectTimer();
    clearHeartbeatTimer();
    lastTaskPayload.current = null;
    activeTaskId.current = null;
    taskStartTime.current = 0;
    if (wsRef.current) {
      wsRef.current.onclose = null;
      wsRef.current.close(1000);
      wsRef.current = null;
    }
    storeActions.setConnectionStatus(false);
  }, [storeActions]);

  useEffect(() => {
    isMounted.current = true;
    isManuallyClosed.current = false;

    return () => {
      isMounted.current = false;
      clearReconnectTimer();
      clearHeartbeatTimer();
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  const sendChat = useCallback(async (
    message: string,
    sender: 'boss' | 'user' = 'user',
    taskPayload?: CreateTaskPayload | null,
  ): Promise<boolean> => {
    if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED || wsRef.current.readyState === WebSocket.CLOSING) {
      isManuallyClosed.current = false;
      connect();
    }

    try {
      await waitForOpen();
      wsRef.current!.send(JSON.stringify({
        type: 'chat',
        message,
        sender,
        taskPayload,
      }));
      return true;
    } catch (err) {
      console.error('[WS] Cannot send chat message:', err);
      return false;
    }
  }, [connect, waitForOpen]);

  return { connect, send, sendChat, disconnect };
}

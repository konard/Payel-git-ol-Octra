import { useEffect, useRef, useCallback } from 'react';
import { useTaskStore } from '../stores/taskStore';
import type { AgentNode } from '../stores/taskStore';
import { t } from './useI18n';

interface WebSocketMessage {
  type: 'connected' | 'reconnected' | 'progress' | 'processing' | 'success' | 'error' | 'chat';
  progress?: number;
  message?: string;
  data?: Record<string, any>;
  task_id?: string;
  status?: string;
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

const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_DELAY = 15000;
const HEARTBEAT_INTERVAL = 30000; // 30 seconds

export function useWebSocket(url: string, onChatMessage?: (message: string, sender: 'boss' | 'user') => void) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const heartbeatTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttempts = useRef(0);
  const isMounted = useRef(false);
  const isManuallyClosed = useRef(false);
  const lastTaskPayload = useRef<CreateTaskPayload | null>(null);
  const activeTaskId = useRef<string | null>(null);
  const taskStartTime = useRef<number>(0);

  // Track managers across messages
  const managerCounter = useRef(0);
  const managersKnown = useRef<Set<string>>(new Set());
  const zipNodeAdded = useRef(false);
  const workerCounters = useRef<Record<string, number>>({});
  const workersByManager = useRef<Record<string, string[]>>({});

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
    nodes: () => useTaskStore.getState().nodes,
  };

  // Check if there are user-created nodes (not auto-generated)
  const hasUserNodes = () => {
    const nodes = storeActions.nodes();
    return nodes.some(node =>
      !node.id.startsWith('boss-') &&
      !node.id.startsWith('manager-') &&
      !node.id.startsWith('worker-') &&
      node.id !== 'zip-archive'
    );
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
        // Skip historical messages to avoid reprocessing
        if (!msg.is_history) {
          handleProgressMessage(msg);
        }
        break;

      case 'success':
        storeActions.setTaskStatus('done');
        storeActions.addLog({
          message: msg.message || 'Project ready!',
          type: 'success',
        });
        if (msg.data?.zipUrl) {
          storeActions.setZipUrl(msg.data.zipUrl);
        }
        // Update all nodes to done including ZIP
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
        storeActions.setTaskStatus('error');
        storeActions.addLog({
          message: msg.message || 'Error occurred',
          type: 'error',
        });
        // Update all nodes to error
        finalizeAllNodes('error');
        // Clear stored task payload — task failed, don't auto-resend
        lastTaskPayload.current = null;
        activeTaskId.current = null;

        // Reset connection state to allow new tasks
        isManuallyClosed.current = false;
        reconnectAttempts.current = 0;
        break;

      case 'chat':
        if (onChatMessage && msg.message && msg.sender) {
          onChatMessage(msg.message, msg.sender);
        }
        break;
    }
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

  // Helper: add ZIP Archive node with edges from all workers
  const addZIPArchiveNode = () => {
    if (zipNodeAdded.current || hasUserNodes()) return;
    zipNodeAdded.current = true;

    // Find all worker nodes
    const nodes = storeActions.nodes();
    const workerNodes = nodes.filter(n => n.type === 'worker');
    const managerNodes = nodes.filter(n => n.type === 'manager');

    // Calculate center position
    const allX = [...workerNodes, ...managerNodes].map(n => n.position?.x || 0);
    const centerX = allX.length > 0 ? Math.min(...allX) + (Math.max(...allX) - Math.min(...allX)) / 2 : 400;

    storeActions.addNode({
      id: 'zip-archive',
      type: 'zip',
      role: 'ZIP Archive',
      status: 'working',
      fileName: 'project.zip',
      position: { x: centerX, y: 520 },
    });

    // Add edges from all workers to ZIP (or managers if no workers)
    const edgeSources = workerNodes.length > 0 ? workerNodes : managerNodes;
    edgeSources.forEach((node) => {
      storeActions.addEdge({ from: node.id, to: 'zip-archive' });
    });

    console.log('[WS] => ZIP Archive node added with', edgeSources.length, 'edges');
  };

  const handleProgressMessage = (msg: WebSocketMessage) => {
    const progress = msg.progress || 0;
    const message = msg.message || '';

    console.log('[WS] handleProgressMessage:', message);

    storeActions.addLog({
      message: `${message} (${progress}%)`,
      type: 'info',
    });

    // === BOSS ===
    if (message.includes('Architecture planned')) {
      console.log('[WS] => Boss detected, adding node');
      storeActions.setTaskStatus('planning');
      managerCounter.current = 0;
      managersKnown.current.clear();
      zipNodeAdded.current = false;
      workerCounters.current = {};
      workersByManager.current = {};

      // Only add auto-generated boss if no user nodes exist
      if (!hasUserNodes()) {
        storeActions.addNode({
          id: 'boss-1',
          type: 'boss',
          role: 'CEO',
          status: 'working',
          techStack: msg.data?.techStack || [],
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
      const role = workingMatch[1].trim();
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

    // === WORKER STARTING — "Starting worker: Network Protocol Engineer" ===
    const workerStartMatch = message.match(/Starting worker:\s*(.+)/i);
    if (workerStartMatch) {
      const role = workerStartMatch[1].trim();
      const managerRole = msg.data?.current_role || '';
      const managerIdx = [...managersKnown.current].indexOf(managerRole);

      // Initialize worker tracking for this manager
      if (!workersByManager.current[managerRole]) {
        workersByManager.current[managerRole] = [];
      }
      if (!workerCounters.current[managerRole]) {
        workerCounters.current[managerRole] = 0;
      }

      // Check if worker already exists (by role uniqueness)
      if (!workersByManager.current[managerRole].includes(role)) {
        const wIdx = workerCounters.current[managerRole]++;
        workersByManager.current[managerRole].push(role);
        const workerId = `worker-${managerRole.replace(/[^a-zA-Z0-9]/g, '')}-${wIdx}`;

        // Only add auto-generated worker if no user nodes exist
        if (!hasUserNodes()) {
          // Position: under the manager
          const mgrPosition = managerIdx >= 0 ?
            (storeActions.nodes().find(n => n.id === `manager-${managerIdx}`)?.position?.x || 200) : 200;
          const workerX = mgrPosition + (wIdx - 1) * 120;
          const workerY = 370;

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

    // === WORKER COMPLETED — "Worker completed: Network Protocol Engineer (5 files)" ===
    const workerCompleteMatch = message.match(/Worker completed:\s*(.+?)(?:\s*\((\d+)\s*files?\))?$/i);
    if (workerCompleteMatch) {
      const role = workerCompleteMatch[1].trim();
      const filesCount = workerCompleteMatch[2] ? parseInt(workerCompleteMatch[2], 10) : undefined;
      const managerRole = msg.data?.current_role || '';

      // Find worker by role and manager
      if (workersByManager.current[managerRole]) {
        const wIdx = workersByManager.current[managerRole].indexOf(role);
        if (wIdx >= 0) {
          const sanitizedManagerRole = managerRole.replace(/[^a-zA-Z0-9]/g, '');
          const workerId = `worker-${sanitizedManagerRole}-${wIdx}`;

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

      // Add ZIP Archive node
      addZIPArchiveNode();
    }

    // === BOSS VALIDATING ===
    if (message.includes('Boss validating')) {
      storeActions.updateNode('boss-1', { status: 'reviewing' });

      // Ensure ZIP node exists (if not added by "All managers completed")
      addZIPArchiveNode();
    }

    // === PACKAGING ===
    if (message.includes('Packaging')) {
      storeActions.updateNode('boss-1', { status: 'done' });

      // Update ZIP node to done
      storeActions.updateNode('zip-archive', { status: 'done' });

      // Ensure ZIP node exists
      addZIPArchiveNode();
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
    const currentStatus = storeActions.nodes ? storeActions.nodes().find(n => n.type === 'boss')?.data?.status : '';
    const taskAge = Date.now() - taskStartTime.current;
    const maxTaskAge = 2 * 60 * 60 * 1000; // 2 hours
    const isReconnect = activeTaskId.current !== null &&
      currentStatus &&
      ['creating', 'planning', 'executing', 'boss_planning', 'managers_assigned'].includes(currentStatus) &&
      taskAge < maxTaskAge;
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

      if (event.code === 1000 || isManuallyClosed.current) {
        // Normal close — no user-facing log needed
      } else {
        storeActions.addLog({
          message: t('console.connectionLost'),
          type: 'warning',
        });
        if (!isMounted.current || isManuallyClosed.current) return;
        clearReconnectTimer();
        const delay = Math.min(RECONNECT_INTERVAL * Math.pow(1.5, reconnectAttempts.current), MAX_RECONNECT_DELAY);
        reconnectAttempts.current++;
        reconnectTimerRef.current = setTimeout(() => {
          connect();
        }, delay);
      }
    };

    ws.onerror = () => {
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
    // Don't send if we have an active task that's not completed
    const currentStatus = storeActions.nodes ? storeActions.nodes().find(n => n.type === 'boss')?.data?.status : '';
    if (currentStatus && !['done', 'error', 'cancelled'].includes(currentStatus) && activeTaskId.current) {
      storeActions.addLog({
        message: 'Task is already running, cannot create new task',
        type: 'warning',
      });
      return;
    }

    if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED || wsRef.current.readyState === WebSocket.CLOSING) {
      storeActions.addLog({
        message: t('bottomInput.reconnecting'),
        type: 'warning',
      });
      isManuallyClosed.current = false;
      connect();
    }

    const waitForOpen = () => {
      return new Promise<void>((resolve, reject) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          resolve();
          return;
        }

        if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED) {
          reject(new Error('WebSocket closed'));
          return;
        }

        const timeout = setTimeout(() => {
          reject(new Error('WebSocket connection timeout'));
        }, 10000);

        const onOpen = () => {
          clearTimeout(timeout);
          wsRef.current?.removeEventListener('open', onOpen);
          resolve();
        };

        wsRef.current?.addEventListener('open', onOpen);
      });
    };

    waitForOpen().then(() => {
      // Only send data if this is a new task creation, not a reconnect
      const isReconnect = activeTaskId.current !== null &&
        currentStatus &&
        ['creating', 'planning', 'executing', 'boss_planning', 'managers_assigned'].includes(currentStatus);

      if (!isReconnect) {
        // Store the payload for potential reconnection
        lastTaskPayload.current = data;
        wsRef.current!.send(JSON.stringify(data));
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
  }, [connect, storeActions]);

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

  const sendChat = useCallback((message: string, sender: 'boss' | 'user' = 'user') => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.error('[WS] Cannot send chat message: WebSocket not connected');
      return;
    }

    wsRef.current.send(JSON.stringify({
      type: 'chat',
      message,
      sender,
    }));
  }, []);

  return { connect, send, sendChat, disconnect };
}

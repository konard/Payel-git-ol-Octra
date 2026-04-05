import { useEffect, useRef, useCallback } from 'react';
import { useTaskStore } from '../stores/taskStore';
import type { AgentNode } from '../stores/taskStore';

interface WebSocketMessage {
  type: 'connected' | 'progress' | 'processing' | 'success' | 'error';
  progress?: number;
  message?: string;
  data?: Record<string, any>;
  task_id?: string;
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
}

const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_DELAY = 15000;

export function useWebSocket(url: string) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttempts = useRef(0);
  const isMounted = useRef(false);
  const isManuallyClosed = useRef(false);

  // Track managers across messages
  const managerCounter = useRef(0);
  const managersKnown = useRef<Set<string>>(new Set());

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
    nodes: useTaskStore((state) => state.nodes),
  };

  const clearReconnectTimer = () => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  };

  const handleWebSocketMessage = (msg: WebSocketMessage) => {
    switch (msg.type) {
      case 'connected':
        storeActions.addLog({
          message: msg.message || 'Connected to task creation service',
          type: 'info',
        });
        break;

      case 'progress':
      case 'processing':
        handleProgressMessage(msg);
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
        break;

      case 'error':
        storeActions.setTaskStatus('error');
        storeActions.addLog({
          message: msg.message || 'Error occurred',
          type: 'error',
        });
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
      storeActions.addNode({
        id: 'boss-1',
        type: 'boss',
        role: 'CEO',
        status: 'working',
        techStack: msg.data?.techStack || [],
        position: { x: 400, y: 50 },
      });
    }

    // === MANAGERS — "Starting manager: backend" ===
    const startingMatch = message.match(/Starting manager:\s*(\S+)/i);
    if (startingMatch) {
      const role = startingMatch[1];
      if (!managersKnown.current.has(role)) {
        managersKnown.current.add(role);
        const idx = managerCounter.current++;
        const managerId = `manager-${idx}`;
        console.log('[WS] => Adding manager:', role, 'at index', idx);
        storeActions.setTaskStatus('executing');
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
      } else {
        // Already added — just update status
        const idx = [...managersKnown.current].indexOf(role);
        if (idx >= 0) {
          storeActions.updateNode(`manager-${idx}`, { status: 'working' });
        }
      }
    }

    // === MANAGER WORKING ===
    const workingMatch = message.match(/Manager working:\s*(\S+)/i);
    if (workingMatch) {
      const role = workingMatch[1];
      const idx = [...managersKnown.current].indexOf(role);
      if (idx >= 0) {
        storeActions.updateNode(`manager-${idx}`, { status: 'working' });
      }
    }

    // === MANAGER COMPLETED ===
    const completedMatch = message.match(/Manager completed:\s*(\S+)/i);
    if (completedMatch) {
      const role = completedMatch[1];
      const idx = [...managersKnown.current].indexOf(role);
      if (idx >= 0) {
        storeActions.updateNode(`manager-${idx}`, { status: 'done' });
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

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('[WS] ✅ Connected to', url);
      reconnectAttempts.current = 0;
      storeActions.setConnectionStatus(true);
      storeActions.addLog({
        message: 'WebSocket подключено',
        type: 'success',
      });
    };

    ws.onclose = (event) => {
      console.log('[WS] ❌ Closed, code:', event.code, 'reason:', event.reason);
      wsRef.current = null;
      storeActions.setConnectionStatus(false);

      if (event.code === 1000 || isManuallyClosed.current) {
        storeActions.addLog({
          message: 'WebSocket отключено',
          type: 'info',
        });
      } else {
        storeActions.addLog({
          message: `WebSocket отключено (code: ${event.code}), переподключение...`,
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
        handleWebSocketMessage(msg);
      } catch (err) {
        console.error('[WS] Failed to parse message:', err, 'Raw data:', event.data);
      }
    };
  }, [url, storeActions]);

  const send = useCallback((data: CreateTaskPayload) => {
    if (!wsRef.current || wsRef.current.readyState === WebSocket.CLOSED || wsRef.current.readyState === WebSocket.CLOSING) {
      storeActions.addLog({
        message: 'Переподключение к серверу...',
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
      wsRef.current!.send(JSON.stringify(data));
      storeActions.setTaskStatus('creating');
      storeActions.setTaskId(`task-${Date.now()}`);
      storeActions.addLog({
        message: `Задача создана: ${data.title}`,
        type: 'success',
      });
    }).catch((err) => {
      console.error('[WS] Send error:', err);
      storeActions.addLog({
        message: 'Ошибка отправки задачи: ' + err.message,
        type: 'error',
      });
    });
  }, [connect, storeActions]);

  const disconnect = useCallback(() => {
    isManuallyClosed.current = true;
    clearReconnectTimer();
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
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  return { connect, send, disconnect };
}

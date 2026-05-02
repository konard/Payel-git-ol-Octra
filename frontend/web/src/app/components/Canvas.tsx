// Fixed contextMenu issue - version 2.0
import { useCallback, useMemo, useState, useRef, useEffect } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  useReactFlow,
  addEdge,
  type Node,
  type Edge,
  type Connection,
  type OnNodesChange,
  type OnEdgesChange,
  type ReactFlowInstance,
  type OnConnect,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useTaskStore, type AgentNodeType, type AgentNode } from '../../stores/taskStore';
import { BossNode, ManagerNode, WorkerNode, GitHubNode } from './nodes';
import { NodeSidebar } from './NodeSidebar';
import { NodeContextMenu } from './NodeContextMenu';

import { WorkflowLibrary } from '../../components/WorkflowLibrary';
import { WorkflowExport } from '../../components/WorkflowExport';
import { useI18n } from '../../hooks/useI18n';

const nodeTypes = {
  boss: BossNode,
  manager: ManagerNode,
  worker: WorkerNode,
  github: GitHubNode,
};

const nodeRoleDefaults: Record<string, string> = {
  boss: 'CEO',
  manager: 'Backend',
  worker: 'Developer',
};

interface CanvasProps {
  mode: 'canvas' | 'chat';
  onModeChange: (mode: 'canvas' | 'chat') => void;
  hasUnreadMessages: boolean;
}

export function Canvas({ mode }: CanvasProps) {
  const nodes = useTaskStore((state) => state.nodes);
  const edges = useTaskStore((state) => state.edges);
  const addEdgeToStore = useTaskStore((state) => state.addEdge);
  const addNodeToStore = useTaskStore((state) => state.addNode);
  const updateNode = useTaskStore((state) => state.updateNode);
  const removeNode = useTaskStore((state) => state.removeNode);
  const { t } = useI18n();
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);

  const [selectedNodeIds, setSelectedNodeIds] = useState<Set<string>>(new Set());

  // Преобразование nodes из Zustand формата в ReactFlow формат
  const rfNodes = useMemo(() => {
    return nodes.map((node) => {
      // Определяем, подключена ли нода (есть ли входящие или исходящие соединения)
      const isConnected = edges.some(edge => edge.from === node.id || edge.to === node.id);

      return {
        id: node.id,
        type: node.type,
        position: node.position,
        data: { ...node, isConnected },
        selected: selectedNodeIds.has(node.id),
      };
    });
  }, [nodes, edges, selectedNodeIds]);

  // Преобразование edges из Zustand формата в ReactFlow формат
  const rfEdges = useMemo(() => {
    return edges.map((edge) => ({
      id: `${edge.from}-${edge.to}`,
      source: edge.from,
      target: edge.to,
      type: 'default',
    }));
  }, [edges]);

  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [workflowLibraryOpen, setWorkflowLibraryOpen] = useState(false);
  const [contextMenu, setContextMenu] = useState<{
    x: number;
    y: number;
    nodeId: string;
    nodeType: AgentNodeType;
    nodeRole: string;
  } | null>(null);

  // Обработчик изменений nodes в ReactFlow
  const onNodesChange = useCallback((changes: OnNodesChange) => {
    changes.forEach((change) => {
      if (change.type === 'select') {
        // Обработка выбора/снятия выбора ноды
        if (change.selected) {
          setSelectedNodeIds(prev => new Set([...prev, change.id]));
        } else {
          setSelectedNodeIds(prev => {
            const newSet = new Set(prev);
            newSet.delete(change.id);
            return newSet;
          });
        }
      } else if (change.type === 'position' && change.position) {
        // Перемещение ноды - обновляем позицию в store
        updateNode(change.id, { position: change.position });
      } else if (change.type === 'remove') {
        // Удаление ноды
        removeNode(change.id);
      }
    });
  }, [updateNode, removeNode]);

  // Обработчик соединений между нодами
  const onConnect = useCallback((connection: Connection) => {
    if (connection.source && connection.target) {
      // Добавляем новое соединение в Zustand store
      addEdgeToStore({
        from: connection.source,
        to: connection.target,
      });
    }
  }, [addEdgeToStore]);


  // Импорт workflow из библиотеки на канвас
  const handleImportWorkflow = useCallback((workflowData: { nodes: any[]; edges: any[] }) => {
    // Очищаем текущий канвас
    useTaskStore.getState().resetTask();

    // Создаём новые ноды из импортированного workflow
    workflowData.nodes.forEach((nodeData) => {
      addNodeToStore({
        id: nodeData.id || `node-${Date.now()}-${Math.random()}`,
        type: nodeData.type as AgentNodeType,
        role: nodeData.role || 'Unknown',
        status: nodeData.status || 'pending',
        progress: 0,
        position: nodeData.position || { x: 0, y: 0 },
        workerCount: nodeData.workerCount,
        filesCount: nodeData.filesCount,
        techStack: nodeData.techStack,
      } as Omit<AgentNode, 'progress'>);
    });

    // Создаём соединения
    workflowData.edges.forEach((edgeData) => {
      addEdgeToStore({ from: edgeData.from, to: edgeData.to });
    });
  }, [addNodeToStore, addEdgeToStore]);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onCanvasDrop = useCallback(
    (event: React.DragEvent) => {
      // Only handle drops from external sources (sidebar), not internal ReactFlow nodes
      const workflowDataStr = event.dataTransfer.getData('application/reactflow/workflow-data');
      if (workflowDataStr && workflowDataStr.trim()) {
        event.preventDefault();
        try {
          const workflowData = JSON.parse(workflowDataStr);
          handleImportWorkflow(workflowData);
          return;
        } catch {
          // ignore
        }
      }

      const type = event.dataTransfer.getData('application/reactflow/node-type') as 'boss' | 'manager' | 'worker';
      if (type && reactFlowInstance) {
        event.preventDefault();
        const position = reactFlowInstance.screenToFlowPosition({
          x: event.clientX,
          y: event.clientY,
        });
        addNodeToStore({
          id: `node-${Date.now()}`,
          type,
          role: nodeRoleDefaults[type],
          status: 'pending',
          progress: 0,
          position,
        });
      }
    },
    [addNodeToStore, handleImportWorkflow, reactFlowInstance]
  );

  // Импорт из JSON строки
  const handleImportJSON = useCallback((jsonText: string) => {
    try {
      const data = JSON.parse(jsonText);
      if (data.nodes && data.edges) {
        handleImportWorkflow({ nodes: data.nodes, edges: data.edges });
      }
    } catch {
      // invalid JSON — ignore
    }
  }, [handleImportWorkflow]);

  const onDragStartFromSidebar = useCallback((type: 'boss' | 'manager' | 'worker', event: React.DragEvent) => {
    event.dataTransfer.setData('application/reactflow/node-type', type);
    event.dataTransfer.effectAllowed = 'move';
  }, []);

  // Удаление нод по клавише Delete
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Delete' || event.key === 'Backspace') {
        selectedNodeIds.forEach((id) => removeNode(id));
        setSelectedNodeIds(new Set());
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodeIds, removeNode]);

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNodeIds((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(node.id)) {
        newSet.delete(node.id);
      } else {
        newSet.add(node.id);
      }
      return newSet;
    });
  }, []);

  const closeContextMenu = useCallback(() => {
    setContextMenu(null);
  }, []);

  const onNodeContextMenu = useCallback((event: React.MouseEvent, node: Node) => {
    event.preventDefault();
    setContextMenu({
      x: event.clientX,
      y: event.clientY,
      nodeId: node.id,
      nodeType: node.type as AgentNodeType,
      nodeRole: node.data?.role || 'Unknown',
    });
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNodeIds(new Set());
    setContextMenu(null);
  }, []);

  // Экспорт workflow в store при изменении нод/соединений
  useEffect(() => {
    // Don't clear workflow when canvas is empty - keep it for user's next task
    if (nodes.length === 0) {
      return;
    }

    // Строим workflow из текущего канваса
    const managerNodes = nodes.filter((n) => n.type === 'manager');
    const bossNodes = nodes.filter((n) => n.type === 'boss');
    const workerNodes = nodes.filter((n) => n.type === 'worker');

    // Определяем есть ли pre-defined workflow (если есть ноды на канвасе)
    const managers = managerNodes.map((mgr, idx) => {
      // Находим воркеров этого менеджера
      const mgrWorkers = workerNodes.filter((w) =>
        edges.some((e) => e.from === mgr.id && e.to === w.id)
      );

      return {
        role: mgr.role || 'Manager',
        description: `${mgr.role} manager for task execution`,
        priority: idx + 1,
        workers: mgrWorkers.map((w) => ({
          role: w.role || 'Developer',
          description: `${w.role} worker`,
        })),
      };
    });

    // Boss connection check
    const hasBoss = bossNodes.length > 0;
    const hasManagers = managerNodes.length > 0;

    // Если есть ноды — используем predefined workflow
    // Если нет — AI planning (useAiPlanning = true)
    const usePredefined = hasManagers;

    const workflow = usePredefined ? {
      useAiPlanning: false,
      managers,
      architecture: hasBoss
        ? `Custom workflow with ${bossNodes[0]?.role || 'Boss'} and ${managerNodes.length} managers`
        : `Custom workflow with ${managerNodes.length} managers`,
      techStack: bossNodes[0]?.techStack || [],
    } : null;

    useTaskStore.getState().setWorkflow(workflow);
  }, [nodes, edges]);

  return (
    <div className="flex-1 bg-[var(--bg-canvas)] relative">
      {/* Кнопка открытия панели - только в режиме canvas */}
      {mode === 'canvas' && (
        <button
          onClick={() => setSidebarOpen(true)}
          className={`absolute right-4 top-4 z-10 bg-[var(--surface)] text-[var(--text)] border border-[var(--border)] rounded-lg px-4 py-2 hover:bg-[var(--accent)] hover:text-white transition-colors shadow-lg ${
            sidebarOpen ? 'hidden' : ''
          }`}
        >
          + {t('sidebar.addAgent')}
        </button>
      )}

      <div ref={reactFlowWrapper} className="w-full h-full"
        onDragOver={onDragOver}
        onDrop={onCanvasDrop}
      >
        <ReactFlow
          nodes={rfNodes}
          edges={rfEdges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          onNodeContextMenu={onNodeContextMenu}
          onPaneClick={onPaneClick}
          fitView
          className="bg-[var(--bg-canvas)]"
          proOptions={{ hideAttribution: true }}
          onInit={(instance) => setReactFlowInstance(instance)}
        >
          <Controls />
          <Background
            color="var(--edge-color)"
            gap={16}
            size={1}
          />
        </ReactFlow>
      </div>

      {/* Правая панель с шаблонами */}
      <NodeSidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onDragStart={onDragStartFromSidebar}
        onOpenWorkflowLibrary={() => setWorkflowLibraryOpen(true)}
      />



      {/* Библиотека Workflow */}
      {workflowLibraryOpen && (
        <WorkflowLibrary
          onClose={() => setWorkflowLibraryOpen(false)}
          onImportWorkflow={handleImportWorkflow}
        />
      )}

      {/* Экспорт/Импорт Workflow */}
      <WorkflowExport onImportJSON={handleImportJSON} />

      {/* Контекстное меню */}
      {contextMenu && (
        <NodeContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          nodeId={contextMenu.nodeId}
          nodeType={contextMenu.nodeType}
          nodeRole={contextMenu.nodeRole}
          onClose={closeContextMenu}
        />
      )}
    </div>
  );
}

import { useCallback, useMemo, useState, useRef, useEffect } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  useReactFlow,
  useNodesState,
  type Node,
  type Edge,
  type Connection,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useTaskStore, type AgentNodeType } from '../../stores/taskStore';
import { BossNode, ManagerNode, WorkerNode, ZIPArchiveNode } from './nodes';
import { NodeSidebar } from './NodeSidebar';
import { NodeContextMenu } from './NodeContextMenu';
import { useI18n } from '../../hooks/useI18n';

const nodeTypes = {
  boss: BossNode,
  manager: ManagerNode,
  worker: WorkerNode,
  zip: ZIPArchiveNode,
};

const nodeRoleDefaults: Record<string, string> = {
  boss: 'CEO',
  manager: 'Backend',
  worker: 'Developer',
};

export function Canvas() {
  const nodes = useTaskStore((state) => state.nodes);
  const edges = useTaskStore((state) => state.edges);
  const addEdgeToStore = useTaskStore((state) => state.addEdge);
  const addNodeToStore = useTaskStore((state) => state.addNode);
  const removeNode = useTaskStore((state) => state.removeNode);
  const { t } = useI18n();

  const { screenToFlowPosition: rfScreenToFlow } = useReactFlow();
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [selectedNodeIds, setSelectedNodeIds] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<{
    x: number;
    y: number;
    nodeId: string;
    nodeType: AgentNodeType;
    nodeRole: string;
  } | null>(null);

  const rfNodes: Node[] = useMemo(() => {
    // Определяем какие ноды подключены
    const connectedNodeIds = new Set<string>();
    edges.forEach((edge) => {
      connectedNodeIds.add(edge.from);
      connectedNodeIds.add(edge.to);
    });

    return nodes.map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position || { x: 0, y: 0 },
      data: {
        role: node.role,
        status: node.status,
        workerCount: node.workerCount,
        filesCount: node.filesCount,
        progress: node.progress,
        techStack: node.techStack,
        // ZIP Archive specific fields
        fileName: node.fileName,
        fileSize: node.fileSize,
        // Статус подключённости
        isConnected: connectedNodeIds.has(node.id),
      },
    }));
  }, [nodes, edges]);

  const rfEdges: Edge[] = useMemo(() =>
    edges.map((edge, idx) => ({
      id: `e-${edge.from}-${edge.to}-${idx}`,
      source: edge.from,
      target: edge.to,
      animated: true,
      style: { stroke: 'var(--edge-color)', strokeWidth: 2 },
    })), [edges]);

  const onConnect = useCallback(
    (params: Connection) => {
      if (params.source && params.target) {
        addEdgeToStore({ from: params.source, to: params.target });
      }
    },
    [addEdgeToStore]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow/node-type') as 'boss' | 'manager' | 'worker';
      if (!type) return;

      const position = rfScreenToFlow(
        {
          x: event.clientX,
          y: event.clientY,
        },
        { snapToGrid: true }
      );

      addNodeToStore({
        id: `node-${Date.now()}`,
        type,
        role: nodeRoleDefaults[type],
        status: 'pending',
        progress: 0,
        position,
      });
    },
    [rfScreenToFlow, addNodeToStore]
  );

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

  const onPaneClick = useCallback(() => {
    setSelectedNodeIds(new Set());
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

  // Экспорт workflow в store при изменении нод/соединений
  useEffect(() => {
    if (nodes.length === 0) {
      useTaskStore.getState().setWorkflow(null);
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
      {/* Кнопка открытия панели */}
      <button
        onClick={() => setSidebarOpen(true)}
        className={`absolute right-4 top-4 z-10 bg-[var(--surface)] text-[var(--text)] border border-[var(--border)] rounded-lg px-4 py-2 hover:bg-[var(--accent)] hover:text-white transition-colors shadow-lg ${
          sidebarOpen ? 'hidden' : ''
        }`}
      >
        + {t('sidebar.addAgent')}
      </button>

      <div ref={reactFlowWrapper} className="w-full h-full">
        <ReactFlow
          nodes={rfNodes}
          edges={rfEdges}
          nodeTypes={nodeTypes}
          onConnect={onConnect}
          onDragOver={onDragOver}
          onDrop={onDrop}
          onNodeClick={onNodeClick}
          onNodeContextMenu={onNodeContextMenu}
          onPaneClick={onPaneClick}
          fitView
          className="bg-[var(--bg-canvas)]"
          proOptions={{ hideAttribution: true }}
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
      />

      {/* Контекстное меню */}
      {contextMenu && (
        <NodeContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          nodeId={contextMenu.nodeId}
          nodeType={contextMenu.nodeType}
          nodeRole={contextMenu.nodeRole}
          onClose={() => setContextMenu(null)}
        />
      )}
    </div>
  );
}

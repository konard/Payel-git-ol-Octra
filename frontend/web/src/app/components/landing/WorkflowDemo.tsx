import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ReactFlow,
  Node,
  Edge,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  Handle,
  Position,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { motion } from 'motion/react';
// import bossImage from '../../../images/boss-image.png';
// import managerImage from '../../../images/manager-image.png';
// import workerImage from '../../../images/worker-image.png';

// Custom node components for landing
function LandingBossNode({ data }: { data: any }) {
  return (
    <div className="bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 border-orange-500 min-w-[220px] overflow-hidden">
      <Handle type="target" position={Position.Top} />
      <div className="bg-gradient-to-r from-orange-500 to-orange-600 text-white px-4 py-3 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-orange-500 rounded flex items-center justify-center text-white font-bold">B</div>
          <span className="font-bold text-base">BOSS</span>
        </div>
        <span className="text-sm font-medium">COORDINATING</span>
      </div>
      <div className="p-3 space-y-2">
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Роль:</span> {data.role}
        </div>
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Статус:</span> <span className="text-orange-500 font-semibold">WORKING</span>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

function LandingManagerNode({ data }: { data: any }) {
  return (
    <div className="bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 border-blue-500 min-w-[200px] overflow-hidden">
      <Handle type="target" position={Position.Top} />
      <div className="bg-gradient-to-r from-blue-500 to-blue-600 text-white px-4 py-3 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-500 rounded flex items-center justify-center text-white font-bold">M</div>
          <span className="font-bold text-base">MANAGER</span>
        </div>
        <span className="text-sm font-medium">WORKING</span>
      </div>
      <div className="p-3 space-y-2">
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Роль:</span> {data.role}
        </div>
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Статус:</span> <span className="text-blue-500 font-semibold">ACTIVE</span>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

function LandingWorkerNode({ data }: { data: any }) {
  return (
    <div className="bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 border-green-500 min-w-[180px] overflow-hidden">
      <Handle type="target" position={Position.Top} />
      <div className="bg-gradient-to-r from-green-500 to-green-600 text-white px-4 py-3 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-green-500 rounded flex items-center justify-center text-white font-bold">W</div>
          <span className="font-bold text-base">WORKER</span>
        </div>
        <span className="text-sm font-medium">DONE</span>
      </div>
      <div className="p-3 space-y-2">
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Роль:</span> {data.role}
        </div>
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Статус:</span> <span className="text-green-500 font-semibold">COMPLETE</span>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} id="source" />
    </div>
  );
}

const nodeTypes = {
  boss: LandingBossNode,
  manager: LandingManagerNode,
  worker: LandingWorkerNode,
};

// Realistic node positions from the actual application
const initialNodes: Node[] = [
  {
    id: 'boss-1',
    type: 'boss',
    position: { x: 377, y: 42 },
    data: { role: 'Project Lead' },
  },
  {
    id: 'manager-1',
    type: 'manager',
    position: { x: 50, y: 280 },
    data: { role: 'Research' },
  },
  {
    id: 'manager-2',
    type: 'manager',
    position: { x: 400, y: 280 },
    data: { role: 'Development' },
  },
  {
    id: 'manager-3',
    type: 'manager',
    position: { x: 750, y: 280 },
    data: { role: 'Testing' },
  },
  {
    id: 'worker-1',
    type: 'worker',
    position: { x: -60, y: 594 },
    data: { role: 'Search' },
  },
  {
    id: 'worker-2',
    type: 'worker',
    position: { x: 146, y: 741 },
    data: { role: 'Analyze' },
  },
  {
    id: 'worker-3',
    type: 'worker',
    position: { x: 266, y: 479 },
    data: { role: 'Frontend' },
  },
  {
    id: 'worker-4',
    type: 'worker',
    position: { x: 456, y: 630 },
    data: { role: 'Backend' },
  },
  {
    id: 'worker-5',
    type: 'worker',
    position: { x: 641, y: 485 },
    data: { role: 'Unit Tests' },
  },
  {
    id: 'worker-6',
    type: 'worker',
    position: { x: 906, y: 575 },
    data: { role: 'Integration' },
  },
];

const initialEdges: Edge[] = [
  { id: 'e1', source: 'boss-1', target: 'manager-1', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e2', source: 'boss-1', target: 'manager-2', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e3', source: 'boss-1', target: 'manager-3', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e4', source: 'manager-1', target: 'worker-1', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e5', source: 'manager-1', target: 'worker-2', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e6', source: 'manager-2', target: 'worker-3', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e7', source: 'manager-2', target: 'worker-4', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e8', source: 'manager-3', target: 'worker-5', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
  { id: 'e9', source: 'manager-3', target: 'worker-6', animated: false, style: { stroke: 'var(--edge-color)', strokeWidth: 2 } },
];

export function WorkflowDemo() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [activeNodeIndex, setActiveNodeIndex] = useState(0);

  // Cycle through nodes to show execution animation
  useEffect(() => {
    const allNodeIds = initialNodes.map((n) => n.id);
    const interval = setInterval(() => {
      setActiveNodeIndex((prev) => (prev + 1) % allNodeIds.length);
    }, 1500);

    return () => clearInterval(interval);
  }, []);

  // Update edges animation based on active node
  useEffect(() => {
    const activeNodeId = initialNodes[activeNodeIndex]?.id;
    setEdges((eds) =>
      eds.map((edge) => ({
        ...edge,
        style: {
          stroke: edge.source === activeNodeId || edge.target === activeNodeId
            ? 'var(--edge-active)'
            : 'var(--edge-color)',
          strokeWidth: edge.source === activeNodeId || edge.target === activeNodeId ? 3 : 2,
        },
        animated: edge.source === activeNodeId || edge.target === activeNodeId,
      }))
    );
  }, [activeNodeIndex, setEdges]);

  return (
    <section id="workflow" className="py-20 px-4 sm:px-6 lg:px-8 bg-[var(--surface)]">
      <div className="max-w-7xl mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Text content */}
          <div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--accent)]/10 border border-[var(--accent)]/20 rounded-full text-sm font-medium text-[var(--accent)] mb-6"
            >
              Как это работает
            </motion.div>

            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="text-4xl sm:text-5xl font-bold text-[var(--text)] mb-6"
            >
              Визуальное управление
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> AI-агентами</span>
            </motion.h2>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="text-lg text-[var(--text-muted)] mb-8"
            >
              Создавайте сложные workflow с помощью простого drag-and-drop.
              Соединяйте агентов, настраивайте роли и наблюдайте за выполнением задач в реальном времени.
            </motion.p>

            <motion.ul
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="space-y-4"
            >
              {[
                'Drag-and-drop перемещение нод',
                'Анимация выполнения в реальном времени',
                'Иерархия Boss → Manager → Worker',
                'Автоматическое масштабирование',
              ].map((feature, index) => (
                <li key={index} className="flex items-start gap-3">
                  <div className="flex-shrink-0 w-6 h-6 bg-gradient-to-br from-orange-500 to-orange-600 rounded-full flex items-center justify-center text-white text-sm font-bold mt-0.5">
                    ✓
                  </div>
                  <span className="text-[var(--text)] font-medium">{feature}</span>
                </li>
              ))}
            </motion.ul>
          </div>

          {/* React Flow Canvas */}
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="relative"
          >
            <div className="relative rounded-xl overflow-hidden shadow-2xl border border-[var(--border)] bg-[var(--bg-canvas)]" style={{ height: '700px' }}>
              <ReactFlow
                nodes={nodes}
                edges={edges}
                nodeTypes={nodeTypes}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                fitView
                fitViewOptions={{ padding: 0.2 }}
                panOnDrag
                zoomOnScroll={false}
                zoomOnPinch={false}
                proOptions={{ hideAttribution: true }}
              >
                <Background color="var(--edge-color)" gap={16} size={1} />
                <Controls showInteractive={false} />
              </ReactFlow>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}

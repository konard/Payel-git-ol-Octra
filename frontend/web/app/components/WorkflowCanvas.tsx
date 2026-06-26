'use client';

import {
  Background,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
} from '@xyflow/react';
import { useEffect, useMemo, useState } from 'react';

type AgentNodeData = {
  [key: string]: unknown;
  role: string;
  agent_id: string;
  status: string;
  progress: string;
  endpoint: string;
  detail: string;
};

type AgentNode = Node<AgentNodeData, 'agentNode'>;

function AgentFlowNode({ data }: NodeProps<AgentNode>) {
  return (
    <article className="octra-flow-node">
      <Handle className="octra-flow-handle" type="target" position={Position.Left} />
      <div className="flow-node-kicker">{data.status}</div>
      <div className="flow-node-title">
        <span>{data.role}</span>
        <strong>{data.progress}</strong>
      </div>
      <p>{data.detail}</p>
      <dl>
        <div>
          <dt>agent_id</dt>
          <dd>{data.agent_id}</dd>
        </div>
        <div>
          <dt>endpoint</dt>
          <dd>{data.endpoint}</dd>
        </div>
      </dl>
      <Handle className="octra-flow-handle" type="source" position={Position.Right} />
    </article>
  );
}

const nodeTypes = {
  agentNode: AgentFlowNode,
};

const graphData = [
  {
    id: 'task-create',
    type: 'agentNode',
    data: {
      role: 'Task ingress',
      agent_id: 'ws:task/create',
      status: 'connected',
      progress: '5%',
      endpoint: 'GET /task/create',
      detail: 'Authenticated WebSocket receives CreateTaskRequest JSON.',
    },
  },
  {
    id: 'boss-planner',
    type: 'agentNode',
    data: {
      role: 'Boss planner',
      agent_id: 'boss:planner:01',
      status: 'boss_planning',
      progress: '32%',
      endpoint: 'CreateTaskStream',
      detail: 'Builds manager roles and streams TaskUpdate messages.',
    },
  },
  {
    id: 'workflow-store',
    type: 'agentNode',
    data: {
      role: 'Workflow library',
      agent_id: 'workflow:template',
      status: 'available',
      progress: '100%',
      endpoint: 'POST /workflows',
      detail: 'Stores saved nodes and edges for reusable workflow templates.',
    },
  },
  {
    id: 'manager-team',
    type: 'agentNode',
    data: {
      role: 'Manager team',
      agent_id: 'mgr:frontend:04',
      status: 'managers_assigned',
      progress: '58%',
      endpoint: 'ManagerConfig.workers',
      detail: 'Routes work to predefined managers and worker roles.',
    },
  },
  {
    id: 'worker-code',
    type: 'agentNode',
    data: {
      role: 'Worker code',
      agent_id: 'worker:code:17',
      status: 'processing',
      progress: '74%',
      endpoint: 'TaskUpdate.data',
      detail: 'Writes implementation artifacts and emits progress payloads.',
    },
  },
  {
    id: 'redis-state',
    type: 'agentNode',
    data: {
      role: 'Redis stream state',
      agent_id: 'redis:stream:history',
      status: 'persisting',
      progress: '256',
      endpoint: 'STREAM:<task_id>',
      detail: 'Keeps reconnect history and PubSub updates per task_id.',
    },
  },
  {
    id: 'status-api',
    type: 'agentNode',
    data: {
      role: 'Status API',
      agent_id: 'api:task/status',
      status: 'ready',
      progress: '72%',
      endpoint: 'GET /task/status',
      detail: 'Returns task progress and current stream state.',
    },
  },
] satisfies Array<Omit<AgentNode, 'position'>>;

const desktopPositions: Record<string, AgentNode['position']> = {
  'task-create': { x: 20, y: 170 },
  'boss-planner': { x: 260, y: 52 },
  'workflow-store': { x: 260, y: 300 },
  'manager-team': { x: 500, y: 112 },
  'worker-code': { x: 735, y: 42 },
  'redis-state': { x: 735, y: 292 },
  'status-api': { x: 970, y: 178 },
};

const compactPositions: Record<string, AgentNode['position']> = {
  'task-create': { x: 20, y: 30 },
  'boss-planner': { x: 280, y: 30 },
  'workflow-store': { x: 20, y: 182 },
  'manager-team': { x: 280, y: 182 },
  'worker-code': { x: 20, y: 334 },
  'redis-state': { x: 280, y: 334 },
  'status-api': { x: 150, y: 486 },
};

function buildNodes(compact: boolean): AgentNode[] {
  const positions = compact ? compactPositions : desktopPositions;
  return graphData.map((node) => ({
    ...node,
    position: positions[node.id],
  }));
}

const edges: Edge[] = [
  { id: 'task-boss', source: 'task-create', target: 'boss-planner', animated: true },
  { id: 'task-workflow', source: 'task-create', target: 'workflow-store' },
  { id: 'boss-manager', source: 'boss-planner', target: 'manager-team', animated: true },
  { id: 'workflow-manager', source: 'workflow-store', target: 'manager-team' },
  { id: 'manager-worker', source: 'manager-team', target: 'worker-code', animated: true },
  { id: 'manager-redis', source: 'manager-team', target: 'redis-state' },
  { id: 'worker-status', source: 'worker-code', target: 'status-api', animated: true },
  { id: 'redis-status', source: 'redis-state', target: 'status-api' },
];

const defaultEdgeOptions = {
  style: { stroke: 'rgba(245, 245, 245, 0.62)', strokeWidth: 1.5 },
};

export function WorkflowCanvas() {
  const [isCompact, setIsCompact] = useState(false);

  useEffect(() => {
    const query = window.matchMedia('(max-width: 780px)');
    const sync = () => setIsCompact(query.matches);

    sync();
    query.addEventListener('change', sync);
    return () => query.removeEventListener('change', sync);
  }, []);

  const visibleNodes = useMemo(() => buildNodes(isCompact), [isCompact]);

  return (
    <ReactFlow
      className="octra-flow"
      nodes={visibleNodes}
      edges={edges}
      nodeTypes={nodeTypes}
      defaultEdgeOptions={defaultEdgeOptions}
      fitView
      fitViewOptions={{ padding: 0.16 }}
      minZoom={0.45}
      maxZoom={1.5}
      nodesDraggable={false}
      nodesConnectable={false}
      proOptions={{ hideAttribution: true }}
    >
      <Background color="rgba(255,255,255,0.16)" gap={28} size={1} />
      <MiniMap
        pannable
        zoomable
        nodeBorderRadius={4}
        nodeColor="rgba(255,255,255,0.36)"
        maskColor="rgba(0,0,0,0.5)"
      />
      <Controls showInteractive={false} />
    </ReactFlow>
  );
}

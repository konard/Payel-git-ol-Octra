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

type EnvironmentNodeData = {
  [key: string]: unknown;
  title: string;
  resource_id: string;
  status: string;
  metric: string;
  endpoint: string;
  detail: string;
};

type EnvironmentNode = Node<EnvironmentNodeData, 'environmentNode'>;

function EnvironmentFlowNode({ data }: NodeProps<EnvironmentNode>) {
  return (
    <article className="octra-flow-node">
      <Handle className="octra-flow-handle" type="target" position={Position.Left} />
      <div className="flow-node-kicker">{data.status}</div>
      <div className="flow-node-title">
        <span>{data.title}</span>
        <strong>{data.metric}</strong>
      </div>
      <p>{data.detail}</p>
      <dl>
        <div>
          <dt>resource_id</dt>
          <dd>{data.resource_id}</dd>
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
  environmentNode: EnvironmentFlowNode,
};

const graphData = [
  {
    id: 'api-chat',
    type: 'environmentNode',
    data: {
      title: 'API chat request',
      resource_id: 'route:/api/chat',
      status: 'incoming',
      metric: 'prompt',
      endpoint: 'POST /api/chat',
      detail: 'Receives a prompt with the octra-api-token header.',
    },
  },
  {
    id: 'token-check',
    type: 'environmentNode',
    data: {
      title: 'Token validation',
      resource_id: 'auth:octra-api-token',
      status: 'verified',
      metric: 'strict',
      endpoint: 'middleware',
      detail: 'Finds the user environment for the supplied API token.',
    },
  },
  {
    id: 'user-environment',
    type: 'environmentNode',
    data: {
      title: 'User environment',
      resource_id: 'env:usr_742',
      status: 'active',
      metric: 'Nix',
      endpoint: 'POST /environment',
      detail: 'Stores LLM config, optional CLI, and selected skills.',
    },
  },
  {
    id: 'nix-profile',
    type: 'environmentNode',
    data: {
      title: 'Nix profile',
      resource_id: 'nix:usr_742',
      status: 'provisioned',
      metric: '3 skills',
      endpoint: 'nix profile install',
      detail: 'Installs the selected CLI and skill packages in isolation.',
    },
  },
  {
    id: 'skill-set',
    type: 'environmentNode',
    data: {
      title: 'Skill set',
      resource_id: 'skills:filesystem+github',
      status: 'enabled',
      metric: 'per prompt',
      endpoint: 'skills[]',
      detail: 'Lets each request enable only the tools it needs.',
    },
  },
  {
    id: 'redis-cli-state',
    type: 'environmentNode',
    data: {
      title: 'Redis CLI state',
      resource_id: 'user:742:cli_state',
      status: 'alive',
      metric: '38m TTL',
      endpoint: 'Redis',
      detail: 'Keeps PID, optional port, startup time, and process TTL.',
    },
  },
  {
    id: 'cli-process',
    type: 'environmentNode',
    data: {
      title: 'Claude Code CLI',
      resource_id: 'pid:4812',
      status: 'running',
      metric: 'stdin/stdout',
      endpoint: 'Nix subprocess',
      detail: 'Reuses a warm CLI process until TTL expiry or crash.',
    },
  },
  {
    id: 'response',
    type: 'environmentNode',
    data: {
      title: 'Response',
      resource_id: 'json:response',
      status: 'returned',
      metric: '200 OK',
      endpoint: 'HTTP JSON',
      detail: 'Returns the CLI or direct LLM output to the caller.',
    },
  },
] satisfies Array<Omit<EnvironmentNode, 'position'>>;

const desktopPositions: Record<string, EnvironmentNode['position']> = {
  'api-chat': { x: 20, y: 170 },
  'token-check': { x: 260, y: 52 },
  'user-environment': { x: 260, y: 300 },
  'nix-profile': { x: 500, y: 92 },
  'skill-set': { x: 500, y: 300 },
  'redis-cli-state': { x: 735, y: 42 },
  'cli-process': { x: 735, y: 292 },
  response: { x: 970, y: 178 },
};

const compactPositions: Record<string, EnvironmentNode['position']> = {
  'api-chat': { x: 20, y: 30 },
  'token-check': { x: 280, y: 30 },
  'user-environment': { x: 20, y: 182 },
  'nix-profile': { x: 280, y: 182 },
  'skill-set': { x: 20, y: 334 },
  'redis-cli-state': { x: 280, y: 334 },
  'cli-process': { x: 20, y: 486 },
  response: { x: 280, y: 486 },
};

function buildNodes(compact: boolean): EnvironmentNode[] {
  const positions = compact ? compactPositions : desktopPositions;
  return graphData.map((node) => ({
    ...node,
    position: positions[node.id],
  }));
}

const edges: Edge[] = [
  { id: 'chat-token', source: 'api-chat', target: 'token-check', animated: true },
  { id: 'token-environment', source: 'token-check', target: 'user-environment', animated: true },
  { id: 'environment-nix', source: 'user-environment', target: 'nix-profile' },
  { id: 'environment-skills', source: 'user-environment', target: 'skill-set' },
  { id: 'nix-redis', source: 'nix-profile', target: 'redis-cli-state' },
  { id: 'skills-cli', source: 'skill-set', target: 'cli-process' },
  { id: 'redis-cli', source: 'redis-cli-state', target: 'cli-process', animated: true },
  { id: 'cli-response', source: 'cli-process', target: 'response', animated: true },
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

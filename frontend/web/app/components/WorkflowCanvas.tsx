'use client';

import {
  Background,
  ConnectionLineType,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  addEdge,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type Node,
  type NodeProps,
} from '@xyflow/react';
import { Bot, Box, Braces, Cpu, Globe2, Layers3 } from 'lucide-react';
import { useCallback, useEffect, useMemo } from 'react';

export type WorkflowCanvasItem = {
  id: string;
  kind: 'provider' | 'cli' | 'skill' | 'custom_provider' | 'environment';
  name: string;
  detail?: string;
  description?: string;
  meta?: Record<string, string | undefined>;
  positionX?: number;
  positionY?: number;
};

type WorkflowNodeData = {
  item: WorkflowCanvasItem;
};

type WorkflowNode = Node<WorkflowNodeData, 'octra'>;

type WorkflowCanvasProps = {
  items?: WorkflowCanvasItem[];
  onItemsChange?: (items: WorkflowCanvasItem[]) => void;
};

const nodeTypes = {
  octra: WorkflowNodeCard,
};

const iconByKind = {
  provider: Globe2,
  cli: Cpu,
  skill: Box,
  custom_provider: Braces,
  environment: Layers3,
} as const;

export function WorkflowCanvas({ items = [], onItemsChange }: WorkflowCanvasProps) {
  const initialNodes = useMemo(() => buildNodes(items), [items]);
  const [nodes, setNodes, onNodesChange] = useNodesState<WorkflowNode>(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  useEffect(() => {
    setNodes(buildNodes(items));
  }, [items, setNodes]);

  const handleNodesChange = useCallback(
    (changes: any) => {
      onNodesChange(changes);
      if (!onItemsChange) return;
      const positionChanges = changes.filter((c: any) => c.type === 'position' && c.dragging === false);
      if (positionChanges.length === 0) return;
      const updated = items.map((item) => {
        const posChange = positionChanges.find((c: any) => c.id === item.id);
        if (!posChange) return item;
        return {
          ...item,
          positionX: posChange.position?.x ?? item.positionX,
          positionY: posChange.position?.y ?? item.positionY,
        };
      });
      onItemsChange(updated);
    },
    [items, onItemsChange, onNodesChange],
  );

  const onConnect = useCallback(
    (connection: Connection) => setEdges((current) => addEdge({ ...connection, animated: true, type: 'smoothstep' }, current)),
    [setEdges],
  );

  return (
      <ReactFlow
      className="octra-flow"
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      onNodesChange={handleNodesChange}
      onEdgesChange={onEdgesChange}
      onConnect={onConnect}
      connectionLineType={ConnectionLineType.SmoothStep}
      fitView
      fitViewOptions={{ padding: 0.16 }}
      minZoom={0.45}
      maxZoom={1.5}
      nodesDraggable
      nodesConnectable
      proOptions={{ hideAttribution: true }}
    >
      <Background color="rgba(255,255,255,0.6)" gap={28} size={1.5} />
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

function buildNodes(items: WorkflowCanvasItem[]): WorkflowNode[] {
  return items.map((item, index) => ({
    id: item.id,
    type: 'octra',
    position: {
      x: item.positionX ?? 80 + (index % 4) * 280,
      y: item.positionY ?? 86 + Math.floor(index / 4) * 220,
    },
    data: { item },
  }));
}

function WorkflowNodeCard({ data }: NodeProps<WorkflowNode>) {
  const item = data.item;
  const Icon = iconByKind[item.kind] ?? Bot;
  const metaRows = Object.entries(item.meta ?? {}).filter((entry): entry is [string, string] => Boolean(entry[1])).slice(0, 3);

  return (
    <article className={`octra-flow-node kind-${item.kind}`}>
      <Handle className="octra-flow-handle" type="target" position={Position.Left} />
      <div className="flow-node-title">
        <span className="flow-node-icon">
          <Icon size={17} />
        </span>
        <div>
          <strong>{item.name}</strong>
          <span>{nodeKindLabel(item.kind)}</span>
        </div>
      </div>
      {item.detail ? <p>{item.detail}</p> : null}
      {metaRows.length > 0 ? (
        <dl>
          {metaRows.map(([key, value]) => (
            <div key={key}>
              <dt>{key}</dt>
              <dd>{value}</dd>
            </div>
          ))}
        </dl>
      ) : null}
      <Handle className="octra-flow-handle" type="source" position={Position.Right} />
    </article>
  );
}

function nodeKindLabel(kind: WorkflowCanvasItem['kind']) {
  switch (kind) {
    case 'provider':
      return 'Provider';
    case 'cli':
      return 'CLI';
    case 'skill':
      return 'Skill';
    case 'custom_provider':
      return 'Custom';
    case 'environment':
      return 'Environment';
    default:
      return 'Node';
  }
}

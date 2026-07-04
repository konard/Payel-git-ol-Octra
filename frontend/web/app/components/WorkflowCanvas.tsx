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
  applyEdgeChanges,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type EdgeChange,
  type Node,
  type NodeProps,
} from '@xyflow/react';
import { Bot, Box, Braces, Cable, Cpu, Globe2, Layers3, Trash2, Unplug, Pencil } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { EditNodeModal } from './EditNodeModal';

export type WorkflowCanvasItem = {
  id: string;
  kind: 'provider' | 'cli' | 'skill' | 'custom_provider' | 'environment' | 'mcp_server' | 'adapter';
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
  edges?: Edge[];
  onEdgesChange?: (edges: Edge[]) => void;
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
  mcp_server: Cable,
  adapter: Unplug,
} as const;

export function WorkflowCanvas({ items = [], onItemsChange, edges = [], onEdgesChange }: WorkflowCanvasProps) {
  const initialNodes = useMemo(() => buildNodes(items), [items]);
  const [nodes, setNodes, onNodesChange] = useNodesState<WorkflowNode>(initialNodes);
  const [edgeState, setEdges, onEdgesStateChange] = useEdgesState<Edge>(edges);
  const [contextMenu, setContextMenu] = useState<{ node: WorkflowNode; x: number; y: number } | null>(null);
  const [editItem, setEditItem] = useState<WorkflowCanvasItem | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    console.log('[canvas] WorkflowCanvas useEffect[items]: items count=', items.length, 'ids:', items.map(i => i.id));
    setNodes(buildNodes(items));
  }, [items, setNodes]);

  useEffect(() => {
    setEdges(edges);
  }, [edges, setEdges]);

  const handleNodeContextMenu = useCallback((event: React.MouseEvent, node: WorkflowNode) => {
    event.preventDefault();
    setContextMenu({
      node,
      x: event.clientX,
      y: event.clientY,
    });
  }, []);

  const closeContextMenu = useCallback(() => {
    setContextMenu(null);
  }, []);

  const handleDelete = useCallback(() => {
    if (!contextMenu) return;
    const nodeId = contextMenu.node.id;
    const updatedItems = items.filter((i) => i.id !== nodeId);
    onItemsChange?.(updatedItems);
    const remainingEdges = edgeState.filter((e) => e.source !== nodeId && e.target !== nodeId);
    setEdges(remainingEdges);
    onEdgesChange?.(remainingEdges);
    setContextMenu(null);
  }, [contextMenu, items, onItemsChange, edgeState, setEdges, onEdgesChange]);

  const handleDisconnect = useCallback(() => {
    if (!contextMenu) return;
    const nodeId = contextMenu.node.id;
    const remainingEdges = edgeState.filter((e) => e.source !== nodeId && e.target !== nodeId);
    setEdges(remainingEdges);
    onEdgesChange?.(remainingEdges);
    setContextMenu(null);
  }, [contextMenu, edgeState, setEdges, onEdgesChange]);

  const handleEdit = useCallback(() => {
    if (!contextMenu) return;
    setEditItem(contextMenu.node.data.item);
    setContextMenu(null);
  }, [contextMenu]);

  const handleEditSave = useCallback(
    (updated: WorkflowCanvasItem) => {
      const updatedItems = items.map((i) => (i.id === updated.id ? updated : i));
      onItemsChange?.(updatedItems);
      setEditItem(null);
    },
    [items, onItemsChange],
  );

  const handleNodesChange = useCallback(
    (changes: any) => {
      onNodesChange(changes);
      if (!onItemsChange) return;
      const positionChanges = changes.filter((c: any) => c.type === 'position' && c.dragging === false);
      if (positionChanges.length === 0) {
        const otherChanges = changes.filter((c: any) => c.type !== 'position');
        if (otherChanges.length > 0) console.log('[canvas] handleNodesChange: non-position changes:', otherChanges);
        return;
      }
      console.log('[canvas] handleNodesChange: drag ended for', positionChanges.length, 'node(s):', positionChanges.map((c: any) => ({ id: c.id, pos: c.position })));
      const updated = items.map((item) => {
        const posChange = positionChanges.find((c: any) => c.id === item.id);
        if (!posChange) return item;
        return {
          ...item,
          positionX: posChange.position?.x ?? item.positionX,
          positionY: posChange.position?.y ?? item.positionY,
        };
      });
      console.log('[canvas] handleNodesChange: calling onItemsChange with', updated.length, 'items');
      onItemsChange(updated);
    },
    [items, onItemsChange, onNodesChange],
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      console.log('[canvas] onConnect: new connection', connection);
      const next = addEdge({ ...connection, animated: true, type: 'smoothstep' }, edgeState);
      console.log('[canvas] onConnect: edges now', next.length);
      setEdges(next);
      onEdgesChange?.(next);
    },
    [edgeState, setEdges, onEdgesChange],
  );

  const handleEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      onEdgesStateChange(changes);
      const next = applyEdgeChanges(changes, edgeState);
      onEdgesChange?.(next);
    },
    [onEdgesStateChange, edgeState, onEdgesChange],
  );

  return (
    <>
      <ReactFlow
        className="octra-flow"
        nodes={nodes}
        edges={edgeState}
        nodeTypes={nodeTypes}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onConnect={onConnect}
        onNodeContextMenu={handleNodeContextMenu}
        onPaneClick={closeContextMenu}
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

      {contextMenu && createPortal(
        <div
          ref={menuRef}
          className="node-context-menu"
          style={{ left: contextMenu.x, top: contextMenu.y, position: 'fixed' }}
          role="menu"
        >
          <button type="button" role="menuitem" onClick={handleEdit}>
            <Pencil size={15} />
            Edit
          </button>
          <button type="button" role="menuitem" onClick={handleDisconnect}>
            <Unplug size={15} />
            Disconnect
          </button>
          <button type="button" role="menuitem" className="danger" onClick={handleDelete}>
            <Trash2 size={15} />
            Delete
          </button>
        </div>,
        document.body,
      )}

      {editItem && createPortal(
        <EditNodeModal
          item={editItem}
          onSave={handleEditSave}
          onClose={() => setEditItem(null)}
        />,
        document.body,
      )}
    </>
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
  const metaRows = Object.entries(item.meta ?? {})
    .filter((entry): entry is [string, string] => Boolean(entry[1]))
    .slice(0, 3);

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
              <dt>{formatMetaKey(key)}</dt>
              <dd>{formatMetaValue(key, value)}</dd>
            </div>
          ))}
        </dl>
      ) : null}
      <Handle className="octra-flow-handle" type="source" position={Position.Right} />
    </article>
  );
}

function formatMetaKey(key: string) {
  return key
    .replace(/[_-]+/g, ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function formatMetaValue(key: string, value: string) {
  if (/api[_-]?key|token|secret|password/i.test(key)) {
    return value === 'set' ? value : '******';
  }
  return value;
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
    case 'mcp_server':
      return 'MCP';
    case 'adapter':
      return 'Adapter';
    default:
      return 'Node';
  }
}

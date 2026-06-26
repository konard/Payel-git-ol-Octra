'use client';

import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  type Edge,
  type Node,
} from '@xyflow/react';
import { useMemo } from 'react';

export function WorkflowCanvas() {
  const nodes: Node[] = useMemo(() => [], []);
  const edges: Edge[] = useMemo(() => [], []);

  return (
    <ReactFlow
      className="octra-flow"
      nodes={nodes}
      edges={edges}
      fitView
      fitViewOptions={{ padding: 0.16 }}
      minZoom={0.45}
      maxZoom={1.5}
      nodesDraggable={false}
      nodesConnectable={false}
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

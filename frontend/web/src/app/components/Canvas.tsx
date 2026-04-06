import { useCallback, useMemo } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  type Node,
  type Edge,
  type Connection,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useTaskStore } from '../../stores/taskStore';
import { BossNode, ManagerNode, WorkerNode, ZIPArchiveNode } from './nodes';

const nodeTypes = {
  boss: BossNode,
  manager: ManagerNode,
  worker: WorkerNode,
  zip: ZIPArchiveNode,
};

export function Canvas() {
  const nodes = useTaskStore((state) => state.nodes);
  const edges = useTaskStore((state) => state.edges);
  const addEdgeToStore = useTaskStore((state) => state.addEdge);

  const rfNodes: Node[] = useMemo(() =>
    nodes.map((node) => ({
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
      },
    })), [nodes]);

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

  return (
    <div className="flex-1 bg-[var(--bg-canvas)]">
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        nodeTypes={nodeTypes}
        onConnect={onConnect}
        fitView
        className="bg-[var(--bg-canvas)]"
        proOptions={{ hideAttribution: true }}
      >
        <Controls className="bg-[var(--bg-node)] text-[var(--text-node)]" />
        <Background 
          color="var(--edge-color)" 
          gap={16} 
          size={1} 
        />
      </ReactFlow>
    </div>
  );
}

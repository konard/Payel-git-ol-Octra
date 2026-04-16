import { useState, useCallback } from 'react';
import { ReactFlow, Background, Controls, MiniMap, useNodesState, useEdgesState, addEdge, Connection } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './components/ui/tabs';

// Variant 1 - Corporate Gradient
import { BossNode as BossNode1 } from './components/nodes/variant1/BossNode';
import { ManagerNode as ManagerNode1 } from './components/nodes/variant1/ManagerNode';
import { WorkerNode as WorkerNode1 } from './components/nodes/variant1/WorkerNode';

// Variant 2 - Minimalist Cards
import { BossNode as BossNode2 } from './components/nodes/variant2/BossNode';
import { ManagerNode as ManagerNode2 } from './components/nodes/variant2/ManagerNode';
import { WorkerNode as WorkerNode2 } from './components/nodes/variant2/WorkerNode';

// Variant 3 - Glassmorphism
import { BossNode as BossNode3 } from './components/nodes/variant3/BossNode';
import { ManagerNode as ManagerNode3 } from './components/nodes/variant3/ManagerNode';
import { WorkerNode as WorkerNode3 } from './components/nodes/variant3/WorkerNode';

const nodeTypes1 = {
  boss: BossNode1,
  manager: ManagerNode1,
  worker: WorkerNode1,
};

const nodeTypes2 = {
  boss: BossNode2,
  manager: ManagerNode2,
  worker: WorkerNode2,
};

const nodeTypes3 = {
  boss: BossNode3,
  manager: ManagerNode3,
  worker: WorkerNode3,
};

const initialNodes = [
  {
    id: '1',
    type: 'boss',
    position: { x: 250, y: 50 },
    data: { 
      name: 'Coordinator AI', 
      model: 'GPT-4'
    },
  },
  {
    id: '2',
    type: 'manager',
    position: { x: 100, y: 300 },
    data: { 
      name: 'Task Manager',
      model: 'GPT-3.5 Turbo'
    },
  },
  {
    id: '3',
    type: 'manager',
    position: { x: 450, y: 300 },
    data: { 
      name: 'Content Manager',
      model: 'Claude 3 Opus'
    },
  },
  {
    id: '4',
    type: 'worker',
    position: { x: 50, y: 550 },
    data: { 
      name: 'Code Worker',
      model: 'Claude 3 Sonnet'
    },
  },
  {
    id: '5',
    type: 'worker',
    position: { x: 250, y: 550 },
    data: { 
      name: 'Data Worker',
      model: 'GPT-3.5'
    },
  },
  {
    id: '6',
    type: 'worker',
    position: { x: 450, y: 550 },
    data: { 
      name: 'Text Worker',
      model: 'Claude 3 Haiku'
    },
  },
  {
    id: '7',
    type: 'worker',
    position: { x: 650, y: 550 },
    data: { 
      name: 'Review Worker',
      model: 'GPT-4'
    },
  },
];

const initialEdges = [
  { id: 'e1-2', source: '1', target: '2', animated: true },
  { id: 'e1-3', source: '1', target: '3', animated: true },
  { id: 'e2-4', source: '2', target: '4' },
  { id: 'e2-5', source: '2', target: '5' },
  { id: 'e3-6', source: '3', target: '6' },
  { id: 'e3-7', source: '3', target: '7' },
];

export default function App() {
  const [variant, setVariant] = useState('variant1');
  const [nodes1, setNodes1, onNodesChange1] = useNodesState(initialNodes);
  const [edges1, setEdges1, onEdgesChange1] = useEdgesState(initialEdges);
  const [nodes2, setNodes2, onNodesChange2] = useNodesState(initialNodes);
  const [edges2, setEdges2, onEdgesChange2] = useEdgesState(initialEdges);
  const [nodes3, setNodes3, onNodesChange3] = useNodesState(initialNodes);
  const [edges3, setEdges3, onEdgesChange3] = useEdgesState(initialEdges);

  const onConnect1 = useCallback((params: Connection) => setEdges1((eds) => addEdge(params, eds)), [setEdges1]);
  const onConnect2 = useCallback((params: Connection) => setEdges2((eds) => addEdge(params, eds)), [setEdges2]);
  const onConnect3 = useCallback((params: Connection) => setEdges3((eds) => addEdge(params, eds)), [setEdges3]);

  return (
    <div className="size-full flex flex-col bg-gradient-to-br from-gray-50 to-gray-100">
      <div className="p-6 bg-white border-b border-gray-200 shadow-sm">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Дизайн React Flow Нод</h1>
        <p className="text-gray-600">Выберите один из 3 вариантов дизайна для ваших нод: Boss, Manager, Worker</p>
      </div>

      <Tabs value={variant} onValueChange={setVariant} className="flex-1 flex flex-col p-6">
        <TabsList className="mb-4 w-fit">
          <TabsTrigger value="variant1" className="px-6">
            Вариант 1: Тёмная Тема
          </TabsTrigger>
          <TabsTrigger value="variant2" className="px-6">
            Вариант 2: Светлая Тема
          </TabsTrigger>
          <TabsTrigger value="variant3" className="px-6">
            Вариант 3: Glassmorphism
          </TabsTrigger>
        </TabsList>

        <TabsContent value="variant1" className="flex-1 relative m-0 data-[state=active]:flex">
          <div className="absolute inset-0 bg-gradient-to-br from-purple-50 to-blue-50 rounded-lg overflow-hidden">
            <ReactFlow
              nodes={nodes1}
              edges={edges1}
              onNodesChange={onNodesChange1}
              onEdgesChange={onEdgesChange1}
              onConnect={onConnect1}
              nodeTypes={nodeTypes1}
              fitView
            >
              <Background />
              <Controls />
              <MiniMap 
                nodeColor={(node) => {
                  if (node.type === 'boss') return '#9333ea';
                  if (node.type === 'manager') return '#0ea5e9';
                  return '#10b981';
                }}
              />
            </ReactFlow>
          </div>
        </TabsContent>

        <TabsContent value="variant2" className="flex-1 relative m-0 data-[state=active]:flex">
          <div className="absolute inset-0 bg-gray-50 rounded-lg overflow-hidden">
            <ReactFlow
              nodes={nodes2}
              edges={edges2}
              onNodesChange={onNodesChange2}
              onEdgesChange={onEdgesChange2}
              onConnect={onConnect2}
              nodeTypes={nodeTypes2}
              fitView
            >
              <Background color="#e5e7eb" />
              <Controls />
              <MiniMap 
                nodeColor={(node) => {
                  if (node.type === 'boss') return '#f59e0b';
                  if (node.type === 'manager') return '#3b82f6';
                  return '#10b981';
                }}
              />
            </ReactFlow>
          </div>
        </TabsContent>

        <TabsContent value="variant3" className="flex-1 relative m-0 data-[state=active]:flex">
          <div className="absolute inset-0 bg-gradient-to-br from-pink-100 via-purple-100 to-cyan-100 rounded-lg overflow-hidden">
            <ReactFlow
              nodes={nodes3}
              edges={edges3}
              onNodesChange={onNodesChange3}
              onEdgesChange={onEdgesChange3}
              onConnect={onConnect3}
              nodeTypes={nodeTypes3}
              fitView
            >
              <Background color="#d8b4fe" />
              <Controls />
              <MiniMap 
                nodeColor={(node) => {
                  if (node.type === 'boss') return '#ec4899';
                  if (node.type === 'manager') return '#3b82f6';
                  return '#10b981';
                }}
              />
            </ReactFlow>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
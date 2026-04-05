import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

interface WorkerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  filesCount?: number;
}

function WorkerNodeComponent({ data }: NodeProps<{ data: WorkerNodeData }>) {
  const { role = 'Developer', status = 'pending', filesCount } = data;
  
  const statusIcons: Record<string, string> = {
    pending: '⏳',
    thinking: '💭',
    working: '⚙️',
    reviewing: '🔍',
    done: '✅',
    error: '❌',
  };

  const statusColors: Record<string, string> = {
    pending: 'border-gray-500',
    thinking: 'border-blue-500',
    working: 'border-green-500 animate-pulse',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  return (
    <div className={`bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 ${statusColors[status] || 'border-gray-500'} min-w-[180px]`}>
      <Handle type="target" position={Position.Top} />
      
      {/* Header */}
      <div className="bg-gradient-to-r from-green-500 to-green-600 text-white px-4 py-2 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-green-300 rounded-full"></div>
          <span className="font-bold text-sm">WORKER</span>
        </div>
        <span className="text-lg">{statusIcons[status] || '⏳'}</span>
      </div>
      
      {/* Body */}
      <div className="p-3 space-y-2">
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Роль:</span> {role}
        </div>
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Статус:</span> {status}
        </div>
        
        {filesCount !== undefined && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)]">Файлов:</span> {filesCount}
          </div>
        )}
      </div>
      
      <Handle type="source" position={Position.Bottom} id="source" />
    </div>
  );
}

export const WorkerNode = memo(WorkerNodeComponent);

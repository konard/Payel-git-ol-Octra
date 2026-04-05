import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

interface ManagerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  workerCount?: number;
  progress?: number;
}

function ManagerNodeComponent({ data }: NodeProps<{ data: ManagerNodeData }>) {
  const { role = 'Backend', status = 'pending', workerCount, progress = 0 } = data;
  
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
    working: 'border-blue-500 animate-pulse',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  const progressBars = Math.round(progress / 20);
  const progressStr = '█'.repeat(progressBars) + '░'.repeat(5 - progressBars);

  return (
    <div className={`bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 ${statusColors[status] || 'border-gray-500'} min-w-[200px]`}>
      <Handle type="target" position={Position.Top} />
      
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-500 to-blue-600 text-white px-4 py-2 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-blue-300 rounded-full"></div>
          <span className="font-bold text-sm">MANAGER</span>
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
        
        {workerCount !== undefined && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)]">Воркеров:</span> {workerCount}
          </div>
        )}
        
        {progress > 0 && (
          <div className="text-xs font-mono">
            <span className="text-[var(--text-muted)]">Прогресс:</span> 
            <span className="text-blue-500 ml-1">{progressStr}</span>
            <span className="ml-1">{progress}%</span>
          </div>
        )}
      </div>
      
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

export const ManagerNode = memo(ManagerNodeComponent);

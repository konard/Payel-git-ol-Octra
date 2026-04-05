import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

interface BossNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  workerCount?: number;
  techStack?: string[];
}

function BossNodeComponent({ data }: NodeProps<{ data: BossNodeData }>) {
  const { role = 'CEO', status = 'pending', workerCount, techStack } = data;
  
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
    working: 'border-orange-500 animate-pulse',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  return (
    <div className={`bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 ${statusColors[status] || 'border-gray-500'} min-w-[220px]`}>
      <Handle type="target" position={Position.Top} />
      
      {/* Header */}
      <div className="bg-gradient-to-r from-orange-500 to-orange-600 text-white px-4 py-2 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 bg-orange-300 rounded-full"></div>
          <span className="font-bold text-sm">BOSS</span>
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
        
        {techStack && techStack.length > 0 && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)]">Стек:</span> {techStack.join(', ')}
          </div>
        )}
        
        {workerCount !== undefined && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)]">Менеджеров:</span> {workerCount}
          </div>
        )}
      </div>
      
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

export const BossNode = memo(BossNodeComponent);

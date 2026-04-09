import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import { useI18n } from '../../../hooks/useI18n';
import managerImage from '../../../images/manager-image.png';

interface ManagerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  workerCount?: number;
  progress?: number;
  isConnected?: boolean;
}

function ManagerNodeComponent({ data }: NodeProps<{ data: ManagerNodeData }>) {
  const { role = 'Backend', status = 'pending', workerCount, progress = 0, isConnected = false } = data;
  const { t } = useI18n();
  
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
    <div className={`bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 ${
      isConnected ? (statusColors[status] || 'border-gray-500') : 'border-gray-600 opacity-60'
    } min-w-[200px] overflow-hidden`}>
      <Handle type="target" position={Position.Top} />

      {/* Header с изображением */}
      <div className="bg-gradient-to-r from-blue-500 to-blue-600 text-white px-4 py-3 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <img src={managerImage} alt="Manager" className="w-10 h-10 object-contain" />
          <span className="font-bold text-base">MANAGER</span>
        </div>
        <span className="text-xl">{statusIcons[status] || '⏳'}</span>
      </div>

      {/* Body */}
      <div className="p-3 space-y-2">
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Роль:</span> {role}
        </div>
        <div className="text-xs">
          <span className="text-[var(--text-muted)]">Статус:</span> {status}
        </div>

        {!isConnected && (
          <div className="text-xs text-orange-500 font-semibold">
            ⚠️ {t('contextMenu.notConnected')}
          </div>
        )}

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

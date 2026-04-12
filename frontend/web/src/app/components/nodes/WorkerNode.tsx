import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import { useI18n } from '../../../hooks/useI18n';
import workerImage from '../../../images/worker-image.png';

interface WorkerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  filesCount?: number;
  isConnected?: boolean;
}

const statusIcons: Record<string, string> = {
  pending: 'WAITING',
  thinking: 'THINKING',
  working: 'WORKING',
  reviewing: 'REVIEWING',
  done: 'DONE',
  error: 'ERROR',
};

function WorkerNodeComponent({ data }: NodeProps<{ data: WorkerNodeData }>) {
  const { role = 'Developer', status = 'pending', filesCount, isConnected = false } = data;
  const { t } = useI18n();

  const statusColors: Record<string, string> = {
    pending: 'border-gray-500',
    thinking: 'border-blue-500',
    working: 'border-green-500 animate-pulse',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  return (
    <div className={`bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 ${
      isConnected ? (statusColors[status] || 'border-gray-500') : 'border-gray-600 opacity-60'
    } min-w-[180px] overflow-hidden`}>
      <Handle type="target" position={Position.Top} />

      {/* Header с изображением */}
      <div className="bg-gradient-to-r from-green-500 to-green-600 text-white px-4 py-3 rounded-t-md flex items-center justify-between">
        <div className="flex items-center gap-3">
          <img src={workerImage} alt="Worker" className="w-10 h-10 object-contain" />
          <span className="font-bold text-base">WORKER</span>
        </div>
        <span className="text-xl">{statusIcons[status] || 'WAITING'}</span>
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
            {t('contextMenu.notConnected')}
          </div>
        )}

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

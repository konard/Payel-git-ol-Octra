import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import { Bot } from 'lucide-react';
import { useI18n } from '../../../hooks/useI18n';

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
    pending: 'WAITING',
    thinking: 'THINKING',
    working: 'WORKING',
    reviewing: 'REVIEWING',
    done: 'DONE',
    error: 'ERROR',
  };

  const statusColors: Record<string, string> = {
    pending: 'border-gray-500',
    thinking: 'border-blue-500',
    working: 'border-blue-500',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  const progressBars = Math.round(progress / 20);
  const progressStr = '█'.repeat(progressBars) + '░'.repeat(5 - progressBars);

  const getNodeClasses = () => {
    let classes = `bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 min-w-[200px] overflow-hidden backdrop-blur-sm`;

    if (!isConnected) {
      classes += ' border-gray-600 opacity-60';
    } else {
      classes += ` ${statusColors[status] || 'border-gray-500'}`;

      // Добавляем анимации в зависимости от статуса
      if (status === 'thinking') {
        classes += ' crewai-node--sending';
      } else if (status === 'working') {
        classes += ' crewai-node--working';
      }
    }

    return classes;
  };

  return (
    <div className={getNodeClasses()}>
      <Handle type="target" position={Position.Top} />

      {/* Header с иконкой */}
      <div className="bg-gradient-to-r from-blue-500 to-blue-600 text-white px-4 py-3 rounded-t-lg flex items-center">
        <div className="flex items-center gap-3">
          <div className="bg-white/10 p-2 rounded-xl backdrop-blur-sm">
            <Bot className="w-6 h-6 text-blue-200" />
          </div>
          <span className="font-bold text-base">MANAGER</span>
        </div>
      </div>

      {/* Body */}
      <div className="p-4 space-y-3">
        <div className="text-sm">
          <span className="text-[var(--text-muted)] font-medium">Роль:</span> <span className="text-[var(--text)]">{role}</span>
        </div>

        {!isConnected && (
          <div className="text-xs text-red-500 font-semibold bg-red-50 dark:bg-red-900/20 px-2 py-1 rounded">
            {t('contextMenu.notConnected')}
          </div>
        )}

        {workerCount !== undefined && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)] font-medium">Воркеров:</span> <span className="text-[var(--text)] ml-1">{workerCount}</span>
          </div>
        )}

        {progress > 0 && (
          <div className="text-xs font-mono bg-blue-50 dark:bg-blue-900/20 px-2 py-1 rounded">
            <span className="text-[var(--text-muted)] font-medium">Прогресс:</span>
            <span className="text-blue-600 dark:text-blue-400 ml-1">{progressStr}</span>
            <span className="ml-1 text-[var(--text)]">{progress}%</span>
          </div>
        )}
      </div>

      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

export const ManagerNode = memo(ManagerNodeComponent);

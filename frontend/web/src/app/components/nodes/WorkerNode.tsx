import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import { Cpu } from 'lucide-react';
import { useI18n } from '../../../hooks/useI18n';
import { useNodeResize } from '../../../hooks/useNodeResize';

interface WorkerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  filesCount?: number;
  isConnected?: boolean;
  scale?: number;
}

function WorkerNodeComponent({ id, data }: NodeProps<{ data: WorkerNodeData }>) {
  const { role = 'Developer', status = 'pending', filesCount, isConnected = false, scale = 1 } = data;
  const { t } = useI18n();
  const { scale: currentScale, handleResize } = useNodeResize(id, scale);

  const statusColors: Record<string, string> = {
    pending: 'border-gray-500',
    thinking: 'border-blue-500',
    working: 'border-green-500',
    reviewing: 'border-purple-500',
    done: 'border-green-500',
    error: 'border-red-500',
  };

  const getNodeClasses = () => {
    let classes = `bg-[var(--bg-node)] text-[var(--text-node)] rounded-lg shadow-lg border-2 min-w-[180px] overflow-hidden backdrop-blur-sm`;

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
    <div className={`${getNodeClasses()} relative`} style={{ transform: `scale(${currentScale})`, transformOrigin: 'center center' }}>
      <Handle type="target" position={Position.Top} />

      {/* Header с иконкой */}
      <div className="bg-gradient-to-r from-green-500 to-green-600 text-white px-4 py-3 rounded-t-lg flex items-center">
        <div className="flex items-center gap-3">
          <div className="bg-white/10 p-2 rounded-xl backdrop-blur-sm">
            <Cpu className="w-6 h-6 text-green-200" />
          </div>
          <span className="font-bold text-base">WORKER</span>
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

        {filesCount !== undefined && (
          <div className="text-xs">
            <span className="text-[var(--text-muted)] font-medium">Файлов:</span> <span className="text-[var(--text)] ml-1">{filesCount}</span>
          </div>
        )}
      </div>

      {/* Resize handle */}
      <div 
        className="absolute bottom-0 right-0 w-4 h-4 cursor-se-resize opacity-50 hover:opacity-100"
        onMouseDown={handleResize}
      >
        <svg viewBox="0 0 24 24" fill="currentColor" className="w-4 h-4 text-white">
          <path d="M22 22H20V20H22V22ZM22 18H20V16H22V18ZM18 22H16V20H18V22ZM22 14H20V12H22V14ZM18 18H16V16H18V18ZM14 22H12V20H14V22Z"/>
        </svg>
      </div>

      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}

export const WorkerNode = memo(WorkerNodeComponent);

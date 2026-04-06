import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { FolderArchive, CheckCircle, XCircle, Loader2 } from 'lucide-react';
import type { AgentNodeStatus } from '../../../stores/taskStore';

interface ZIPArchiveNodeProps {
  data: {
    role?: string;
    status?: AgentNodeStatus;
    fileName?: string;
    fileSize?: string;
    filesCount?: number;
  };
}

const statusConfig: Record<string, { border: string; bg: string; icon: React.ReactNode; label: string }> = {
  pending: {
    border: 'border-gray-400',
    bg: 'bg-gray-100 dark:bg-gray-800',
    icon: <Loader2 size={16} className="text-gray-400" />,
    label: 'Ожидание',
  },
  working: {
    border: 'border-orange-500',
    bg: 'bg-orange-50 dark:bg-orange-950/30',
    icon: <Loader2 size={16} className="text-orange-500 animate-spin" />,
    label: 'Сборка...',
  },
  done: {
    border: 'border-green-500',
    bg: 'bg-green-50 dark:bg-green-950/30',
    icon: <CheckCircle size={16} className="text-green-500" />,
    label: 'Готово',
  },
  error: {
    border: 'border-red-500',
    bg: 'bg-red-50 dark:bg-red-950/30',
    icon: <XCircle size={16} className="text-red-500" />,
    label: 'Ошибка',
  },
};

function ZIPArchiveNodeComponent({ data }: ZIPArchiveNodeProps) {
  const status = data.status || 'pending';
  const config = statusConfig[status] || statusConfig.pending;
  const fileName = data.fileName || 'project.zip';
  const fileSize = data.fileSize || '';
  const filesCount = data.filesCount;

  return (
    <div className={`rounded-xl border-2 ${config.border} shadow-lg min-w-[180px]`}>
      {/* Header */}
      <div className={`${config.bg} px-3 py-2 rounded-t-lg border-b ${config.border}`}>
        <div className="flex items-center gap-2">
          <FolderArchive size={18} className="text-[var(--text)]" />
          <span className="text-sm font-semibold text-[var(--text)]">ZIP Archive</span>
        </div>
      </div>

      {/* Body */}
      <div className="px-3 py-2 bg-[var(--surface)] space-y-1.5">
        {/* Status */}
        <div className="flex items-center gap-1.5 text-xs">
          {config.icon}
          <span className="text-[var(--text-muted)]">{config.label}</span>
        </div>

        {/* File name */}
        <div className="text-xs font-mono text-[var(--text)] truncate">
          📦 {fileName}
        </div>

        {/* File size */}
        {fileSize && (
          <div className="text-xs text-[var(--text-muted)]">
            Размер: {fileSize}
          </div>
        )}

        {/* Files count */}
        {filesCount !== undefined && (
          <div className="text-xs text-[var(--text-muted)]">
            Файлов: {filesCount}
          </div>
        )}
      </div>

      {/* Input handles (from managers/workers) */}
      <Handle type="target" position={Position.Top} id="input-0" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '20%' }} />
      <Handle type="target" position={Position.Top} id="input-1" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '40%' }} />
      <Handle type="target" position={Position.Top} id="input-2" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '60%' }} />
      <Handle type="target" position={Position.Top} id="input-3" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '80%' }} />
    </div>
  );
}

export const ZIPArchiveNode = memo(ZIPArchiveNodeComponent);

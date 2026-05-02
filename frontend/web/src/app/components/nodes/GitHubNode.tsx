import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { CheckCircle, XCircle, Loader2, Github } from 'lucide-react';
import type { AgentNodeStatus } from '../../../stores/taskStore';
import { t } from '../../../hooks/useI18n';
import { useNodeResize } from '../../../hooks/useNodeResize';
import githubImage from '../../../images/github-image.png';

interface GitHubNodeProps {
  id: string;
  data: {
    role?: string;
    status?: AgentNodeStatus;
    repoUrl?: string;
    commitCount?: number;
    scale?: number;
  };
}

const statusConfig: Record<string, { border: string; bg: string; icon: React.ReactNode; label: string }> = {
  pending: {
    border: 'border-gray-400',
    bg: 'bg-gray-100 dark:bg-gray-800',
    icon: <Loader2 size={16} className="text-gray-400" />,
    label: t('nodes.pending'),
  },
  working: {
    border: 'border-orange-500',
    bg: 'bg-orange-50 dark:bg-orange-950/30',
    icon: <Loader2 size={16} className="text-orange-500 animate-spin" />,
    label: t('nodes.building'),
  },
  done: {
    border: 'border-green-500',
    bg: 'bg-green-50 dark:bg-green-950/30',
    icon: <CheckCircle size={16} className="text-green-500" />,
    label: t('nodes.ready'),
  },
  error: {
    border: 'border-red-500',
    bg: 'bg-red-50 dark:bg-red-950/30',
    icon: <XCircle size={16} className="text-red-500" />,
    label: t('nodes.error'),
  },
};

function GitHubNodeComponent({ id, data }: GitHubNodeProps) {
  const status = data.status || 'pending';
  const config = statusConfig[status] || statusConfig.pending;
  const repoUrl = data.repoUrl || '';
  const commitCount = data.commitCount;
  const { scale: currentScale, handleResize } = useNodeResize(id, data.scale || 1);

  return (
    <div className={`rounded-xl border-2 ${config.border} shadow-lg min-w-[200px] relative`} style={{ transform: `scale(${currentScale})`, transformOrigin: 'center center' }}>
      {/* Header */}
      <div className={`${config.bg} px-3 py-2 rounded-t-lg border-b ${config.border}`}>
        <div className="flex items-center gap-2">
          <img src={githubImage} alt="GitHub" className="w-5 h-5" />
          <span className="text-sm font-semibold text-[var(--text)]">GitHub</span>
        </div>
      </div>

      {/* Body */}
      <div className="px-3 py-2 bg-[var(--surface)] space-y-1.5">
        {/* Status */}
        <div className="flex items-center gap-1.5 text-xs">
          {config.icon}
          <span className="text-[var(--text-muted)]">{config.label}</span>
        </div>

        {/* Repo URL */}
        {repoUrl && (
          <div className="text-xs font-mono text-[var(--text)] truncate">
            <Github size={12} className="inline mr-1" />
            {repoUrl}
          </div>
        )}

        {/* Commit count */}
        {commitCount !== undefined && (
          <div className="text-xs text-[var(--text-muted)]">
            Commits: {commitCount}
          </div>
        )}
      </div>

      {/* Input handles (from managers/workers) */}
      <Handle type="target" position={Position.Top} id="input-0" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '20%' }} />
      <Handle type="target" position={Position.Top} id="input-1" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '40%' }} />
      <Handle type="target" position={Position.Top} id="input-2" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '60%' }} />
      <Handle type="target" position={Position.Top} id="input-3" className="!bg-blue-500 !w-2 !h-2" style={{ top: -4, left: '80%' }} />

      {/* Resize handle */}
      <div 
        className="absolute bottom-0 right-0 w-4 h-4 cursor-se-resize opacity-50 hover:opacity-100"
        onMouseDown={handleResize}
      >
        <svg viewBox="0 0 24 24" fill="currentColor" className="w-4 h-4 text-[var(--text)]">
          <path d="M22 22H20V20H22V22ZM22 18H20V16H22V18ZM18 22H16V20H18V22ZM22 14H20V12H22V14ZM18 18H16V16H18V18ZM14 22H12V20H14V22Z"/>
        </svg>
      </div>
    </div>
  );
}

export const GitHubNode = memo(GitHubNodeComponent);
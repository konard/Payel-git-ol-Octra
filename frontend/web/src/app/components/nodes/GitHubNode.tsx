import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { CheckCircle, XCircle, Loader2 } from 'lucide-react';
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
    prUrl?: string;
    commitCount?: number;
    scale?: number;
  };
}

// GitHub "sink" node — neutral accent so it reads as the output target.
const GITHUB_ACCENT = '#6e7681';

const statusConfig: Record<string, { dot: string; icon: React.ReactNode; label: string }> = {
  pending: { dot: '#9ca3af', icon: <Loader2 size={14} className="text-gray-400" />, label: t('nodes.pending') },
  working: { dot: '#f97316', icon: <Loader2 size={14} className="text-orange-500 animate-spin" />, label: t('nodes.building') },
  done: { dot: '#22c55e', icon: <CheckCircle size={14} className="text-green-500" />, label: t('nodes.ready') },
  error: { dot: '#ef4444', icon: <XCircle size={14} className="text-red-500" />, label: t('nodes.error') },
};

function GitHubNodeComponent({ id, data }: GitHubNodeProps) {
  const status = data.status || 'pending';
  const config = statusConfig[status] || statusConfig.pending;
  const repoUrl = data.repoUrl || '';
  const prUrl = data.prUrl || '';
  const commitCount = data.commitCount;
  const { scale: currentScale, handleResize } = useNodeResize(id, data.scale || 1);
  const isActive = status === 'working';
  const hasLink = !!(repoUrl || prUrl);

  return (
    <div
      className={`agent-node group relative ${isActive ? 'agent-node--active' : ''} ${hasLink ? 'cursor-pointer' : ''}`}
      style={{
        minWidth: 200,
        transform: `scale(${currentScale})`,
        transformOrigin: 'center center',
        ['--node-accent' as any]: GITHUB_ACCENT,
      }}
      title={hasLink ? (prUrl ? t('nodes.openPr') : t('nodes.openGithub')) : undefined}
    >
      {/* Input handles (from managers/workers) */}
      <Handle type="target" position={Position.Top} id="input-0" className="agent-node__handle" style={{ top: -4, left: '20%' }} />
      <Handle type="target" position={Position.Top} id="input-1" className="agent-node__handle" style={{ top: -4, left: '40%' }} />
      <Handle type="target" position={Position.Top} id="input-2" className="agent-node__handle" style={{ top: -4, left: '60%' }} />
      <Handle type="target" position={Position.Top} id="input-3" className="agent-node__handle" style={{ top: -4, left: '80%' }} />

      {/* Header */}
      <div className="agent-node__header">
        <div className="agent-node__badge" style={{ backgroundColor: GITHUB_ACCENT }}>
          <img src={githubImage} alt="GitHub" className="w-4 h-4" />
        </div>
        <div className="agent-node__title">GITHUB</div>
        <span
          className={`agent-node__status-dot ${isActive ? 'agent-node__status-dot--pulse' : ''}`}
          style={{ backgroundColor: config.dot, boxShadow: `0 0 6px ${config.dot}` }}
          title={status}
        />
      </div>

      {/* Body */}
      <div className="agent-node__body">
        <div className="flex items-center gap-1.5 text-xs text-[var(--text-muted)]">
          {config.icon}
          <span>{config.label}</span>
        </div>

        {repoUrl && (
          <div className="text-xs font-mono text-[var(--text)] truncate" title={prUrl ? 'Pull Request' : 'Repository'}>
            {prUrl || repoUrl}
          </div>
        )}

        {commitCount !== undefined && (
          <div className="agent-node__meta">
            <span className="agent-node__meta-label">Commits:</span>
            <span className="agent-node__meta-value">{commitCount}</span>
          </div>
        )}
      </div>

      {/* Resize handle */}
      <div className="agent-node__resize" onMouseDown={handleResize} title="Resize">
        <svg viewBox="0 0 24 24" fill="currentColor" className="w-3.5 h-3.5">
          <path d="M22 22H20V20H22V22ZM22 18H20V16H22V18ZM18 22H16V20H18V22ZM22 14H20V12H22V14ZM18 18H16V16H18V18ZM14 22H12V20H14V22Z" />
        </svg>
      </div>
    </div>
  );
}

export const GitHubNode = memo(GitHubNodeComponent);

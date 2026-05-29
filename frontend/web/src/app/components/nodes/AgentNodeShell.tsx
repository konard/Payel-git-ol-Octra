import type { ReactNode } from 'react';
import { Handle, Position } from '@xyflow/react';
import { useI18n } from '../../../hooks/useI18n';
import { useNodeResize } from '../../../hooks/useNodeResize';
import type { AgentNodeStatus } from '../../../stores/taskStore';

interface AgentNodeShellProps {
  id: string;
  /** Accent color (hex) used for the icon badge, accent strip and active glow. */
  accent: string;
  icon: ReactNode;
  /** Short type label rendered in the header, e.g. BOSS / MANAGER / WORKER. */
  typeLabel: string;
  role: string;
  status: AgentNodeStatus;
  isConnected: boolean;
  scale: number;
  minWidth?: number;
  /** Extra rows rendered in the node body (files count, progress, …). */
  children?: ReactNode;
  /** Render top target handle (false for source-only roots like Boss). */
  withTargetHandle?: boolean;
  /** Render bottom source handle (false for sink nodes). */
  withSourceHandle?: boolean;
}

const statusDotColors: Record<AgentNodeStatus, string> = {
  pending: '#9ca3af',
  thinking: '#3b82f6',
  working: '#22c55e',
  reviewing: '#a855f7',
  done: '#22c55e',
  error: '#ef4444',
};

/**
 * Shared n8n-inspired node container. Replaces the previous bright
 * full-width gradient header with a clean surface card, a small colored
 * icon badge and a left accent strip. Active states get a soft glow ring.
 */
export function AgentNodeShell({
  id,
  accent,
  icon,
  typeLabel,
  role,
  status,
  isConnected,
  scale,
  minWidth = 200,
  children,
  withTargetHandle = true,
  withSourceHandle = true,
}: AgentNodeShellProps) {
  const { t } = useI18n();
  const { scale: currentScale, handleResize } = useNodeResize(id, scale);

  const isActive = isConnected && (status === 'thinking' || status === 'working' || status === 'reviewing');
  const dotColor = statusDotColors[status] || statusDotColors.pending;

  return (
    <div
      className={`agent-node group relative ${isActive ? 'agent-node--active' : ''} ${!isConnected ? 'agent-node--disconnected' : ''}`}
      style={{
        minWidth,
        transform: `scale(${currentScale})`,
        transformOrigin: 'center center',
        // CSS custom props consumed by .agent-node styles
        ['--node-accent' as any]: accent,
      }}
    >
      {withTargetHandle && <Handle type="target" position={Position.Top} className="agent-node__handle" />}

      {/* Header */}
      <div className="agent-node__header">
        <div className="agent-node__badge" style={{ backgroundColor: accent }}>
          {icon}
        </div>
        <div className="agent-node__title">{typeLabel}</div>
        <span
          className={`agent-node__status-dot ${isActive ? 'agent-node__status-dot--pulse' : ''}`}
          style={{ backgroundColor: dotColor, boxShadow: `0 0 6px ${dotColor}` }}
          title={status}
        />
      </div>

      {/* Body */}
      <div className="agent-node__body">
        <div className="agent-node__role">
          <span className="agent-node__role-label">{t('nodes.role')}</span>
          <span className="agent-node__role-value">{role}</span>
        </div>

        {!isConnected && (
          <div className="agent-node__warning">{t('contextMenu.notConnected')}</div>
        )}

        {children}
      </div>

      {/* Resize handle */}
      <div
        className="agent-node__resize"
        onMouseDown={handleResize}
        title="Resize"
      >
        <svg viewBox="0 0 24 24" fill="currentColor" className="w-3.5 h-3.5">
          <path d="M22 22H20V20H22V22ZM22 18H20V16H22V18ZM18 22H16V20H18V22ZM22 14H20V12H22V14ZM18 18H16V16H18V18ZM14 22H12V20H14V22Z" />
        </svg>
      </div>

      {withSourceHandle && <Handle type="source" position={Position.Bottom} className="agent-node__handle" />}
    </div>
  );
}

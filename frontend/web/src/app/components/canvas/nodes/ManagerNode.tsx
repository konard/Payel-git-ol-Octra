import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';
import { Bot } from 'lucide-react';
import { useI18n } from '../../../../hooks/useI18n';
import { AgentNodeShell } from './AgentNodeShell';

interface ManagerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  workerCount?: number;
  progress?: number;
  isConnected?: boolean;
  scale?: number;
}

const MANAGER_ACCENT = '#7c4dff';

function ManagerNodeComponent({ id, data }: NodeProps<{ data: ManagerNodeData }>) {
  const { role = 'Coordinator', status = 'pending', workerCount, progress = 0, isConnected = false, scale = 1 } = data;
  const { t } = useI18n();

  return (
    <AgentNodeShell
      id={id}
      accent={MANAGER_ACCENT}
      icon={<Bot className="w-4 h-4 text-white" />}
      typeLabel="MANAGER"
      role={role}
      status={status}
      isConnected={isConnected}
      scale={scale}
      minWidth={200}
    >
      {workerCount !== undefined && (
        <div className="agent-node__meta">
          <span className="agent-node__meta-label">{t('nodes.workers')}</span>
          <span className="agent-node__meta-value">{workerCount}</span>
        </div>
      )}
      {progress > 0 && (
        <div className="agent-node__progress">
          <div className="agent-node__progress-track">
            <div
              className="agent-node__progress-fill"
              style={{ width: `${Math.max(0, Math.min(100, progress))}%`, backgroundColor: MANAGER_ACCENT }}
            />
          </div>
          <span className="agent-node__progress-value">{progress}%</span>
        </div>
      )}
    </AgentNodeShell>
  );
}

export const ManagerNode = memo(ManagerNodeComponent);

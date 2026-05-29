import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';
import { Cpu } from 'lucide-react';
import { useI18n } from '../../../hooks/useI18n';
import { AgentNodeShell } from './AgentNodeShell';

interface WorkerNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  filesCount?: number;
  isConnected?: boolean;
  scale?: number;
}

const WORKER_ACCENT = '#22c55e';

function WorkerNodeComponent({ id, data }: NodeProps<{ data: WorkerNodeData }>) {
  const { role = 'Developer', status = 'pending', filesCount, isConnected = false, scale = 1 } = data;
  const { t } = useI18n();

  return (
    <AgentNodeShell
      id={id}
      accent={WORKER_ACCENT}
      icon={<Cpu className="w-4 h-4 text-white" />}
      typeLabel="WORKER"
      role={role}
      status={status}
      isConnected={isConnected}
      scale={scale}
      minWidth={180}
    >
      {filesCount !== undefined && (
        <div className="agent-node__meta">
          <span className="agent-node__meta-label">{t('nodes.filesCount')}</span>
          <span className="agent-node__meta-value">{filesCount}</span>
        </div>
      )}
    </AgentNodeShell>
  );
}

export const WorkerNode = memo(WorkerNodeComponent);

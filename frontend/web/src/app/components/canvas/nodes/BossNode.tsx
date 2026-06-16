import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';
import { Brain } from 'lucide-react';
import { useI18n } from '../../../../hooks/useI18n';
import { AgentNodeShell } from './AgentNodeShell';

interface BossNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  workerCount?: number;
  techStack?: string[];
  isConnected?: boolean;
  scale?: number;
}

const BOSS_ACCENT = '#f97316';

function BossNodeComponent({ id, data }: NodeProps<{ data: BossNodeData }>) {
  const { role = 'CEO', status = 'pending', workerCount, techStack, isConnected = false, scale = 1 } = data;
  const { t } = useI18n();

  return (
    <AgentNodeShell
      id={id}
      accent={BOSS_ACCENT}
      icon={<Brain className="w-4 h-4 text-white" />}
      typeLabel="BOSS"
      role={role}
      status={status}
      isConnected={isConnected}
      scale={scale}
      minWidth={220}
    >
      {Array.isArray(techStack) && techStack.length > 0 && (
        <div className="agent-node__meta">
          <span className="agent-node__meta-label">{t('nodes.stack')}</span>
          <span className="agent-node__meta-value">{techStack.join(', ')}</span>
        </div>
      )}
      {workerCount !== undefined && (
        <div className="agent-node__meta">
          <span className="agent-node__meta-label">{t('nodes.managers')}</span>
          <span className="agent-node__meta-value">{workerCount}</span>
        </div>
      )}
    </AgentNodeShell>
  );
}

export const BossNode = memo(BossNodeComponent);

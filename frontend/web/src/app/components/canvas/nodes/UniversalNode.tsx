import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';
import { Sparkles } from 'lucide-react';
import { useI18n } from '../../../../hooks/useI18n';
import { AgentNodeShell } from './AgentNodeShell';

interface UniversalNodeData {
  role?: string;
  status?: 'pending' | 'thinking' | 'working' | 'reviewing' | 'done' | 'error';
  filesCount?: number;
  isConnected?: boolean;
  scale?: number;
}

const UNIVERSAL_ACCENT = '#ffffff';

function UniversalNodeComponent({ id, data }: NodeProps<{ data: UniversalNodeData }>) {
  const { role = 'Universal', status = 'pending', filesCount, isConnected = false, scale = 1 } = data;
  const { t } = useI18n();

  return (
    <AgentNodeShell
      id={id}
      accent={UNIVERSAL_ACCENT}
      icon={<Sparkles className="w-4 h-4 text-neutral-950" />}
      typeLabel="UNIVERSAL"
      role={role}
      status={status}
      isConnected={true}
      scale={scale}
      minWidth={200}
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

export const UniversalNode = memo(UniversalNodeComponent);

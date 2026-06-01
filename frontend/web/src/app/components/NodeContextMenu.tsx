import { useEffect, useRef, useState } from 'react';
import { Brain, Bot, Cpu, Archive, Zap, Copy, Trash2 } from 'lucide-react';
import { useTaskStore, type AgentNodeType } from '../../stores/taskStore';
import { useIntegrationStore } from '../../stores/integrationStore';
import { n8nService, type N8nWorkflow } from '../../services/n8nService';
import { useI18n } from '../../hooks/useI18n';


const nodeIcons: Record<AgentNodeType, React.ComponentType<{ className?: string }>> = {
  boss: Brain,
  manager: Bot,
  worker: Cpu,
  github: Archive,
};

// Accent colors mirror the redesigned nodes so the menu reads as the same object.
const nodeAccents: Record<AgentNodeType, string> = {
  boss: '#f97316',
  manager: '#7c4dff',
  worker: '#22c55e',
  github: '#6e7681',
};

interface ContextMenuProps {
  x: number;
  y: number;
  nodeId: string;
  nodeType: AgentNodeType;
  nodeRole: string;
  onClose: () => void;
}

const roleOptions: Record<AgentNodeType, string[]> = {
  boss: ['CEO', 'CTO', 'Technical Director', 'Architect'],
  manager: ['Coordinator', 'Research Lead', 'Content Lead', 'Backend', 'Frontend', 'QA Lead'],
  worker: ['Specialist', 'Researcher', 'Designer', 'Writer', 'Developer', 'Tester'],
  github: ['GitHub'],
};

const SCALE_OPTIONS: { label: string; value: number }[] = [
  { label: 'S', value: 0.75 },
  { label: 'M', value: 1 },
  { label: 'L', value: 1.25 },
  { label: 'XL', value: 1.5 },
];

export function NodeContextMenu({ x, y, nodeId, nodeType, nodeRole, onClose }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const updateNode = useTaskStore((state) => state.updateNode);
  const removeNode = useTaskStore((state) => state.removeNode);
  const n8nIntegration = useIntegrationStore((state) => state.integrations.n8n);
  const { t } = useI18n();
  const currentNode = useTaskStore((state) => state.nodes.find((n) => n.id === nodeId));
  const [position, setPosition] = useState({ left: x, top: y });
  const [n8nTrigger, setN8nTrigger] = useState<'start' | 'end' | 'middle' | 'custom' | null>(currentNode?.n8nTrigger || null);
  const [customPercentage, setCustomPercentage] = useState(currentNode?.n8nPercentage || 50);
  const [n8nWorkflows, setN8nWorkflows] = useState<N8nWorkflow[]>([]);
  const [loadingWorkflows, setLoadingWorkflows] = useState(false);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string>(currentNode?.n8nWorkflowId || '');
  const [webhookUrl, setWebhookUrl] = useState<string>(currentNode?.n8nWebhookUrl || '');
  const [webhookUrlError, setWebhookUrlError] = useState<string>('');

  const accent = nodeAccents[nodeType] || nodeAccents.worker;
  const IconComponent = nodeIcons[nodeType];

  // Close on outside click
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [onClose]);

  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [onClose]);

  // Keep the menu inside the viewport
  useEffect(() => {
    if (!menuRef.current) return;

    const menuWidth = menuRef.current.offsetWidth;
    const menuHeight = menuRef.current.offsetHeight;
    const padding = 10;

    let newLeft = x;
    let newTop = y;

    if (x + menuWidth > window.innerWidth - padding) {
      newLeft = window.innerWidth - menuWidth - padding;
    }
    if (y + menuHeight > window.innerHeight - padding) {
      newTop = window.innerHeight - menuHeight - padding;
    }

    newLeft = Math.max(padding, newLeft);
    newTop = Math.max(padding, newTop);

    setPosition({ left: newLeft, top: newTop });
  }, [x, y]);

  // Load n8n workflows when the menu opens
  useEffect(() => {
    const loadWorkflows = async () => {
      if (n8nIntegration.connected && n8nIntegration.config.apiKey) {
        setLoadingWorkflows(true);
        try {
          const workflows = await n8nService.getWorkflows(n8nIntegration.config);
          setN8nWorkflows(workflows);
        } catch (error) {
          console.error('Failed to load n8n workflows:', error);
        } finally {
          setLoadingWorkflows(false);
        }
      }
    };

    loadWorkflows();
  }, [n8nIntegration]);

  const handleRoleChange = (newRole: string) => {
    updateNode(nodeId, { role: newRole });
    onClose();
  };

  const handleDelete = () => {
    removeNode(nodeId);
    onClose();
  };

  const handleScaleChange = (newScale: number) => {
    updateNode(nodeId, { scale: newScale });
  };

  const handleDuplicate = () => {
    const addNode = useTaskStore.getState().addNode;
    const node = useTaskStore.getState().nodes.find((n) => n.id === nodeId);
    if (node) {
      addNode({
        id: `node-${Date.now()}`,
        type: node.type,
        role: node.role,
        status: node.status,
        progress: 0,
        position: {
          x: (node.position?.x || 0) + 50,
          y: (node.position?.y || 0) + 50,
        },
      });
    }
    onClose();
  };

  const handleN8nTriggerChange = (trigger: 'start' | 'end' | 'middle' | 'custom') => {
    setN8nTrigger(trigger);
    const updates: any = { n8nTrigger: trigger };
    if (trigger === 'custom') {
      updates.n8nPercentage = customPercentage;
    } else {
      updates.n8nPercentage = undefined;
    }
    updateNode(nodeId, updates);
  };

  const handleWorkflowSelect = (workflowId: string) => {
    setSelectedWorkflowId(workflowId);
    updateNode(nodeId, { n8nWorkflowId: workflowId });
  };

  const validateUrl = (url: string): boolean => {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  };

  const handleWebhookUrlChange = (url: string) => {
    setWebhookUrl(url);
    if (url && !validateUrl(url)) {
      setWebhookUrlError('Invalid URL format');
    } else {
      setWebhookUrlError('');
      updateNode(nodeId, { n8nWebhookUrl: url });
    }
  };

  const handleCustomPercentageChange = (value: number) => {
    setCustomPercentage(value);
    if (n8nTrigger === 'custom') {
      updateNode(nodeId, { n8nPercentage: value });
    }
  };

  const triggerOptions: { key: 'start' | 'middle' | 'end' | 'custom'; label: string }[] = [
    { key: 'start', label: t('contextMenu.n8nAutomation.atStart') },
    { key: 'middle', label: t('contextMenu.n8nAutomation.atMiddle') },
    { key: 'end', label: t('contextMenu.n8nAutomation.atEnd') },
    { key: 'custom', label: t('contextMenu.n8nAutomation.atCustom', { percentage: customPercentage } as any) },
  ];

  return (
    <div
      ref={menuRef}
      className="fixed z-[100] bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-[248px] overflow-hidden backdrop-blur-sm"
      style={{ left: position.left, top: position.top, ['--node-accent' as any]: accent }}
    >
      {/* Accent strip */}
      <div className="h-1" style={{ backgroundColor: accent }} />

      {/* Header */}
      <div className="px-3 py-3 flex items-center gap-3 border-b border-[var(--border)]">
        <div
          className="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
          style={{ backgroundColor: accent, boxShadow: `0 2px 8px ${accent}66` }}
        >
          {IconComponent && <IconComponent className="w-5 h-5 text-white" />}
        </div>
        <div className="min-w-0">
          <div className="text-[11px] font-bold tracking-wider text-[var(--text-muted)]">
            {nodeType.toUpperCase()}
          </div>
          <div className="text-sm font-semibold text-[var(--text)] truncate">{nodeRole}</div>
        </div>
      </div>

      {/* Role */}
      <div className="p-2 border-b border-[var(--border)]">
        <div className="text-[11px] font-medium uppercase tracking-wide text-[var(--text-muted)] mb-1.5 px-2">
          {t('contextMenu.changeRole')}
        </div>
        <div className="grid grid-cols-2 gap-1">
          {roleOptions[nodeType]?.map((role) => {
            const active = role === nodeRole;
            return (
              <button
                key={role}
                onClick={() => handleRoleChange(role)}
                className="text-left px-2.5 py-1.5 text-xs rounded-lg transition-colors truncate"
                style={
                  active
                    ? { backgroundColor: accent, color: '#fff' }
                    : undefined
                }
                onMouseEnter={(e) => {
                  if (!active) e.currentTarget.style.backgroundColor = 'var(--background)';
                }}
                onMouseLeave={(e) => {
                  if (!active) e.currentTarget.style.backgroundColor = '';
                }}
              >
                {role}
              </button>
            );
          })}
        </div>
      </div>

      {/* N8n integration */}
      {n8nIntegration.connected && (
        <div className="p-2 border-b border-[var(--border)]">
          <div className="text-[11px] font-medium uppercase tracking-wide text-[var(--text-muted)] mb-2 px-2 flex items-center gap-1.5">
            <Zap className="w-3 h-3" />
            {t('contextMenu.n8nAutomation.title')}
          </div>

          <div className="px-2 mb-2">
            <select
              value={selectedWorkflowId}
              onChange={(e) => handleWorkflowSelect(e.target.value)}
              className="w-full px-2 py-1.5 text-xs bg-[var(--background)] border border-[var(--border)] rounded-lg text-[var(--text)]"
            >
              <option value="">{t('contextMenu.n8nAutomation.selectWorkflow')}</option>
              {loadingWorkflows ? (
                <option disabled>{t('contextMenu.n8nAutomation.loading')}</option>
              ) : (
                n8nWorkflows.map((workflow) => (
                  <option key={workflow.id} value={workflow.id}>
                    {workflow.name}
                  </option>
                ))
              )}
            </select>

            <input
              type="url"
              value={webhookUrl}
              onChange={(e) => handleWebhookUrlChange(e.target.value)}
              placeholder={t('contextMenu.n8nAutomation.orWebhookUrl')}
              className={`w-full mt-2 px-2 py-1.5 text-xs bg-[var(--background)] border rounded-lg text-[var(--text)] placeholder-[var(--text-muted)] ${
                webhookUrlError ? 'border-red-500' : 'border-[var(--border)]'
              }`}
            />
            {webhookUrlError && <div className="text-[11px] text-red-500 mt-1">{webhookUrlError}</div>}
          </div>

          <div className="px-2">
            <div className="text-[11px] text-[var(--text-muted)] mb-1.5">{t('contextMenu.n8nAutomation.triggerWhen')}</div>
            <div className="grid grid-cols-2 gap-1">
              {triggerOptions.map((opt) => {
                const active = n8nTrigger === opt.key;
                return (
                  <button
                    key={opt.key}
                    onClick={() => handleN8nTriggerChange(opt.key)}
                    className={`px-2 py-1.5 text-[11px] rounded-lg transition-colors text-left ${
                      active
                        ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400 font-medium'
                        : 'hover:bg-[var(--background)] text-[var(--text)]'
                    }`}
                  >
                    {opt.label}
                  </button>
                );
              })}
            </div>
            {n8nTrigger === 'custom' && (
              <input
                type="range"
                min="0"
                max="100"
                value={customPercentage}
                onChange={(e) => handleCustomPercentageChange(Number(e.target.value))}
                className="w-full mt-2 accent-[var(--accent)]"
              />
            )}
          </div>
        </div>
      )}

      {/* Scale */}
      <div className="p-2 border-b border-[var(--border)]">
        <div className="text-[11px] font-medium uppercase tracking-wide text-[var(--text-muted)] mb-1.5 px-2">
          {t('contextMenu.scale')}
        </div>
        <div className="flex items-center gap-1 px-2 p-1 bg-[var(--background)] rounded-lg">
          {SCALE_OPTIONS.map((opt) => {
            const active = (currentNode?.scale || 1) === opt.value;
            return (
              <button
                key={opt.label}
                onClick={() => handleScaleChange(opt.value)}
                className="flex-1 py-1 text-xs font-medium rounded-md transition-colors"
                style={active ? { backgroundColor: accent, color: '#fff' } : { color: 'var(--text-muted)' }}
              >
                {opt.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* Actions */}
      <div className="p-2 space-y-0.5">
        <button
          onClick={handleDuplicate}
          className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-[var(--background)] text-[var(--text)] transition-colors flex items-center gap-2.5"
        >
          <Copy className="w-4 h-4 text-[var(--text-muted)]" />
          {t('contextMenu.duplicate')}
        </button>
        <button
          onClick={handleDelete}
          className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-red-500/10 text-red-500 transition-colors flex items-center gap-2.5"
        >
          <Trash2 className="w-4 h-4" />
          {t('contextMenu.delete')}
        </button>
      </div>
    </div>
  );
}

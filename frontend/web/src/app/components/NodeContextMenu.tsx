import { useEffect, useRef, useState } from 'react';
import { Brain, Bot, Cpu, Archive, Zap } from 'lucide-react';
import { useTaskStore, type AgentNodeType } from '../../stores/taskStore';
import { useIntegrationStore } from '../../stores/integrationStore';
import { n8nService, type N8nWorkflow } from '../../services/n8nService';
import { useI18n } from '../../hooks/useI18n';


const nodeIcons: Record<AgentNodeType, React.ComponentType<{ className?: string }>> = {
  boss: Brain,
  manager: Bot,
  worker: Cpu,
  zip: Archive,
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
  manager: ['Backend', 'Frontend', 'Fullstack', 'DevOps', 'Mobile', 'QA Lead'],
  worker: ['Developer', 'UI Developer', 'API Developer', 'Database', 'Tester', 'Documentation'],
  zip: ['Archive'],
};

export function NodeContextMenu({ x, y, nodeId, nodeType, nodeRole, onClose }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const updateNode = useTaskStore((state) => state.updateNode);
  const removeNode = useTaskStore((state) => state.removeNode);
  const n8nIntegration = useIntegrationStore((state) => state.integrations.n8n);
  const { t } = useI18n();
  const [n8nTrigger, setN8nTrigger] = useState<'start' | 'end' | 'middle' | 'custom' | null>(node?.n8nTrigger || null);
  const [customPercentage, setCustomPercentage] = useState(node?.n8nPercentage || 50);
  const [n8nWorkflows, setN8nWorkflows] = useState<N8nWorkflow[]>([]);
  const [loadingWorkflows, setLoadingWorkflows] = useState(false);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string>(node?.n8nWorkflowId || '');
  const [webhookUrl, setWebhookUrl] = useState<string>(node?.n8nWebhookUrl || '');

  // Закрытие при клике вне меню
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [onClose]);

  // Загрузка n8n workflow при открытии меню
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

  const handleDuplicate = () => {
    const addNode = useTaskStore.getState().addNode;
    const currentNode = useTaskStore.getState().nodes.find((n) => n.id === nodeId);
    if (currentNode) {
      addNode({
        id: `node-${Date.now()}`,
        type: currentNode.type,
        role: currentNode.role,
        status: currentNode.status,
        progress: 0,
        position: {
          x: (currentNode.position?.x || 0) + 50,
          y: (currentNode.position?.y || 0) + 50,
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

  const handleWebhookUrlChange = (url: string) => {
    setWebhookUrl(url);
    updateNode(nodeId, { n8nWebhookUrl: url });
  };

  const handleCustomPercentageChange = (value: number) => {
    setCustomPercentage(value);
    if (n8nTrigger === 'custom') {
      updateNode(nodeId, { n8nPercentage: value });
    }
  };

  return (
    <div
      ref={menuRef}
      className="fixed z-[100] bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-xl min-w-[220px] overflow-hidden"
      style={{ left: x, top: y }}
    >
      {/* Header с изображением */}
      <div className="px-4 py-3 border-b border-[var(--border)] bg-[var(--background)] flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center">
          {(() => {
            const IconComponent = nodeIcons[nodeType];
            return <IconComponent className="w-6 h-6 text-white" />;
          })()}
        </div>
        <div className="text-sm font-semibold text-[var(--text)]">
          {nodeType.toUpperCase()} — {nodeRole}
        </div>
      </div>

      {/* Смена роли */}
      <div className="p-2 border-b border-[var(--border)]">
        <div className="text-xs text-[var(--text-muted)] mb-1 px-2">{t('contextMenu.changeRole')}</div>
        {roleOptions[nodeType]?.map((role) => (
          <button
            key={role}
            onClick={() => handleRoleChange(role)}
            className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
              role === nodeRole
                ? 'bg-[var(--accent)] text-white'
                : 'hover:bg-[var(--background)] text-[var(--text)]'
            }`}
          >
            {role}
          </button>
        ))}
      </div>

      {/* N8n интеграция */}
      {n8nIntegration.connected && (
        <div className="p-2 border-b border-[var(--border)]">
          <div className="text-xs text-[var(--text-muted)] mb-2 px-2 flex items-center gap-1">
            <Zap className="w-3 h-3" />
            N8n автоматизация
          </div>

          {/* Выбор workflow или webhook */}
          <div className="px-2 mb-3">
            <div className="text-xs font-medium text-[var(--text)] mb-1">Workflow</div>
            <select
              value={selectedWorkflowId}
              onChange={(e) => handleWorkflowSelect(e.target.value)}
              className="w-full px-2 py-1 text-sm bg-[var(--background)] border border-[var(--border)] rounded text-[var(--text)]"
            >
              <option value="">Выберите workflow...</option>
              {loadingWorkflows ? (
                <option disabled>Загрузка...</option>
              ) : (
                n8nWorkflows.map((workflow) => (
                  <option key={workflow.id} value={workflow.id}>
                    {workflow.name}
                  </option>
                ))
              )}
            </select>

            <div className="text-xs font-medium text-[var(--text)] mt-2 mb-1">Или Webhook URL</div>
            <input
              type="url"
              value={webhookUrl}
              onChange={(e) => handleWebhookUrlChange(e.target.value)}
              placeholder="https://your-n8n.com/webhook/..."
              className="w-full px-2 py-1 text-sm bg-[var(--background)] border border-[var(--border)] rounded text-[var(--text)] placeholder-[var(--text-muted)]"
            />
          </div>

          {/* Выбор триггера */}
          <div className="space-y-1">
            <div className="text-xs font-medium text-[var(--text)] px-2 mb-1">Когда запускать:</div>
            <button
              onClick={() => handleN8nTriggerChange('start')}
              className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
                n8nTrigger === 'start'
                  ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400'
                  : 'hover:bg-[var(--background)] text-[var(--text)]'
              }`}
            >
              В начале задачи
            </button>
            <button
              onClick={() => handleN8nTriggerChange('middle')}
              className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
                n8nTrigger === 'middle'
                  ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400'
                  : 'hover:bg-[var(--background)] text-[var(--text)]'
              }`}
            >
              На 50% выполнения
            </button>
            <button
              onClick={() => handleN8nTriggerChange('end')}
              className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
                n8nTrigger === 'end'
                  ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400'
                  : 'hover:bg-[var(--background)] text-[var(--text)]'
              }`}
            >
              В конце задачи
            </button>
            <button
              onClick={() => handleN8nTriggerChange('custom')}
              className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
                n8nTrigger === 'custom'
                  ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400'
                  : 'hover:bg-[var(--background)] text-[var(--text)]'
              }`}
            >
              На {customPercentage}% выполнения
            </button>
            {n8nTrigger === 'custom' && (
              <div className="px-3 py-1">
                <input
                  type="range"
                  min="0"
                  max="100"
                  value={customPercentage}
                  onChange={(e) => handleCustomPercentageChange(Number(e.target.value))}
                  className="w-full"
                />
              </div>
            )}
          </div>
        </div>
      )}

      {/* Действия */}
      <div className="p-2">
        <button
          onClick={handleDuplicate}
          className="w-full text-left px-3 py-1.5 text-sm rounded hover:bg-[var(--background)] text-[var(--text)] transition-colors flex items-center gap-2"
        >
          <span>Duplicate</span> {t('contextMenu.duplicate')}
        </button>
        <button
          onClick={handleDelete}
          className="w-full text-left px-3 py-1.5 text-sm rounded hover:bg-red-500/20 text-red-500 transition-colors flex items-center gap-2"
        >
          <span>Delete</span> {t('contextMenu.delete')}
        </button>
      </div>
    </div>
  );
}

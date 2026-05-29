import { useState, useEffect, useCallback } from 'react';
import { Brain, Bot, Cpu, X, GripVertical, Library, Trash2, Workflow as WorkflowIcon } from 'lucide-react';
import { useI18n } from '../../hooks/useI18n';
import { useTaskStore } from '../../stores/taskStore';
import { getMyWorkflows, deleteWorkflow, type Workflow } from '../../services/workflowService';

interface NodeTemplate {
  type: 'boss' | 'manager' | 'worker';
  label: string;
  typeLabel: string;
  description: string;
  accent: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NodeSidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onDragStart: (type: 'boss' | 'manager' | 'worker', event: React.DragEvent) => void;
  onOpenWorkflowLibrary?: () => void;
}

export function NodeSidebar({ isOpen, onClose, onOpenWorkflowLibrary }: NodeSidebarProps) {
  const { t } = useI18n();
  // Subscribe so the sidebar re-renders if the store is reset elsewhere.
  useTaskStore((state) => state.nodes.length);

  const [draggingType, setDraggingType] = useState<'boss' | 'manager' | 'worker' | null>(null);
  const [draggingWorkflowId, setDraggingWorkflowId] = useState<string | null>(null);
  const [myWorkflows, setMyWorkflows] = useState<Workflow[]>([]);
  const [loadingWorkflows, setLoadingWorkflows] = useState(false);
  const [loadedOnce, setLoadedOnce] = useState(false);

  const loadWorkflows = useCallback(async () => {
    setLoadingWorkflows(true);
    try {
      const workflows = await getMyWorkflows();
      setMyWorkflows(workflows);
    } catch {
      // ignore
    } finally {
      setLoadingWorkflows(false);
      setLoadedOnce(true);
    }
  }, []);

  // Load workflows the first time the sidebar opens.
  useEffect(() => {
    if (isOpen && !loadedOnce) {
      loadWorkflows();
    }
  }, [isOpen, loadedOnce, loadWorkflows]);

  const templates: NodeTemplate[] = [
    {
      type: 'boss',
      label: t('sidebar.boss.label'),
      typeLabel: t('sidebar.boss.type'),
      description: t('sidebar.boss.description'),
      accent: '#f97316',
      icon: Brain,
    },
    {
      type: 'manager',
      label: t('sidebar.manager.label'),
      typeLabel: t('sidebar.manager.type'),
      description: t('sidebar.manager.description'),
      accent: '#7c4dff',
      icon: Bot,
    },
    {
      type: 'worker',
      label: t('sidebar.worker.label'),
      typeLabel: t('sidebar.worker.type'),
      description: t('sidebar.worker.description'),
      accent: '#22c55e',
      icon: Cpu,
    },
  ];

  const handleDragStart = (template: NodeTemplate, event: React.DragEvent) => {
    setDraggingType(template.type);
    event.dataTransfer.setData('application/reactflow/node-type', template.type);
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleDragEnd = () => setDraggingType(null);

  const handleWorkflowDragStart = (workflow: Workflow, event: React.DragEvent) => {
    setDraggingWorkflowId(workflow.id);
    try {
      const nodesData = workflow.nodes ? JSON.parse(workflow.nodes) : [];
      const edgesData = workflow.edges ? JSON.parse(workflow.edges) : [];
      event.dataTransfer.setData('application/reactflow/workflow-id', workflow.id);
      event.dataTransfer.setData(
        'application/reactflow/workflow-data',
        JSON.stringify({ nodes: nodesData, edges: edgesData })
      );
      event.dataTransfer.effectAllowed = 'move';
    } catch {
      event.preventDefault();
    }
  };

  const handleWorkflowDragEnd = () => setDraggingWorkflowId(null);

  const handleDeleteWorkflow = useCallback(async (workflowId: string) => {
    try {
      await deleteWorkflow(workflowId);
      setMyWorkflows((prev) => prev.filter((w) => w.id !== workflowId));
    } catch {
      // ignore
    }
  }, []);

  return (
    <>
      {/* Overlay (mobile) */}
      {isOpen && <div className="fixed inset-0 bg-black/30 z-40 lg:hidden" onClick={onClose} />}

      {/* Panel */}
      <div
        data-sidebar
        className={`fixed right-0 top-0 h-full w-80 bg-[var(--surface)] border-l border-[var(--border)] shadow-2xl z-50 flex flex-col transform transition-transform duration-300 ease-in-out ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="flex items-start justify-between px-4 py-4 border-b border-[var(--border)] flex-shrink-0">
          <div>
            <h2 className="text-lg font-bold text-[var(--text)]">{t('sidebar.title')}</h2>
            <p className="text-xs text-[var(--text-muted)] mt-0.5 leading-snug max-w-[220px]">
              {t('sidebar.subtitle')}
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-[var(--text-muted)] hover:text-[var(--text)] hover:bg-[var(--background)] transition-colors p-1.5 rounded-lg"
            aria-label="Close"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 overflow-y-auto sidebar-scroll">
          <div className="p-4 space-y-6">
            {/* Agent palette */}
            <section>
              <div className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-2 px-1">
                {t('sidebar.addAgent')}
              </div>
              <div className="space-y-2">
                {templates.map((template) => {
                  const Icon = template.icon;
                  const isDragging = draggingType === template.type;
                  return (
                    <div
                      key={template.type}
                      draggable
                      onDragStart={(e) => handleDragStart(template, e)}
                      onDragEnd={handleDragEnd}
                      className={`group relative flex items-center gap-3 p-3 rounded-xl border border-[var(--border)] bg-[var(--background)] cursor-grab active:cursor-grabbing transition-all duration-150 hover:shadow-md ${
                        isDragging ? 'opacity-50 scale-95' : ''
                      }`}
                      style={{ borderLeft: `3px solid ${template.accent}` }}
                    >
                      <div
                        className="w-10 h-10 rounded-lg flex items-center justify-center shrink-0"
                        style={{ backgroundColor: template.accent, boxShadow: `0 2px 8px ${template.accent}55` }}
                      >
                        <Icon className="w-5 h-5 text-white" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-semibold text-[var(--text)]">{template.label}</span>
                          <span
                            className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                            style={{ backgroundColor: `${template.accent}22`, color: template.accent }}
                          >
                            {template.typeLabel}
                          </span>
                        </div>
                        <div className="text-xs text-[var(--text-muted)] mt-0.5 leading-snug">
                          {template.description}
                        </div>
                      </div>
                      <GripVertical className="w-4 h-4 text-[var(--text-muted)] opacity-0 group-hover:opacity-60 transition-opacity shrink-0" />
                    </div>
                  );
                })}
              </div>
            </section>

            {/* My workflows */}
            <section>
              <div className="flex items-center justify-between mb-2 px-1">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">
                  {t('sidebar.myWorkflows')}
                </span>
                {onOpenWorkflowLibrary && (
                  <button
                    onClick={onOpenWorkflowLibrary}
                    className="flex items-center gap-1 text-[11px] font-medium text-[var(--accent)] hover:text-[var(--accent-hover)] transition-colors"
                  >
                    <Library className="w-3.5 h-3.5" />
                    {t('sidebar.openLibrary')}
                  </button>
                )}
              </div>

              <div className="rounded-xl border border-[var(--border)] bg-[var(--background)] overflow-hidden">
                {loadingWorkflows ? (
                  <div className="px-4 py-6 text-center text-sm text-[var(--text-muted)]">
                    {t('workflowLibrary.loading')}
                  </div>
                ) : myWorkflows.length === 0 ? (
                  <div className="px-4 py-8 text-center">
                    <WorkflowIcon className="w-6 h-6 mx-auto text-[var(--text-muted)] opacity-50 mb-2" />
                    <div className="text-sm text-[var(--text-muted)]">{t('sidebar.noWorkflows')}</div>
                  </div>
                ) : (
                  myWorkflows.map((workflow) => {
                    const isDragging = draggingWorkflowId === workflow.id;
                    return (
                      <div
                        key={workflow.id}
                        draggable
                        onDragStart={(e) => handleWorkflowDragStart(workflow, e)}
                        onDragEnd={handleWorkflowDragEnd}
                        className={`group flex items-center gap-3 px-3 py-2.5 hover:bg-[var(--surface)] cursor-grab active:cursor-grabbing transition-colors border-b border-[var(--border)] last:border-b-0 ${
                          isDragging ? 'opacity-50' : ''
                        }`}
                      >
                        <div className="w-9 h-9 rounded-lg bg-[var(--surface)] border border-[var(--border)] flex items-center justify-center flex-shrink-0">
                          <WorkflowIcon className="w-4 h-4 text-[var(--accent)]" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-semibold text-[var(--text)] truncate">{workflow.name}</div>
                          <div className="text-xs text-[var(--text-muted)] flex items-center gap-1">
                            <span>{workflow.downloads} ↓</span>
                            {workflow.category && (
                              <>
                                <span>•</span>
                                <span>{workflow.category}</span>
                              </>
                            )}
                          </div>
                        </div>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            handleDeleteWorkflow(workflow.id);
                          }}
                          className="opacity-0 group-hover:opacity-100 p-1.5 text-red-400 hover:text-red-500 hover:bg-red-500/10 rounded-lg transition-all flex-shrink-0"
                          title={t('contextMenu.delete')}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    );
                  })
                )}
              </div>
            </section>

            {/* Tip */}
            <div className="p-3 bg-[var(--background)] rounded-xl border border-[var(--border)]">
              <div className="text-xs text-[var(--text-muted)] space-y-2">
                <p className="font-semibold text-[var(--text)]">{t('sidebar.tip.title')}</p>
                <ul className="space-y-1">
                  {[0, 1, 2, 3].map((i) => (
                    <li key={i} className="flex items-start gap-1.5">
                      <span className="text-[var(--accent)] mt-0.5">•</span>
                      <span>{t(`sidebar.tip.items.${i}`)}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

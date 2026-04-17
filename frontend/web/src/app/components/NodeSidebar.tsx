import { useState, useEffect, useCallback } from 'react';
import { Brain, Bot, Cpu } from 'lucide-react';
import { useI18n } from '../../hooks/useI18n';
import { useTaskStore, type AgentNode } from '../../stores/taskStore';
import { getMyWorkflows, deleteWorkflow, type Workflow } from '../../services/workflowService';

interface NodeTemplate {
  type: 'boss' | 'manager' | 'worker';
  label: string;
  typeLabel: string;
  description: string;
  color: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NodeSidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onDragStart: (type: 'boss' | 'manager' | 'worker', event: React.DragEvent) => void;
  onOpenWorkflowLibrary?: () => void;
}

export function NodeSidebar({ isOpen, onClose, onDragStart, onOpenWorkflowLibrary }: NodeSidebarProps) {
  const { t } = useI18n();
  const addNodeToStore = useTaskStore((state) => state.addNode);
  const addEdgeToStore = useTaskStore((state) => state.addEdge);

  // Dropdown state
  const [createOpen, setCreateOpen] = useState(false);
  const [workflowsOpen, setWorkflowsOpen] = useState(false);
  const [draggingType, setDraggingType] = useState<'boss' | 'manager' | 'worker' | null>(null);
  const [draggingWorkflow, setDraggingWorkflow] = useState<Workflow | null>(null);
  const [myWorkflows, setMyWorkflows] = useState<Workflow[]>([]);
  const [loadingWorkflows, setLoadingWorkflows] = useState(false);

  // Закрытие dropdown при закрытии sidebar
  useEffect(() => {
    if (!isOpen) {
      setCreateOpen(false);
      setWorkflowsOpen(false);
    }
  }, [isOpen]);

  // Загрузка workflow при открытии dropdown
  useEffect(() => {
    if (workflowsOpen && myWorkflows.length === 0) {
      loadWorkflows();
    }
  }, [workflowsOpen]);

  const loadWorkflows = async () => {
    setLoadingWorkflows(true);
    try {
      const workflows = await getMyWorkflows();
      setMyWorkflows(workflows);
    } catch {
      // ignore
    } finally {
      setLoadingWorkflows(false);
    }
  };

  const templates: NodeTemplate[] = [
    {
      type: 'boss',
      label: t('sidebar.boss.label'),
      typeLabel: t('sidebar.boss.type'),
      description: t('sidebar.boss.description'),
      color: 'from-orange-500 to-orange-600',
      icon: Brain,
    },
    {
      type: 'manager',
      label: t('sidebar.manager.label'),
      typeLabel: t('sidebar.manager.type'),
      description: t('sidebar.manager.description'),
      color: 'from-blue-500 to-blue-600',
      icon: Bot,
    },
    {
      type: 'worker',
      label: t('sidebar.worker.label'),
      typeLabel: t('sidebar.worker.type'),
      description: t('sidebar.worker.description'),
      color: 'from-green-500 to-green-600',
      icon: Cpu,
    },
  ];

  const handleDragStart = (template: NodeTemplate, event: React.DragEvent) => {
    setDraggingType(template.type);
    event.dataTransfer.setData('application/reactflow/node-type', template.type);
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleDragEnd = () => {
    setDraggingType(null);
  };

  // Drag-n-drop workflow на канвас
  const handleWorkflowDragStart = (workflow: Workflow, event: React.DragEvent) => {
    setDraggingWorkflow(workflow);
    try {
      const nodesData = workflow.nodes ? JSON.parse(workflow.nodes) : [];
      const edgesData = workflow.edges ? JSON.parse(workflow.edges) : [];
      event.dataTransfer.setData('application/reactflow/workflow-id', workflow.id);
      event.dataTransfer.setData('application/reactflow/workflow-data', JSON.stringify({
        nodes: nodesData,
        edges: edgesData,
      }));
      event.dataTransfer.effectAllowed = 'move';
    } catch {
      event.preventDefault();
    }
  };

  const handleWorkflowDragEnd = () => {
    setDraggingWorkflow(null);
  };

  const handleDeleteWorkflow = useCallback(async (workflowId: string) => {
    if (!confirm('Удалить этот workflow?')) return;
    try {
      await deleteWorkflow(workflowId);
      setMyWorkflows((prev) => prev.filter((w) => w.id !== workflowId));
    } catch {
      // ignore
    }
  }, []);

  // Close dropdown when clicking outside the sidebar
  const handleClickOutside = useCallback((e: MouseEvent) => {
    // Only close if clicking outside the entire sidebar
    const sidebar = document.querySelector('[data-sidebar]');
    if (sidebar && !sidebar.contains(e.target as Node)) {
      setCreateOpen(false);
      setWorkflowsOpen(false);
    }
  }, []);

  useEffect(() => {
    if (createOpen || workflowsOpen) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [createOpen, workflowsOpen, handleClickOutside]);

  const toggleCreate = useCallback(() => {
    setCreateOpen((prev) => !prev);
  }, []);

  const toggleWorkflows = useCallback(() => {
    setWorkflowsOpen((prev) => !prev);
  }, []);

  return (
    <>
      {/* Overlay для закрытия */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/30 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Панель */}
      <div
        data-sidebar
        className={`fixed right-0 top-0 h-full w-72 bg-[var(--surface)] border-l border-[var(--border)] shadow-xl z-50 flex flex-col transform transition-transform duration-300 ease-in-out ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-[var(--border)] flex-shrink-0">
          <h2 className="text-lg font-bold text-[var(--text)]">{t('sidebar.title')}</h2>
          <button
            onClick={onClose}
            className="text-[var(--text-muted)] hover:text-[var(--text)] transition-colors p-1 rounded"
          >
            ✕
          </button>
        </div>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto sidebar-scroll">
          <div className="p-4 space-y-3">

            {/* Dropdown: Создать */}
            <div className="relative" onClick={(e) => e.stopPropagation()}>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  toggleCreate();
                }}
                className="w-full flex items-center justify-between px-4 py-3 bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-hover)] text-[var(--text-inverse)] font-semibold rounded-lg transition-all duration-200 shadow-md hover:shadow-lg"
              >
                <div className="flex items-center gap-2">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 1.5a6.5 6.5 0 100 13 6.5 6.5 0 000-13zM0 8a8 8 0 1116 0A8 8 0 010 8z"/>
                    <path d="M8 4a.75.75 0 01.75.75v2.5h2.5a.75.75 0 010 1.5h-2.5v2.5a.75.75 0 01-1.5 0v-2.5h-2.5a.75.75 0 010-1.5h2.5v-2.5A.75.75 0 018 4z"/>
                  </svg>
                  <span>{t('sidebar.create')}</span>
                </div>
                <svg
                  width="12" height="12" viewBox="0 0 12 12" fill="currentColor"
                  className={`transition-transform duration-200 ${createOpen ? 'rotate-180' : ''}`}
                >
                  <path d="M2.5 4.5a.75.75 0 011.06 0L6 6.94l2.44-2.44a.75.75 0 111.06 1.06l-3 3a.75.75 0 01-1.06 0l-3-3a.75.75 0 010-1.06z"/>
                </svg>
              </button>

              {/* Create Dropdown Items */}
              {createOpen && (
                <div className="w-full mt-2 bg-[var(--bg-elevated)] border border-[var(--border)] rounded-lg shadow-xl overflow-hidden z-40">
                  {templates.map((template) => (
                    <div
                      key={template.type}
                      draggable
                      onDragStart={(e) => handleDragStart(template, e)}
                      onDragEnd={handleDragEnd}
                      onClick={() => setCreateOpen(false)}
                      className="cursor-grab active:cursor-grabbing border-b border-[var(--border)] last:border-b-0 hover:opacity-90 transition-opacity"
                    >
                      {/* Icon header on top */}
                      <div className={`bg-gradient-to-br ${template.color} p-4 flex items-center justify-center`}>
                        <div className="bg-white/10 p-3 rounded-xl backdrop-blur-sm">
                          <template.icon className="w-12 h-12 text-white drop-shadow-lg" />
                        </div>
                      </div>
                      {/* Info below */}
                      <div className="p-4 flex items-center justify-between">
                        <div className="flex-1">
                          <div className="text-sm font-bold text-[var(--text)]">{template.label}</div>
                          <div className="text-xs text-[var(--text-tertiary)] mt-1">{template.description}</div>
                        </div>
                        <svg width="16" height="16" viewBox="0 0 14 14" fill="var(--text-tertiary)" className="flex-shrink-0 ml-3">
                          <path d="M1 7h12M7 1v12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                        </svg>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Dropdown: Мои Workflow */}
            <div className="relative" onClick={(e) => e.stopPropagation()}>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  toggleWorkflows();
                }}
                className="w-full flex items-center justify-between px-4 py-3 bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white font-semibold rounded-lg transition-all duration-200 shadow-md hover:shadow-lg"
              >
                <div className="flex items-center gap-2">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M1 3a1 1 0 011-1h12a1 1 0 011 1v10a1 1 0 01-1 1H2a1 1 0 01-1-1V3zm1 0v10h12V3H2z"/>
                    <path d="M4 5h8v1H4V5zm0 2h8v1H4V7zm0 2h5v1H4V9z"/>
                  </svg>
                  <span>{t('sidebar.myWorkflows')}</span>
                </div>
                <svg
                  width="12" height="12" viewBox="0 0 12 12" fill="currentColor"
                  className={`transition-transform duration-200 ${workflowsOpen ? 'rotate-180' : ''}`}
                >
                  <path d="M2.5 4.5a.75.75 0 011.06 0L6 6.94l2.44-2.44a.75.75 0 111.06 1.06l-3 3a.75.75 0 01-1.06 0l-3-3a.75.75 0 010-1.06z"/>
                </svg>
              </button>

              {/* Workflows Dropdown Items */}
              {workflowsOpen && (
                <div className="w-full mt-2 bg-[var(--bg-elevated)] border border-[var(--border)] rounded-lg shadow-xl overflow-hidden z-40 max-h-64 overflow-y-auto sidebar-scroll">
                  {loadingWorkflows ? (
                    <div className="px-4 py-6 text-center text-sm text-[var(--text-tertiary)]">
                      {t('workflowLibrary.loading')}
                    </div>
                  ) : myWorkflows.length === 0 ? (
                    <div className="px-4 py-6 text-center text-sm text-[var(--text-tertiary)]">
                      {t('sidebar.noWorkflows')}
                    </div>
                  ) : (
                    myWorkflows.map((workflow) => (
                      <div
                        key={workflow.id}
                        draggable
                        onDragStart={(e) => handleWorkflowDragStart(workflow, e)}
                        onDragEnd={handleWorkflowDragEnd}
                        className="group flex items-center gap-3 px-3 py-2.5 hover:bg-[var(--bg-tertiary)] cursor-grab active:cursor-grabbing transition-colors border-b border-[var(--border)] last:border-b-0"
                      >
                        <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-500/20 to-purple-600/20 border border-purple-500/30 flex items-center justify-center flex-shrink-0">
                          <svg width="18" height="18" viewBox="0 0 16 16" fill="var(--color-secondary)">
                            <path d="M1 3a1 1 0 011-1h12a1 1 0 011 1v10a1 1 0 01-1 1H2a1 1 0 01-1-1V3zm1 0v10h12V3H2z"/>
                            <path d="M4 5h8v1H4V5zm0 2h8v1H4V7zm0 2h5v1H4V9z"/>
                          </svg>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-semibold text-[var(--text)] truncate">{workflow.name}</div>
                          <div className="text-xs text-[var(--text-tertiary)] flex items-center gap-1">
                            <span>{workflow.downloads}</span>
                            {workflow.category && (
                              <>
                                <span>•</span>
                                <span className="text-purple-400">{workflow.category}</span>
                              </>
                            )}
                          </div>
                        </div>
                        {/* Delete button — shows on hover */}
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            handleDeleteWorkflow(workflow.id);
                          }}
                          className="opacity-0 group-hover:opacity-100 p-1 text-red-400 hover:text-red-500 hover:bg-red-500/10 rounded transition-all flex-shrink-0"
                          title="Удалить"
                        >
                          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                            <path d="M2 4h10M5 4V3a1 1 0 011-1h2a1 1 0 011 1v1M11 4v7a1 1 0 01-1 1H4a1 1 0 01-1-1V4"/>
                            <line x1="6" y1="6.5" x2="6" y2="9.5"/>
                            <line x1="8" y1="6.5" x2="8" y2="9.5"/>
                          </svg>
                        </button>
                      </div>
                    ))
                  )}

                  {/* Divider + Library button */}
                  {onOpenWorkflowLibrary && (
                    <>
                      <div className="border-t border-[var(--border)]" />
                      <button
                        onClick={() => {
                          setWorkflowsOpen(false);
                          onOpenWorkflowLibrary();
                        }}
                        className="w-full px-3 py-2.5 text-xs font-medium text-[var(--color-secondary)] hover:bg-[var(--bg-tertiary)] transition-colors"
                      >
                        {t('sidebar.openLibrary')} →
                      </button>
                    </>
                  )}
                </div>
              )}
            </div>

            {/* Подсказка */}
            <div className="mt-6 p-3 bg-[var(--background)] rounded-lg border border-[var(--border)]">
              <div className="text-xs text-[var(--text-muted)] space-y-2">
                <p><strong>{t('sidebar.tip.title')}</strong></p>
                <ul className="list-disc list-inside space-y-1">
                  <li>{t('sidebar.tip.items.0')}</li>
                  <li>{t('sidebar.tip.items.1')}</li>
                  <li>{t('sidebar.tip.items.2')}</li>
                  <li>{t('sidebar.tip.items.3')}</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

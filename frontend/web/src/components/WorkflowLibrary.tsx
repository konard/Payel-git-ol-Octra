/**
 * WorkflowLibrary Component
 * Модальное окно с библиотекой workflow - просмотр, сохранение, публикация
 */

import { useState, useEffect } from 'react';
import {
  getPublicWorkflows,
  getMyWorkflows,
  createWorkflow,
  deleteWorkflow,
  downloadWorkflow,
  getWorkflowCategories,
  type Workflow,
} from '../services/workflowService';
import { useAuthStore } from '../stores/authStore';
import { useTaskStore } from '../stores/taskStore';
import { useI18n } from '../hooks/useI18n';

type LibraryView = 'public' | 'my';

interface WorkflowLibraryProps {
  onClose: () => void;
  onImportWorkflow: (workflow: { nodes: any[]; edges: any[] }) => void;
}

export function WorkflowLibrary({ onClose, onImportWorkflow }: WorkflowLibraryProps) {
  const { t } = useI18n();
  const [view, setView] = useState<LibraryView>('public');
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [saveName, setSaveName] = useState('');
  const [saveDescription, setSaveDescription] = useState('');
  const [saveCategory, setSaveCategory] = useState('');
  const [isPublic, setIsPublic] = useState(true);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [importJson, setImportJson] = useState('');
  const [importName, setImportName] = useState('');
  const [importDescription, setImportDescription] = useState('');
  const [importCategory, setImportCategory] = useState('');
  const [importIsPublic, setImportIsPublic] = useState(true);
  const [importError, setImportError] = useState<string | null>(null);

  const { isAuthenticated, user } = useAuthStore();
  const nodes = useTaskStore((state) => state.nodes);
  const edges = useTaskStore((state) => state.edges);

  // Load workflows on mount and view change
  useEffect(() => {
    loadWorkflows();
    loadCategories();
  }, [view, selectedCategory]);

  const loadWorkflows = async () => {
    setIsLoading(true);
    setError(null);
    try {
      if (view === 'public') {
        const data = await getPublicWorkflows(selectedCategory || undefined);
        setWorkflows(data);
      } else {
        if (!isAuthenticated) {
          setError(t('workflowLibrary.noAuth'));
          setWorkflows([]);
          return;
        }
        const data = await getMyWorkflows();
        setWorkflows(data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t('workflowLibrary.noPublic'));
    } finally {
      setIsLoading(false);
    }
  };

  const loadCategories = async () => {
    try {
      const cats = await getWorkflowCategories();
      setCategories(cats);
    } catch {
      // ignore
    }
  };

  const handleImport = async (workflow: Workflow) => {
    try {
      const parsedNodes = JSON.parse(workflow.nodes);
      const parsedEdges = JSON.parse(workflow.edges);
      await downloadWorkflow(workflow.id);
      onImportWorkflow({ nodes: parsedNodes, edges: parsedEdges });
      onClose();
    } catch {
      setError('Ошибка импорта workflow');
    }
  };

  const handleDelete = async (workflowId: string) => {
    if (!confirm(t('workflowLibrary.deleteConfirm'))) return;
    try {
      await deleteWorkflow(workflowId);
      setWorkflows(workflows.filter((w) => w.id !== workflowId));
    } catch {
      setError('Ошибка удаления workflow');
    }
  };

  const handleImportJSON = async () => {
    if (!importJson.trim()) {
      setImportError('Вставьте JSON или загрузите файл');
      return;
    }
    try {
      const data = JSON.parse(importJson);
      if (!data.nodes || !data.edges) {
        setImportError('Неверный формат: должны быть поля nodes и edges');
        return;
      }
      if (!importName.trim()) {
        setImportError('Введите название workflow');
        return;
      }
      setIsLoading(true);
      setImportError(null);
      await createWorkflow({
        name: importName,
        description: importDescription,
        category: importCategory,
        tags: [],
        nodes: JSON.stringify(data.nodes),
        edges: JSON.stringify(data.edges),
        is_public: importIsPublic,
      });
      setShowImportDialog(false);
      setImportJson('');
      setImportName('');
      setImportDescription('');
      setImportCategory('');
      setImportIsPublic(true);
      loadWorkflows();
    } catch (err) {
      setImportError(err instanceof Error ? err.message : 'Ошибка сохранения');
    } finally {
      setIsLoading(false);
    }
  };

  const handleImportFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      try {
        const text = ev.target?.result as string;
        const data = JSON.parse(text);
        // Поддерживаем оба формата: прямой (nodes/edges) и из экспорта
        const nodes = data.nodes || [];
        const edges = data.edges || [];
        setImportJson(JSON.stringify({ nodes, edges }, null, 2));
        // Авто-имя из файла
        if (!importName) {
          setImportName(file.name.replace('.json', ''));
        }
        setImportError(null);
      } catch {
        setImportError(t('workflowLibrary.invalidJson'));
      }
    };
    reader.readAsText(file);
  };

  const handleSaveCurrentWorkflow = async () => {
    if (!saveName.trim()) {
      setError(t('workflowLibrary.nameLabel'));
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      await createWorkflow({
        name: saveName,
        description: saveDescription,
        category: saveCategory,
        tags: [],
        nodes: JSON.stringify(nodes),
        edges: JSON.stringify(edges),
        is_public: isPublic,
      });
      setShowSaveDialog(false);
      setSaveName('');
      setSaveDescription('');
      setSaveCategory('');
      setIsPublic(true);
      loadWorkflows();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения');
    } finally {
      setIsLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="relative bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-[900px] max-h-[85vh] mx-4 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--border)]">
          <h2 className="text-xl font-semibold text-[var(--text)]">{t('workflowLibrary.title')}</h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-[var(--background)] rounded-md transition-colors"
            aria-label="Close"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="1" y1="1" x2="13" y2="13" />
              <line x1="13" y1="1" x2="1" y2="13" />
            </svg>
          </button>
        </div>

        {/* Tabs & Controls */}
        <div className="flex items-center justify-between px-6 py-3 border-b border-[var(--border)] gap-4 flex-wrap">
          <div className="flex gap-2">
            <button
              onClick={() => setView('public')}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                view === 'public'
                  ? 'bg-[var(--accent)] text-white'
                  : 'bg-[var(--background)] text-[var(--text)] hover:bg-[var(--accent)]/10'
              }`}
            >
              {t('workflowLibrary.public')}
            </button>
            <button
              onClick={() => setView('my')}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                view === 'my'
                  ? 'bg-[var(--accent)] text-white'
                  : 'bg-[var(--background)] text-[var(--text)] hover:bg-[var(--accent)]/10'
              }`}
            >
              {t('workflowLibrary.my')}
            </button>
          </div>

          {view === 'public' && categories.length > 0 && (
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text)] focus:outline-none focus:border-[var(--accent)]"
            >
              <option value="">{t('workflowLibrary.allCategories')}</option>
              {categories.map((cat) => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </select>
          )}

          {nodes.length > 0 && (
            <button
              onClick={() => setShowSaveDialog(true)}
              className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-md transition-colors"
            >
              {t('workflowLibrary.saveCurrent')}
            </button>
          )}

          {isAuthenticated && (
            <button
              onClick={() => {
                setShowImportDialog(true);
                setImportError(null);
                setImportJson('');
                setImportName('');
                setImportDescription('');
                setImportCategory('');
                setImportIsPublic(true);
              }}
              className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white text-sm font-medium rounded-md transition-colors"
            >
              {t('workflowLibrary.importWorkflow')}
            </button>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500 text-red-500 rounded-md text-sm">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="flex items-center justify-center h-40">
              <div className="text-[var(--text-muted)]">{t('workflowLibrary.loading')}</div>
            </div>
          ) : workflows.length === 0 ? (
            <div className="flex items-center justify-center h-40">
              <div className="text-[var(--text-muted)]">
                {view === 'my' ? t('workflowLibrary.noSaved') : t('workflowLibrary.noPublic')}
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {workflows.map((workflow) => (
                <div
                  key={workflow.id}
                  className="bg-[var(--background)] border border-[var(--border)] rounded-lg p-4 hover:border-[var(--accent)] transition-colors"
                >
                  <div className="flex items-start justify-between mb-2">
                    <h3 className="text-base font-semibold text-[var(--text)] flex-1 truncate">
                      {workflow.name}
                    </h3>
                    <span className="text-xs text-[var(--text-muted)] ml-2">
                      {workflow.downloads}
                    </span>
                  </div>

                  {workflow.description && (
                    <p className="text-sm text-[var(--text-secondary)] mb-2 line-clamp-2">
                      {workflow.description}
                    </p>
                  )}

                  <div className="flex items-center gap-2 text-xs text-[var(--text-muted)] mb-3 flex-wrap">
                    <span>{t('workflowLibrary.author')}: {workflow.author_name}</span>
                    <span>•</span>
                    <span>{formatDate(workflow.created_at)}</span>
                    {workflow.category && (
                      <>
                        <span>•</span>
                        <span className="bg-[var(--accent)]/10 text-[var(--accent)] px-2 py-0.5 rounded">
                          {workflow.category}
                        </span>
                      </>
                    )}
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => handleImport(workflow)}
                      className="flex-1 py-2 px-3 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white text-sm font-medium rounded-md transition-colors"
                    >
                      {t('workflowLibrary.import')}
                    </button>
                    {view === 'my' && (
                      <button
                        onClick={() => handleDelete(workflow.id)}
                        className="py-2 px-3 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-md transition-colors"
                      >
                        {t('workflowLibrary.delete')}
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Save Dialog */}
        {showSaveDialog && (
          <div
            className="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/60"
            onClick={() => setShowSaveDialog(false)}
          >
            <div
              className="relative bg-[var(--surface)] border border-[var(--border)] rounded-lg w-full max-w-md flex flex-col max-h-[90vh]"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Header — fixed */}
              <div className="px-6 py-4 border-b border-[var(--border)] flex-shrink-0">
                <h3 className="text-lg font-semibold text-[var(--text)]">
                  {t('workflowLibrary.saveTitle')}
                </h3>
                {user && (
                  <div className="inline-flex items-center gap-2 mt-2 px-3 py-1.5 bg-gradient-to-r from-green-500 to-green-600 rounded-lg shadow-sm">
                    <svg width="14" height="14" viewBox="0 0 14 14" fill="white">
                      <path d="M7 0a3.5 3.5 0 100 7 3.5 3.5 0 000-7zM2 12.5C2 10 4.5 8.5 7 8.5s5 1.5 5 4a.5.5 0 01-.5.5h-9a.5.5 0 01-.5-.5z"/>
                    </svg>
                    <span className="text-sm font-semibold text-white">{user.username}</span>
                  </div>
                )}
              </div>

              {/* Scrollable content */}
              <div className="px-6 py-4 space-y-4 overflow-y-auto flex-1 sidebar-scroll">
                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.nameLabel')}
                  </label>
                  <input
                    type="text"
                    value={saveName}
                    onChange={(e) => setSaveName(e.target.value)}
                    placeholder={t('workflowLibrary.namePlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
                    disabled={isLoading}
                    required
                  />
                </div>

                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.descriptionLabel')}
                  </label>
                  <textarea
                    value={saveDescription}
                    onChange={(e) => setSaveDescription(e.target.value)}
                    placeholder={t('workflowLibrary.descriptionPlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] resize-none"
                    rows={3}
                    disabled={isLoading}
                  />
                </div>

                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.categoryLabel')}
                  </label>
                  <input
                    type="text"
                    value={saveCategory}
                    onChange={(e) => setSaveCategory(e.target.value)}
                    placeholder={t('workflowLibrary.categoryPlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
                    disabled={isLoading}
                    list="category-suggestions"
                  />
                  <datalist id="category-suggestions">
                    {categories.map((cat) => (
                      <option key={cat} value={cat} />
                    ))}
                  </datalist>
                </div>

                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="is-public"
                    checked={isPublic}
                    onChange={(e) => setIsPublic(e.target.checked)}
                    className="w-4 h-4"
                    disabled={isLoading}
                  />
                  <label htmlFor="is-public" className="text-sm text-[var(--text)]">
                    {t('workflowLibrary.publishLabel')}
                  </label>
                </div>
              </div>

              {/* Footer — fixed */}
              <div className="px-6 py-4 border-t border-[var(--border)] flex gap-2 flex-shrink-0">
                <button
                  onClick={handleSaveCurrentWorkflow}
                  className="flex-1 py-2 px-4 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-md transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  {isLoading ? t('workflowLibrary.savingBtn') : t('workflowLibrary.saveBtn')}
                </button>
                <button
                  onClick={() => setShowSaveDialog(false)}
                  className="py-2 px-4 bg-[var(--background)] hover:bg-[var(--background)]/80 text-[var(--text)] font-medium rounded-md transition-colors text-sm border border-[var(--border)]"
                  disabled={isLoading}
                >
                  {t('workflowLibrary.cancelBtn')}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Import JSON Dialog */}
        {showImportDialog && (
          <div
            className="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/60"
            onClick={() => setShowImportDialog(false)}
          >
            <div
              className="relative bg-[var(--surface)] border border-[var(--border)] rounded-lg w-full max-w-2xl flex flex-col max-h-[90vh]"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Header */}
              <div className="px-6 py-4 border-b border-[var(--border)] flex-shrink-0">
                <h3 className="text-lg font-semibold text-[var(--text)]">
                  {t('workflowLibrary.importTitle')}
                </h3>
                <p className="text-xs text-[var(--text-tertiary)] mt-1">
                  {t('workflowLibrary.importDesc')}
                </p>
              </div>

              {/* Scrollable content */}
              <div className="px-6 py-4 space-y-4 overflow-y-auto flex-1 sidebar-scroll">
                {importError && (
                  <div className="p-3 bg-red-500/10 border border-red-500 text-red-500 rounded-md text-sm">
                    {importError}
                  </div>
                )}

                {/* File upload */}
                <div className="flex items-center gap-3">
                  <label className="px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text)] cursor-pointer hover:bg-[var(--bg-tertiary)] transition-colors">
                    {t('workflowLibrary.chooseFile')}
                    <input
                      type="file"
                      accept=".json"
                      onChange={handleImportFileUpload}
                      className="hidden"
                    />
                  </label>
                </div>

                {/* JSON textarea */}
                <textarea
                  value={importJson}
                  onChange={(e) => {
                    setImportJson(e.target.value);
                    setImportError(null);
                  }}
                  placeholder={t('workflowLibrary.pasteJson')}
                  className="w-full h-48 bg-[var(--background)] border border-[var(--border)] rounded-lg p-3 font-mono text-xs text-[var(--text)] resize-none focus:outline-none focus:border-[var(--accent)]"
                  spellCheck={false}
                />

                {/* Name */}
                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.nameLabel')}
                  </label>
                  <input
                    type="text"
                    value={importName}
                    onChange={(e) => setImportName(e.target.value)}
                    placeholder={t('workflowLibrary.namePlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
                  />
                </div>

                {/* Description */}
                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.descriptionLabel')}
                  </label>
                  <textarea
                    value={importDescription}
                    onChange={(e) => setImportDescription(e.target.value)}
                    placeholder={t('workflowLibrary.descriptionPlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] resize-none"
                    rows={2}
                  />
                </div>

                {/* Category */}
                <div>
                  <label className="text-sm font-medium text-[var(--text)] mb-1 block">
                    {t('workflowLibrary.categoryLabel')}
                  </label>
                  <input
                    type="text"
                    value={importCategory}
                    onChange={(e) => setImportCategory(e.target.value)}
                    placeholder={t('workflowLibrary.categoryPlaceholder')}
                    className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
                    list="import-category-suggestions"
                  />
                  <datalist id="import-category-suggestions">
                    {categories.map((cat) => (
                      <option key={cat} value={cat} />
                    ))}
                  </datalist>
                </div>

                {/* Public */}
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="import-is-public"
                    checked={importIsPublic}
                    onChange={(e) => setImportIsPublic(e.target.checked)}
                    className="w-4 h-4"
                  />
                  <label htmlFor="import-is-public" className="text-sm text-[var(--text)]">
                    {t('workflowLibrary.publishLabel')}
                  </label>
                </div>
              </div>

              {/* Footer */}
              <div className="px-6 py-4 border-t border-[var(--border)] flex gap-2 flex-shrink-0">
                <button
                  onClick={handleImportJSON}
                  className="flex-1 py-2 px-4 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-md transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                  disabled={isLoading}
                >
                  {isLoading ? t('workflowLibrary.savingBtn') : t('workflowLibrary.importBtn')}
                </button>
                <button
                  onClick={() => setShowImportDialog(false)}
                  className="py-2 px-4 bg-[var(--background)] hover:bg-[var(--background)]/80 text-[var(--text)] font-medium rounded-md transition-colors text-sm border border-[var(--border)]"
                  disabled={isLoading}
                >
                  {t('workflowLibrary.cancelBtn')}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

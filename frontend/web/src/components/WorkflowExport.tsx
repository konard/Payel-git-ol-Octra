/**
 * WorkflowExport Component
 * Кнопка экспорта workflow в JSON + модалка с просмотром/копированием/скачиванием
 */

import { useState, useCallback } from 'react';
import { useTaskStore } from '../stores/taskStore';
import { useAuthStore } from '../stores/authStore';
import { useI18n } from '../hooks/useI18n';

interface WorkflowExportProps {
  onImportJSON: (json: string) => void;
}

export function WorkflowExport({ onImportJSON }: WorkflowExportProps) {
  const { t } = useI18n();
  const nodes = useTaskStore((state) => state.nodes);
  const edges = useTaskStore((state) => state.edges);
  const user = useAuthStore((state) => state.user);
  const [isOpen, setIsOpen] = useState(false);
  const [jsonText, setJsonText] = useState('');
  const [copied, setCopied] = useState(false);
  const [isPublic, setIsPublic] = useState(true);

  const handleExport = useCallback(() => {
    const workflowData = {
      nodes,
      edges,
      author: user?.username || 'Anonymous',
      is_public: isPublic,
      exported_at: new Date().toISOString(),
    };
    const json = JSON.stringify(workflowData, null, 2);
    setJsonText(json);
    setCopied(false);
    setIsOpen(true);
  }, [nodes, edges, user, isPublic]);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(jsonText);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
      const textarea = document.createElement('textarea');
      textarea.value = jsonText;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [jsonText]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([jsonText], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `workflow-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }, [jsonText]);

  const handleImport = useCallback(() => {
    onImportJSON(jsonText);
    setIsOpen(false);
  }, [jsonText, onImportJSON]);

  const handleFileUpload = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const text = ev.target?.result as string;
      try {
        JSON.parse(text);
        setJsonText(text);
        setCopied(false);
      } catch {
        // invalid json
      }
    };
    reader.readAsText(file);
  }, []);

  return (
    <>
      {/* Кнопка экспорта — в панели ReactFlow controls */}
      <div className="react-flow__panel react-flow__controls vertical bottom left">
        <button
          onClick={handleExport}
          className="react-flow__controls-button bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] hover:bg-[var(--bg-tertiary)] hover:text-[var(--accent)] rounded-lg p-2 shadow-md transition-colors"
          title={t('workflowLibrary.exportWorkflow')}
        >
          <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
            <path d="M1 3a1 1 0 011-1h14a1 1 0 011 1v14a1 1 0 01-1 1H2a1 1 0 01-1-1V3zm1 0v14h14V3H2z"/>
            <path d="M9 5a.75.75 0 01.75.75v5.19l1.72-1.72a.75.75 0 111.06 1.06l-3 3a.75.75 0 01-1.06 0l-3-3a.75.75 0 111.06-1.06L7.25 10.94V5.75A.75.75 0 019 5z"/>
          </svg>
        </button>
      </div>

      {/* Модальное окно */}
      {isOpen && (
        <div
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50"
          onClick={() => setIsOpen(false)}
        >
          <div
            className="relative bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-[700px] max-h-[80vh] mx-4 flex flex-col overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--border)]">
              <h2 className="text-lg font-semibold text-[var(--text)]">{t('workflowLibrary.exportTitle')}</h2>
              <button onClick={() => setIsOpen(false)} className="p-1 hover:bg-[var(--background)] rounded-md transition-colors">
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <line x1="1" y1="1" x2="13" y2="13" />
                  <line x1="13" y1="1" x2="1" y2="13" />
                </svg>
              </button>
            </div>

            {/* Body */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4">
              <p className="text-sm text-[var(--text-secondary)]">{t('workflowLibrary.exportDesc')}</p>

              {/* Public toggle */}
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="export-public"
                  checked={isPublic}
                  onChange={(e) => {
                    setIsPublic(e.target.checked);
                    // Re-generate JSON with updated value
                    const workflowData = {
                      nodes,
                      edges,
                      author: user?.username || 'Anonymous',
                      is_public: e.target.checked,
                      exported_at: new Date().toISOString(),
                    };
                    setJsonText(JSON.stringify(workflowData, null, 2));
                    setCopied(false);
                  }}
                  className="w-4 h-4"
                />
                <label htmlFor="export-public" className="text-sm text-[var(--text)]">
                  {t('workflowLibrary.publishLabel')}
                </label>
              </div>

              {/* JSON textarea */}
              <div className="relative">
                <textarea
                  value={jsonText}
                  onChange={(e) => {
                    setJsonText(e.target.value);
                    setCopied(false);
                  }}
                  className="w-full h-64 bg-[var(--background)] border border-[var(--border)] rounded-lg p-3 font-mono text-xs text-[var(--text)] resize-none focus:outline-none focus:border-[var(--accent)]"
                  spellCheck={false}
                />
              </div>

              {/* File upload */}
              <div className="flex items-center gap-3">
                <label className="px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text)] cursor-pointer hover:bg-[var(--bg-tertiary)] transition-colors">
                  {t('workflowLibrary.chooseFile')}
                  <input
                    type="file"
                    accept=".json"
                    onChange={handleFileUpload}
                    className="hidden"
                  />
                </label>
                <span className="text-xs text-[var(--text-tertiary)]">
                  {t('workflowLibrary.importDesc')}
                </span>
              </div>
            </div>

            {/* Footer */}
            <div className="flex items-center justify-between px-6 py-4 border-t border-[var(--border)] gap-2">
              <div className="flex gap-2">
                <button
                  onClick={handleCopy}
                  className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                    copied
                      ? 'bg-green-600 text-white'
                      : 'bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white'
                  }`}
                >
                  {copied ? t('workflowLibrary.copied') : t('workflowLibrary.copyJson')}
                </button>
                <button
                  onClick={handleDownload}
                  className="px-4 py-2 bg-[var(--background)] border border-[var(--border)] text-[var(--text)] text-sm font-medium rounded-md hover:bg-[var(--bg-tertiary)] transition-colors"
                >
                  {t('workflowLibrary.downloadJson')}
                </button>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleImport}
                  className="px-4 py-2 bg-[var(--background)] border border-[var(--border)] text-[var(--text)] text-sm font-medium rounded-md hover:bg-[var(--bg-tertiary)] transition-colors"
                >
                  {t('workflowLibrary.importBtn')}
                </button>
                <button
                  onClick={() => setIsOpen(false)}
                  className="px-4 py-2 bg-[var(--background)] border border-[var(--border)] text-[var(--text)] text-sm font-medium rounded-md hover:bg-[var(--bg-tertiary)] transition-colors"
                >
                  {t('workflowLibrary.cancelBtn')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

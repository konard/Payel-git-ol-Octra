import { useState, useEffect, useRef } from 'react';
import { X, Star, Check } from 'lucide-react';
import { PROVIDERS, ALL_MODELS, type ProviderModel } from '../../config/providers';

interface ModelSelectorProps {
  selectedProvider: string;
  selectedModel: string;
  onSelect: (modelId: string) => void;
  isOpen: boolean;
  onClose: () => void;
  anchorRef: React.RefObject<HTMLElement | null>;
}

export function ModelSelector({
  selectedProvider,
  selectedModel,
  onSelect,
  isOpen,
  onClose,
  anchorRef,
}: ModelSelectorProps) {
  const [search, setSearch] = useState('');
  const modalRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // OpenRouter — ВСЕ модели, остальные — только свои
  const provider = PROVIDERS.find((p) => p.id === selectedProvider);
  const displayModels = selectedProvider === 'openrouter'
    ? ALL_MODELS.map((m) => ({ ...m, providerName: PROVIDERS.find((p) => p.id === m.providerId)?.name || '', providerColor: PROVIDERS.find((p) => p.id === m.providerId)?.color || '' }))
    : provider
      ? provider.models.map((m) => ({ ...m, providerName: provider.name, providerColor: provider.color }))
      : [];

  const filtered = displayModels.filter((m) => {
    if (!search) return true;
    const s = search.toLowerCase();
    return m.name.toLowerCase().includes(s) || m.id.toLowerCase().includes(s);
  });

  const [position, setPosition] = useState({ bottom: 0, left: 0 });

  useEffect(() => {
    if (isOpen && anchorRef.current) {
      const rect = anchorRef.current.getBoundingClientRect();
      setPosition({
        bottom: window.innerHeight - rect.top + 8,
        left: rect.left,
      });
    }
  }, [isOpen, anchorRef]);

  // Скролл к выбранной модели при открытии
  useEffect(() => {
    if (isOpen && selectedModel && listRef.current) {
      const selectedEl = listRef.current.querySelector('[data-selected="true"]');
      if (selectedEl) {
        setTimeout(() => {
          selectedEl.scrollIntoView({ block: 'center', behavior: 'instant' });
        }, 50);
      }
    }
  }, [isOpen, selectedModel]);

  useEffect(() => {
    if (!isOpen) return;
    const handler = (e: MouseEvent) => {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      ref={modalRef}
      className="fixed z-50 w-80 max-h-96 bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl overflow-hidden"
      style={{ bottom: `${position.bottom}px`, left: `${position.left}px` }}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--border)]">
        <h3 className="text-sm font-semibold text-[var(--text)]">Выбор модели</h3>
        <button onClick={onClose} className="p-1 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text-muted)]">
          <X size={14} />
        </button>
      </div>

      {/* Search */}
      <div className="px-3 py-2">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Поиск модели..."
          className="w-full px-3 py-1.5 bg-[var(--background)] border border-[var(--border)] rounded-lg text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
          autoFocus
        />
      </div>

      {/* Models list */}
      <div ref={listRef} className="overflow-y-auto max-h-72 px-2 pb-2">
        {filtered.length === 0 ? (
          <div className="text-center py-6 text-[var(--text-muted)] text-sm">Модели не найдены</div>
        ) : (
          <div className="space-y-0.5">
            {filtered.map((model) => {
              const isSelected = model.id === selectedModel;
              return (
                <button
                  key={model.id}
                  data-selected={isSelected ? 'true' : 'false'}
                  onClick={() => {
                    onSelect(model.id);
                    onClose();
                    setSearch('');
                  }}
                  className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors ${
                    isSelected
                      ? 'bg-[var(--accent)]/15 border border-[var(--accent)]/30'
                      : 'hover:bg-[var(--background)] border border-transparent'
                  }`}
                >
                  {/* Icon */}
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 overflow-hidden bg-[var(--background)]"
                  >
                    <img src={model.icon} alt={model.name} className="w-6 h-6 object-contain" />
                  </div>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5">
                      <span className="text-sm text-[var(--text)] truncate">{model.name}</span>
                      {model.recommended && (
                        <Star size={12} className="text-[var(--accent)] flex-shrink-0" fill="currentColor" />
                      )}
                    </div>
                  </div>

                  {/* Check */}
                  {isSelected && (
                    <div className="flex-shrink-0 text-[var(--accent)]">
                      <Check size={16} />
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

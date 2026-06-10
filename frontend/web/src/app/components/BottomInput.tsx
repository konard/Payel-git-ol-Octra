import { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import { ArrowRight, Settings2, Square, Search, ChevronDown, Puzzle, Paperclip, Globe, FileText, X } from 'lucide-react';
import { ModelSelector } from './ModelSelector';
import { PROVIDERS } from '../../config/providers';
import { useSettingsStore } from '../../stores/settingsStore';
import { useCustomProvidersStore } from '../../stores/customProvidersStore';
import { t } from '../../hooks/useI18n';

interface BottomInputProps {
  onSubmit: (data: TaskData) => void;
  onStop?: () => void;
  isSubmitting: boolean;
  isExpanded: boolean;
  onToggleExpand: () => void;
}

export interface TaskData {
  title: string;
  description: string;
  provider: string;
  model: string;
  apiKey: string;
  files: File[];
}

type AttachedFileItem = {
  id: string;
  file: File;
  previewUrl: string | null;
};

function createAttachmentId(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function createAttachedFileItem(file: File): AttachedFileItem {
  return {
    id: createAttachmentId(),
    file,
    previewUrl: file.type.startsWith('image/') ? URL.createObjectURL(file) : null,
  };
}

function revokeAttachmentPreview(attachment: AttachedFileItem) {
  if (attachment.previewUrl) {
    URL.revokeObjectURL(attachment.previewUrl);
  }
}

function formatFileSize(size: number): string {
  if (size < 1024) {
    return `${size} B`;
  }

  const units = ['KB', 'MB', 'GB'];
  let value = size / 1024;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${value >= 10 ? value.toFixed(0) : value.toFixed(1)} ${units[unitIndex]}`;
}

export function BottomInput({ onSubmit, onStop, isSubmitting, isExpanded, onToggleExpand }: BottomInputProps) {
  const [showModelSelector, setShowModelSelector] = useState(false);
  const [showProviderDropdown, setShowProviderDropdown] = useState(false);
  const [attachedFiles, setAttachedFiles] = useState<AttachedFileItem[]>([]);
  const modelInputRef = useRef<HTMLDivElement>(null);
  const providerBtnRef = useRef<HTMLButtonElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const attachedFilesRef = useRef<AttachedFileItem[]>([]);
  const hideApiKeyInput = useSettingsStore((state) => state.hideApiKeyInput);
  const defaultProvider = useSettingsStore((state) => state.defaultProvider);
  const defaultModel = useSettingsStore((state) => state.defaultModel);
  const setDefaultProvider = useSettingsStore((state) => state.setDefaultProvider);
  const setDefaultModel = useSettingsStore((state) => state.setDefaultModel);

  // Bubble animation state
  const [isAnimating, setIsAnimating] = useState(false);

  const createBubble = useCallback(() => {
    if (!buttonRef.current) return;
    const bubble = document.createElement('span');
    bubble.className = 'bubble';

    const randomX = Math.random() * 80 + 10;
    const randomY = Math.random() * 80 + 10;
    bubble.style.setProperty('--start-x', `${randomX}%`);
    bubble.style.setProperty('--start-y', `${randomY}%`);

    const size = Math.random() * 6 + 4;
    bubble.style.width = `${size}px`;
    bubble.style.height = `${size}px`;

    const wobble = (Math.random() - 0.5) * 15;
    bubble.style.setProperty('--wobble-dist', `${wobble}px`);

    const speed = Math.random() * 1.5 + 1.5;
    bubble.style.setProperty('--speed', `${speed}s`);

    buttonRef.current.appendChild(bubble);

    setTimeout(() => {
      if (bubble.parentNode) bubble.remove();
    }, speed * 1000);
  }, []);

  useEffect(() => {
    let intervalId: ReturnType<typeof setInterval> | null = null;
    if (isAnimating) {
      intervalId = setInterval(createBubble, 100);
    }
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [isAnimating, createBubble]);

  useEffect(() => {
    attachedFilesRef.current = attachedFiles;
  }, [attachedFiles]);

  useEffect(() => {
    return () => {
      attachedFilesRef.current.forEach(revokeAttachmentPreview);
    };
  }, []);

  // Custom providers
  const { providers: customProviders, models: customModels } = useCustomProvidersStore();

  // Combined providers list (static + custom)
  const allProviders = useMemo(() => {
    const combined = [...PROVIDERS];

    // Add custom providers with their models
    customProviders.forEach(customProvider => {
      if (!customProvider.id || !customProvider.name) {
        return;
      }

      const customProviderWithModels = {
        id: customProvider.id,
        name: customProvider.name,
        color: '#8b5cf6', // Purple color for custom providers
        bgColor: 'rgba(139, 92, 246, 0.15)',
        icon: '', // Custom providers use a rendered fallback icon.
        description: `Custom provider: ${customProvider.base_url}`,
        defaultModel: customModels.find(m => m.provider_id === customProvider.id)?.name || '',
        pricing: 'Custom',
        models: customModels
          .filter(model => model.provider_id === customProvider.id)
          .map(model => ({
            id: model.name, // Use model name as ID since custom models don't have standard IDs
            name: model.name,
            icon: '', // Custom models don't have specific icons
            free: false,
            recommended: false,
            providerId: customProvider.id,
          }))
      };
      combined.push(customProviderWithModels);
    });

    return combined;
  }, [customProviders, customModels]);

  const [formData, setFormData] = useState({
    title: '',
    description: '',
    provider: defaultProvider,
    model: defaultModel,
    apiKey: '',
  });

  // Sync formData with settings changes
  useEffect(() => {
    setFormData((prev) => ({
      ...prev,
      provider: defaultProvider,
      model: defaultModel,
    }));
  }, [defaultProvider, defaultModel]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.description.trim() || isSubmitting) return;

    const title = formData.title.trim() || formData.description.trim().slice(0, 50);

    await onSubmit({
      ...formData,
      title,
      apiKey: formData.apiKey.trim(),
      files: attachedFiles.map((attachment) => attachment.file),
    });

    setFormData({
      title: '',
      description: '',
      // Keep provider and model unchanged
      provider: formData.provider,
      model: formData.model,
      apiKey: '',
    });
    attachedFiles.forEach(revokeAttachmentPreview);
    setAttachedFiles([]);
    setShowModelSelector(false);
  };

  const handleProviderSelect = useCallback((providerId: string) => {
    const provider = allProviders.find(p => p.id === providerId);
    if (provider) {
      setFormData((prev) => ({ ...prev, provider: providerId, model: provider.defaultModel }));
      setDefaultProvider(providerId);
      setDefaultModel(provider.defaultModel);
    }
    setShowProviderDropdown(false);
  }, [allProviders, setDefaultProvider, setDefaultModel]);

  const handleModelSelect = useCallback((modelId: string) => {
    setFormData((prev) => ({ ...prev, model: modelId }));
    setDefaultModel(modelId);
  }, [setDefaultModel]);

  const handleFilesSelected = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length > 0) {
      setAttachedFiles((prev) => [...prev, ...files.map(createAttachedFileItem)]);
    }
    e.target.value = '';
  }, []);

  const removeAttachedFile = useCallback((attachmentId: string) => {
    setAttachedFiles((prev) => {
      const attachment = prev.find((item) => item.id === attachmentId);
      if (attachment) {
        revokeAttachmentPreview(attachment);
      }
      return prev.filter((item) => item.id !== attachmentId);
    });
  }, []);

  const selectedProvider = allProviders.find(p => p.id === formData.provider);

  // Submit on Enter (Shift+Enter inserts a newline), matching the reference chat
  // input behaviour.
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (!isSubmitting) handleSubmit(e as unknown as React.FormEvent);
    }
  };

  return (
    <div className="border-t border-[var(--border)] bg-[var(--surface)] p-3" data-tour="task-input">
      <form onSubmit={handleSubmit}>
        {/* A single unified, rounded input card — the task description, the
            provider/model selectors and the send button all live in one field,
            matching the reference design from issue #40. */}
        <div className="rounded-2xl border border-[var(--border)] bg-[var(--background)] px-3 pb-2 pt-3 transition-colors focus-within:border-[var(--accent)]">
          {attachedFiles.length > 0 && (
            <div className="mb-2 flex max-h-32 gap-2 overflow-x-auto pb-1" aria-label="Attached files">
              {attachedFiles.map((attachment) => (
                <div
                  key={attachment.id}
                  className="flex h-14 min-w-44 max-w-60 flex-shrink-0 items-center gap-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5"
                >
                  {attachment.previewUrl ? (
                    <img
                      src={attachment.previewUrl}
                      alt={attachment.file.name}
                      className="h-10 w-10 flex-shrink-0 rounded-md object-cover"
                    />
                  ) : (
                    <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-md bg-[var(--background)] text-[var(--text-muted)]">
                      <FileText size={18} />
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-xs font-medium text-[var(--text)]">{attachment.file.name}</div>
                    <div className="truncate text-[11px] text-[var(--text-muted)]">{formatFileSize(attachment.file.size)}</div>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeAttachedFile(attachment.id)}
                    className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md text-[var(--text-muted)] transition-colors hover:bg-[var(--background)] hover:text-[var(--text)]"
                    aria-label={`Remove ${attachment.file.name}`}
                    title={`Remove ${attachment.file.name}`}
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Row 1 — description + send */}
          <div className="flex items-end gap-2">
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              onKeyDown={handleKeyDown}
              rows={2}
              className="min-h-[2.5rem] flex-1 resize-none bg-transparent px-1 text-sm leading-6 text-[var(--text)] placeholder:text-[var(--text-muted)] focus:outline-none"
              placeholder={t('bottomInput.taskDescription')}
            />

            <button
              ref={buttonRef}
              type={isSubmitting ? 'button' : 'submit'}
              onClick={() => {
                if (!isSubmitting && !isAnimating) {
                  setIsAnimating(true);
                  setTimeout(() => setIsAnimating(false), 400);
                }
                if (isSubmitting && onStop) onStop();
              }}
              className={`flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-visible rounded-xl transition-colors ${
                isSubmitting
                  ? 'animate-pulse cursor-pointer bg-red-500 text-white hover:bg-red-600'
                  : 'bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)]'
              }`}
              title={isSubmitting ? t('bottomInput.settings') : undefined}
            >
              {isSubmitting ? <Square size={18} fill="white" /> : <ArrowRight size={20} />}
            </button>
          </div>

          {/* Row 2 — provider / model pills + secondary actions */}
          <div className="mt-1 flex items-center gap-2">
            {/* Provider pill */}
            <div className="relative">
              <button
                ref={providerBtnRef}
                type="button"
                onClick={() => {
                  setShowProviderDropdown(!showProviderDropdown);
                  setShowModelSelector(false);
                }}
                className="flex items-center gap-1.5 rounded-full border border-[var(--border)] bg-[var(--surface)] px-2.5 py-1 text-xs text-[var(--text)] transition-colors hover:border-[var(--accent)]"
              >
                {selectedProvider && (
                  selectedProvider.icon ? (
                    <img src={selectedProvider.icon} alt={selectedProvider.name} className="h-3.5 w-3.5 object-contain" />
                  ) : (
                    <Puzzle size={13} className="text-[var(--accent)]" />
                  )
                )}
                <span className="max-w-[8rem] truncate">{selectedProvider?.name || formData.provider}</span>
                <ChevronDown size={13} className={`text-[var(--text-muted)] transition-transform ${showProviderDropdown ? 'rotate-180' : ''}`} />
              </button>

              {showProviderDropdown && (
                <div className="absolute bottom-full left-0 mb-2 w-56 overflow-hidden rounded-lg border border-[var(--border)] bg-[var(--surface)] shadow-xl z-20">
                  <div className="max-h-64 overflow-y-auto p-2">
                    {allProviders.map((provider) => (
                      <button
                        key={provider.id}
                        type="button"
                        onClick={() => handleProviderSelect(provider.id)}
                        className={`flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors ${
                          formData.provider === provider.id
                            ? 'border border-[var(--accent)]/30 bg-[var(--accent)]/15'
                            : 'border border-transparent hover:bg-[var(--background)]'
                        }`}
                      >
                        <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center overflow-hidden rounded-lg bg-[var(--background)]">
                          {provider.icon ? (
                            <img src={provider.icon} alt={provider.name} className="h-6 w-6 object-contain" />
                          ) : (
                            <Puzzle size={16} className="text-[var(--accent)]" />
                          )}
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="text-sm font-medium text-[var(--text)]">{provider.name}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Model pill */}
            <div ref={modelInputRef} className="relative">
              <button
                type="button"
                onClick={() => {
                  setShowModelSelector(!showModelSelector);
                  setShowProviderDropdown(false);
                }}
                className="flex items-center gap-1.5 rounded-full border border-[var(--border)] bg-[var(--surface)] px-2.5 py-1 text-xs text-[var(--text)] transition-colors hover:border-[var(--accent)]"
              >
                <Search size={12} className="flex-shrink-0 text-[var(--text-muted)]" />
                <span className="max-w-[10rem] truncate">{formData.model}</span>
                <ChevronDown size={13} className={`text-[var(--text-muted)] transition-transform ${showModelSelector ? 'rotate-180' : ''}`} />
              </button>

              <ModelSelector
                selectedProvider={formData.provider}
                selectedModel={formData.model}
                providers={allProviders}
                customModels={customModels}
                onSelect={handleModelSelect}
                isOpen={showModelSelector}
                onClose={() => setShowModelSelector(false)}
                anchorRef={modelInputRef as React.RefObject<HTMLElement>}
              />
            </div>

            <div className="ml-auto flex items-center gap-0.5 text-[var(--text-muted)]">
              {!hideApiKeyInput && (
                <button
                  type="button"
                  onClick={onToggleExpand}
                  className={`rounded-lg p-1.5 transition-colors hover:bg-[var(--surface)] hover:text-[var(--text)] ${isExpanded ? 'text-[var(--accent)]' : ''}`}
                  title={t('bottomInput.settings')}
                >
                  <Settings2 size={16} />
                </button>
              )}
              <input
                type="file"
                multiple
                ref={fileInputRef}
                className="hidden"
                onChange={handleFilesSelected}
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="rounded-lg p-1.5 transition-colors hover:bg-[var(--surface)] hover:text-[var(--text)] relative"
                title="Attach files"
              >
                <Paperclip size={16} />
                {attachedFiles.length > 0 && (
                  <span className="absolute -right-0.5 -top-0.5 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-[var(--accent)] text-[9px] font-medium text-white">
                    {attachedFiles.length}
                  </span>
                )}
              </button>
              <span className="rounded-lg p-1.5" aria-hidden="true">
                <Globe size={16} />
              </span>
            </div>
          </div>

          {/* Advanced — API key, revealed only when the user opts in and the key
              field isn't globally hidden. */}
          {isExpanded && !hideApiKeyInput && (
            <div className="mt-2 border-t border-[var(--border)] pt-2">
              <label className="mb-1 block text-xs font-medium text-[var(--text-muted)]">
                {t('bottomInput.apiKey')}
              </label>
              <input
                type="password"
                value={formData.apiKey}
                onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
                className="w-full rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm text-[var(--text)] placeholder:text-[var(--text-muted)] transition-colors focus:border-[var(--accent)] focus:outline-none"
                placeholder={
                  formData.provider === 'openrouter' ? 'sk-or-v1-...'
                    : formData.provider === 'gemini' ? 'AIzaSy...'
                    : formData.provider === 'openai' ? 'sk-...'
                    : formData.provider === 'claude' ? 'sk-ant-...'
                    : formData.provider === 'deepseek' ? 'sk-...'
                    : 'xai-...'
                }
              />
            </div>
          )}
        </div>
      </form>
    </div>
  );
}

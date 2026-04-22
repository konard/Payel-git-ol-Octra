import { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import { Send, Settings2, Square, ChevronUp, ChevronDown, Search, ChevronRight } from 'lucide-react';
import { ModelSelector } from './ModelSelector';
import { PROVIDERS, getProviderById } from '../../config/providers';
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
}

export function BottomInput({ onSubmit, onStop, isSubmitting, isExpanded, onToggleExpand }: BottomInputProps) {
  const [showModelSelector, setShowModelSelector] = useState(false);
  const [showProviderDropdown, setShowProviderDropdown] = useState(false);
  const modelInputRef = useRef<HTMLDivElement>(null);
  const providerBtnRef = useRef<HTMLButtonElement>(null);
  const hideApiKeyInput = useSettingsStore((state) => state.hideApiKeyInput);
  const defaultToken = useSettingsStore((state) => state.defaultToken);
  const defaultProvider = useSettingsStore((state) => state.defaultProvider);
  const defaultModel = useSettingsStore((state) => state.defaultModel);
  const setDefaultProvider = useSettingsStore((state) => state.setDefaultProvider);
  const setDefaultModel = useSettingsStore((state) => state.setDefaultModel);

  // Custom providers
  const { providers: customProviders, models: customModels } = useCustomProvidersStore();

  // Combined providers list (static + custom)
  const allProviders = useMemo(() => {
    const combined = [...PROVIDERS];

    // Add custom providers with their models
    customProviders.forEach(customProvider => {
      const customProviderWithModels = {
        id: customProvider.id,
        name: customProvider.name,
        color: '#8b5cf6', // Purple color for custom providers
        bgColor: 'rgba(139, 92, 246, 0.15)',
        icon: '', // Custom providers don't have specific icons
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

    const apiKey = hideApiKeyInput ? defaultToken : formData.apiKey;

    await onSubmit({
      ...formData,
      title,
      apiKey,
    });

    setFormData({
      title: '',
      description: '',
      // Keep provider and model unchanged
      provider: formData.provider,
      model: formData.model,
      apiKey: '',
    });
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
  }, [setDefaultProvider, setDefaultModel]);

  const handleModelSelect = useCallback((modelId: string) => {
    setFormData((prev) => ({ ...prev, model: modelId }));
    setDefaultModel(modelId);
  }, [setDefaultModel]);

  const selectedProvider = allProviders.find(p => p.id === formData.provider);

  return (
    <div className="bg-[var(--surface)] border-t border-[var(--border)]">
      <div className="flex justify-center -mt-3 relative z-10">
        <button
          onClick={onToggleExpand}
          className="bg-[var(--surface)] border border-[var(--border)] rounded-full p-1.5 hover:bg-[var(--background)] transition-colors text-[var(--text-muted)]"
        >
          {isExpanded ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
        </button>
      </div>

      <form onSubmit={handleSubmit} className="p-4">
        {isExpanded && (
          <div className={`grid gap-3 mb-3 p-3 bg-[var(--background)] rounded-lg ${
            hideApiKeyInput ? 'grid-cols-2' : 'grid-cols-3'
          }`}>
            {/* Provider Dropdown */}
            <div className="relative">
              <label className="block text-xs font-medium text-[var(--text-muted)] mb-1">
                {t('bottomInput.provider')}
              </label>
              <button
                ref={providerBtnRef}
                type="button"
                onClick={() => {
                  setShowProviderDropdown(!showProviderDropdown);
                  setShowModelSelector(false);
                }}
                className="w-full flex items-center gap-2 px-2 py-1.5 bg-[var(--surface)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm hover:border-[var(--accent)] transition-colors"
              >
                {selectedProvider && (
                  <div className="w-5 h-5 flex items-center justify-center overflow-hidden">
                    <img src={selectedProvider.icon} alt={selectedProvider.name} className="w-5 h-5 object-contain" />
                  </div>
                )}
                <span className="flex-1 text-left truncate">{selectedProvider?.name || formData.provider}</span>
                <ChevronRight size={14} className={`transition-transform text-[var(--text-muted)] ${showProviderDropdown ? 'rotate-90' : ''}`} />
              </button>

              {/* Provider dropdown menu */}
              {showProviderDropdown && (
                <div className="absolute bottom-full left-0 right-0 mb-1 bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-xl overflow-hidden z-20">
                  <div className="p-2 max-h-64 overflow-y-auto">
                     {allProviders.map((provider) => (
                      <button
                        key={provider.id}
                        type="button"
                        onClick={() => handleProviderSelect(provider.id)}
                        className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors ${
                          formData.provider === provider.id
                            ? 'bg-[var(--accent)]/15 border border-[var(--accent)]/30'
                            : 'hover:bg-[var(--background)] border border-transparent'
                        }`}
                      >
                        <div
                          className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 overflow-hidden bg-[var(--background)]"
                        >
                          <img src={provider.icon} alt={provider.name} className="w-6 h-6 object-contain" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-sm text-[var(--text)] font-medium">{provider.name}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Model Selector */}
            <div ref={modelInputRef} className="relative">
              <label className="block text-xs font-medium text-[var(--text-muted)] mb-1">
                {t('bottomInput.model')}
              </label>
              <button
                type="button"
                onClick={() => {
                  setShowModelSelector(!showModelSelector);
                  setShowProviderDropdown(false);
                }}
                className="w-full flex items-center gap-2 px-2 py-1.5 bg-[var(--surface)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm hover:border-[var(--accent)] transition-colors"
              >
                <Search size={12} className="text-[var(--text-muted)] flex-shrink-0" />
                <span className="flex-1 text-left truncate">{formData.model}</span>
                <ChevronRight size={14} className={`transition-transform text-[var(--text-muted)] ${showModelSelector ? 'rotate-90' : ''}`} />
              </button>

              {/* Model selector modal */}
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

            {/* API Key */}
            {!hideApiKeyInput && (
              <div>
                <label className="block text-xs font-medium text-[var(--text-muted)] mb-1">
                  {t('bottomInput.apiKey')}
                </label>
                <input
                  type="password"
                  value={formData.apiKey}
                  onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
                  className="w-full px-2 py-1.5 bg-[var(--surface)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
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
        )}

        {!isExpanded && (
          <div className="flex items-center gap-2 mb-2">
            <button
              type="button"
              onClick={onToggleExpand}
              className="p-1.5 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text-muted)]"
              title={t('bottomInput.settings')}
            >
              <Settings2 size={16} />
            </button>
            <span className="text-xs text-[var(--text-muted)] flex items-center gap-1.5">
              {selectedProvider && (
                <img src={selectedProvider.icon} alt="" className="w-3.5 h-3.5 object-contain" />
              )}
              {formData.provider} • {formData.model}
            </span>
          </div>
        )}

        <div className="flex items-end gap-2">
          <div className="flex-1">
            {isExpanded && (
              <input
                type="text"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm mb-2 placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                placeholder={t('bottomInput.taskTitle')}
              />
            )}

            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={2}
              className="w-full px-4 py-3 bg-[var(--background)] border border-[var(--border)] rounded-xl text-[var(--text)] text-sm resize-none placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('bottomInput.taskDescription')}
            />
          </div>

          <button
            type={isSubmitting ? "button" : "submit"}
            disabled={false}
            onClick={isSubmitting ? onStop : undefined}
            className={`flex items-center justify-center w-12 h-12 rounded-xl transition-colors flex-shrink-0 ${
              isSubmitting
                ? 'bg-red-500 hover:bg-red-600 text-white animate-pulse cursor-pointer'
                : 'bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white'
            }`}
          >
            {isSubmitting ? (
              <Square size={20} fill="white" />
            ) : (
              <Send size={20} />
            )}
          </button>
        </div>
      </form>
    </div>
  );
}

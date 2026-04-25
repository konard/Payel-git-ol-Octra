import { useState, useEffect } from 'react';
import { Sun, Moon, Download, Settings, User, Key, Palette, Eye, Languages, LogOut, Crown, Puzzle, Plus, X, Edit, Trash2, MessageSquare } from 'lucide-react';
import { useTaskStore } from '../../stores/taskStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { useI18n, SUPPORTED_LANGUAGES, type LanguageCode } from '../../hooks/useI18n';
import { LANGUAGES_INFO } from '../../config/languages';
import { useAuthStore } from '../../stores/authStore';
import { useThemeStore } from '../../stores/themeStore';
import { useIntegrationStore, type IntegrationType } from '../../stores/integrationStore';
import { useCustomProvidersStore } from '../../stores/customProvidersStore';
import { customProviderService } from '../../services/customProviderService';
import { IntegrationCard } from '../../components/IntegrationCard';
import { CustomProviderCard } from '../../components/CustomProviderCard';
import { UserProfile } from '../../components/UserProfile';
import lefineIcon from '../../images/lefine.pro.jpg';
import telegramIcon from '../../images/Telegram.webp';
import n8nIcon from '../../images/n8n-color.png';
import crewaiMascot from '../../images/crewai-mascot.png';

type SettingsTab = 'api' | 'custom-providers' | 'custom-models' | 'language' | 'appearance' | 'visibility' | 'integrations';

// Add Model Form Component
function AddModelForm({ onSave, onCancel, providers }: {
  onSave: (model: { name: string; provider_id?: string }) => void;
  onCancel: () => void;
  providers: CustomProvider[];
}) {
  const { t } = useI18n();
  const [formData, setFormData] = useState({
    name: '',
    providerId: '',
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) return;

    onSave({
      name: formData.name.trim(),
      provider_id: formData.providerId || undefined,
    });
  };

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-[var(--text)] mb-2">
          {t('models.name')}
        </label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => handleInputChange('name', e.target.value)}
          className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
          placeholder={t('models.namePlaceholder') || 'Например: gpt-4, minimax-m2.7:cloud'}
          autoFocus
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-[var(--text)] mb-2">
          {t('models.provider')} ({t('common.optional')})
        </label>
        <select
          value={formData.providerId}
          onChange={(e) => handleInputChange('providerId', e.target.value)}
          className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm focus:outline-none focus:border-[var(--accent)] transition-colors"
        >
          <option value="">{t('models.noProvider')}</option>
          {providers.map((provider) => (
            <option key={provider.id} value={provider.id}>
              {provider.name}
            </option>
          ))}
        </select>
      </div>

      <div className="flex items-center gap-2 pt-2">
        <button
          type="submit"
          disabled={!formData.name.trim()}
          className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
        >
          {t('models.save')}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
        >
          {t('models.cancel')}
        </button>
      </div>
    </form>
  );
}

interface SettingsSection {
  id: SettingsTab;
  labelKey: string;
  icon: typeof Key;
}

const SETTINGS_SECTIONS: SettingsSection[] = [
  { id: 'api', labelKey: 'settings.apiTokens', icon: Key },
  { id: 'custom-providers', labelKey: 'settings.customProviders', icon: Puzzle },
  { id: 'custom-models', labelKey: 'settings.customModels', icon: Puzzle },
  { id: 'language', labelKey: 'settings.language', icon: Languages },
  { id: 'appearance', labelKey: 'settings.appearance', icon: Palette },
  { id: 'visibility', labelKey: 'settings.interface', icon: Eye },
  { id: 'integrations', labelKey: 'settings.integrations', icon: Puzzle },
];

interface TopBarProps {
  isAuthenticated: boolean;
  hasSubscription: boolean;
  onShowAuth: () => void;
  onShowSubscription: () => void;
  mode: 'canvas' | 'chat';
  onModeChange: (mode: 'canvas' | 'chat') => void;
  hasUnreadMessages: boolean;
  onToggleSidebar?: () => void;
}

export function TopBar({ isAuthenticated, hasSubscription, onShowAuth, onShowSubscription, mode, onModeChange, hasUnreadMessages, onToggleSidebar }: TopBarProps) {
  const status = useTaskStore((state) => state.status);
  const zipUrl = useTaskStore((state) => state.zipUrl);
  const [showSettings, setShowSettings] = useState(false);
  const [activeTab, setActiveTab] = useState<SettingsTab>('api');
  const defaultToken = useSettingsStore((state) => state.defaultToken);
  const hideApiKeyInput = useSettingsStore((state) => state.hideApiKeyInput);
  const hideServerStatus = useSettingsStore((state) => state.hideServerStatus);
  const hideConsole = useSettingsStore((state) => state.hideConsole);
  const setDefaultToken = useSettingsStore((state) => state.setDefaultToken);
  const setHideApiKeyInput = useSettingsStore((state) => state.setHideApiKeyInput);
  const setHideServerStatus = useSettingsStore((state) => state.setHideServerStatus);
  const setHideConsole = useSettingsStore((state) => state.setHideConsole);
  const { language, changeLanguage, t } = useI18n();
   const { isAuthenticated: isUserAuthenticated, logout } = useAuthStore();
   const { isDark, toggleTheme } = useThemeStore();
   const [showProfileMenu, setShowProfileMenu] = useState(false);
   const [showUserProfile, setShowUserProfile] = useState(false);

    // Custom providers state
    const { providers: customProviders, models: customModels, addProvider, updateProvider, deleteProvider, addModel, updateModel, deleteModel } = useCustomProvidersStore();
    const [showAddModel, setShowAddModel] = useState(false);
   const [showAddProvider, setShowAddProvider] = useState(false);
   const [editingProvider, setEditingProvider] = useState<string | null>(null);
  
  // Integration state
  const lefineIntegration = useIntegrationStore((state) => state.integrations.lefine);
  const telegramIntegration = useIntegrationStore((state) => state.integrations.telegram);
  const n8nIntegration = useIntegrationStore((state) => state.integrations.n8n);
  const setIntegrationConnected = useIntegrationStore((state) => state.setIntegrationConnected);
  const disconnectIntegration = useIntegrationStore((state) => state.disconnectIntegration);

  useEffect(() => {
    if (!showSettings) {
      setActiveTab('api');
    }
  }, [showSettings]);

  // Load custom providers and models on mount
  useEffect(() => {
    console.log('TopBar useEffect - isAuthenticated:', isAuthenticated);
    const loadCustomData = async () => {
      if (!isAuthenticated) {
        console.log('User not authenticated, skipping API calls');
        return;
      }

      try {
        const [providers, models] = await Promise.all([
          customProviderService.getUserCustomProviders(),
          customProviderService.getUserCustomModels()
        ]);

        console.log('API returned providers:', providers.length, 'models:', models.length);

        // Sync providers to store
        providers.forEach(provider => {
          if (!customProviders.find(p => p.id === provider.id)) {
            addProvider(provider);
          }
        });

        // Sync models to store
        models.forEach(model => {
          if (!customModels.find(m => m.id === model.id)) {
            addModel(model);
          }
        });
      } catch (error) {
        console.warn('API not available, using local store only:', error.message);
      }
    };

    loadCustomData();
  }, [isAuthenticated]); // Remove customProviders and customModels from dependencies to prevent infinite loop

  const handleOpenSettings = () => {
    setActiveTab('api');
    setShowSettings(true);
  };

  const handleDownload = () => {
    if (zipUrl) {
      const a = document.createElement('a');
      a.href = zipUrl;
      a.download = 'project.zip';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }
  };

  return (
    <>
      <header className="bg-[var(--surface)] border-b border-[var(--border)] px-4 py-3 flex items-center justify-between">
        {/* Left: Logo + Mode Switch */}
        <div className="flex items-center gap-3">
          {/* History & New Chat buttons */}
          <div className="flex items-center gap-1">
            <button
              onClick={onToggleSidebar}
              className="p-2 hover:bg-[var(--background)] rounded-lg transition-colors text-[var(--text-secondary)]"
              title="Chat History"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path fillRule="evenodd" clipRule="evenodd" d="M9.67272 0.522841C10.8339 0.522841 11.76 0.522714 12.4963 0.602493C13.2453 0.683657 13.8789 0.854248 14.4264 1.25197C14.7504 1.48739 15.0355 1.77247 15.2709 2.0965C15.6686 2.64394 15.8392 3.27758 15.9204 4.02655C16.0002 4.7629 16 5.68895 16 6.85014V9.14986C16 10.3111 16.0002 11.2371 15.9204 11.9735C15.8392 12.7224 15.6686 13.3561 15.2709 13.9035C15.0355 14.2275 14.7504 14.5126 14.4264 14.748C13.8789 15.1458 13.2453 15.3163 12.4963 15.3975C11.76 15.4773 10.8339 15.4772 9.67272 15.4772H6.3273C5.16611 15.4772 4.24006 15.4773 3.50371 15.3975C2.75474 15.3163 2.1211 15.1458 1.57366 14.748C1.24963 14.5126 0.964549 14.2275 0.729131 13.9035C0.331407 13.3561 0.160817 12.7224 0.0796529 11.9735C-0.000126137 11.2371 1.25338e-09 10.3111 1.25338e-09 9.14986V6.85014C1.25329e-09 5.68895 -0.000126137 4.7629 0.0796529 4.02655C0.160817 3.27758 0.331407 2.64394 0.729131 2.0965C0.964549 1.77247 1.24963 1.48739 1.57366 1.25197C2.1211 0.854248 2.75474 0.683657 3.50371 0.602493C4.24006 0.522714 5.16611 0.522841 6.3273 0.522841H9.67272ZM5.54303 1.88715V14.1118C5.78636 14.1128 6.04709 14.1169 6.3273 14.1169H9.67272C10.8639 14.1169 11.7032 14.1164 12.3493 14.0465C12.9824 13.9779 13.3497 13.8494 13.6268 13.6482C13.8354 13.4966 14.0195 13.3125 14.1711 13.1039C14.3723 12.8268 14.5007 12.4595 14.5693 11.8264C14.6393 11.1803 14.6398 10.341 14.6398 9.14986V6.85014C14.6398 5.65896 14.6393 4.81967 14.5693 4.1736C14.5007 3.54048 14.3723 3.17318 14.1711 2.89609C14.0195 2.68747 13.8354 2.50337 13.6268 2.35179C13.3497 2.1506 12.9824 2.02212 12.3493 1.95353C11.7032 1.88358 10.8639 1.88307 9.67272 1.88307H6.3273C6.04709 1.88307 5.78636 1.8862 5.54303 1.88715ZM4.1828 1.91166C3.99125 1.9216 3.8148 1.93577 3.65076 1.95353C3.01764 2.02212 2.65034 2.1506 2.37325 2.35179C2.16463 2.50337 1.98052 2.68747 1.82895 2.89609C1.62776 3.17318 1.49928 3.54048 1.43069 4.1736C1.36074 4.81967 1.36023 5.65896 1.36023 6.85014V9.14986C1.36023 10.341 1.36074 11.1803 1.43069 11.8264C1.49928 12.4595 1.62776 12.8268 1.82895 13.1039C1.98052 13.3125 2.16463 13.4966 2.37325 13.6482C2.65034 13.8494 3.01764 13.9779 3.65076 14.0465C3.81478 14.0642 3.99127 14.0774 4.1828 14.0873V1.91166Z" fill="currentColor"/>
              </svg>
            </button>
            <button
              onClick={() => {
                onToggleSidebar();
                // Trigger new chat after sidebar opens
              }}
              className="p-2 hover:bg-[var(--background)] rounded-lg transition-colors text-[var(--text-secondary)]"
              title="New Chat"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M8 0.599609C3.91309 0.599609 0.599609 3.91309 0.599609 8C0.599609 9.13376 0.855461 10.2098 1.3125 11.1719L1.5918 11.7588L2.76562 11.2012L2.48633 10.6143C2.11034 9.82278 1.90039 8.93675 1.90039 8C1.90039 4.63106 4.63106 1.90039 8 1.90039C11.3689 1.90039 14.0996 4.63106 14.0996 8C14.0996 11.3689 11.3689 14.0996 8 14.0996C7.31041 14.0996 6.80528 14.0514 6.35742 13.9277C5.91623 13.8059 5.49768 13.6021 4.99707 13.2529C4.26492 12.7422 3.21611 12.5616 2.35156 13.1074L2.33789 13.1162L2.32422 13.126L1.58789 13.6436L2.01953 14.9297L3.0459 14.207C3.36351 14.0065 3.83838 14.0294 4.25293 14.3184C4.84547 14.7317 5.39743 15.011 6.01172 15.1807C6.61947 15.3485 7.25549 15.4004 8 15.4004C12.0869 15.4004 15.4004 12.0869 15.4004 8C15.4004 3.91309 12.0869 0.599609 8 0.599609ZM7.34473 4.93945V7.34961H4.93945V8.65039H7.34473V11.0605H8.64551V8.65039H11.0605V7.34961H8.64551V4.93945H7.34473Z" fill="currentColor"/>
              </svg>
            </button>
          </div>
          <div style={{ width: 3 }}></div>
          <img
            src={crewaiMascot}
            alt="CrewAI Mascot"
            className="w-10 h-10 rounded-lg object-contain"
          />
          <h1 className="text-lg font-semibold text-[var(--text)]">CrewAI</h1>
          {/* Переключатель режимов */}
          <div className="border border-[var(--border)] rounded-lg overflow-hidden shadow-sm ml-4">
            <button
              onClick={() => onModeChange('canvas')}
              className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                mode === 'canvas'
                  ? 'bg-[var(--accent)] text-white'
                  : 'bg-[var(--surface)] text-[var(--text)] hover:bg-[var(--background)]'
              }`}
            >
              Canvas
            </button>
            <button
              onClick={() => onModeChange('chat')}
              className={`px-3 py-1.5 text-sm font-medium transition-colors relative ${
                mode === 'chat'
                  ? 'bg-[var(--accent)] text-white'
                  : 'bg-[var(--surface)] text-[var(--text)] hover:bg-[var(--background)]'
              }`}
            >
              Chat
            </button>
          </div>
        </div>

        {/* Center: Download button */}
        <div className="flex items-center gap-2">
          {status === 'done' && zipUrl && (
            <button
              onClick={handleDownload}
              className="flex items-center gap-2 px-3 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md text-sm font-medium transition-colors"
            >
              <Download size={16} />
              Скачать ZIP
            </button>
          )}
        </div>

        {/* Right: Theme & Settings & Profile */}
        <div className="flex items-center gap-2">
          <button
            onClick={toggleTheme}
            className="p-2 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text)]"
            title={isDark ? t('topbar.lightTheme') : t('topbar.darkTheme')}
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </button>
          <button
            onClick={handleOpenSettings}
            className={`p-2 rounded-md transition-colors text-[var(--text)] ${
              showSettings ? 'bg-[var(--accent)]/15 text-[var(--accent)]' : 'hover:bg-[var(--background)]'
            }`}
            title={t('topbar.settings')}
          >
            <Settings size={18} />
          </button>
          {isAuthenticated ? (
            <div className="relative">
              <button
                onClick={() => {
                  setShowUserProfile(true);
                  setShowProfileMenu(false);
                }}
                className="p-2 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text)]"
                title={t('topbar.profile')}
              >
                <User size={18} />
              </button>
            </div>
          ) : (
            <button
              onClick={onShowAuth}
              className="px-3 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
              title={t('auth.login')}
            >
              {t('auth.login')}
            </button>
          )}
          {isAuthenticated && !hasSubscription && (
            <button
              onClick={onShowSubscription}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white text-sm font-semibold rounded-full transition-all shadow-md hover:shadow-lg hover:shadow-orange-500/25"
              title="Оформить подписку Pro"
            >
              <Crown size={14} />
              <span>Pro</span>
            </button>
          )}
        </div>
      </header>

      {/* Settings Modal (JetBrains style) */}
      {showSettings && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          onClick={() => setShowSettings(false)}
        >
          <div
            className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-[700px] max-w-[95vw] h-[420px] flex overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Left Sidebar */}
            <div className="w-52 border-r border-[var(--border)] bg-[var(--background)] flex-shrink-0 flex flex-col py-3">
              {SETTINGS_SECTIONS.map((section) => {
                const Icon = section.icon;
                const isActive = activeTab === section.id;
                return (
                  <button
                    key={section.id}
                    onClick={() => setActiveTab(section.id)}
                    className={`flex items-center gap-2.5 px-3 py-2 mx-1 rounded-md text-sm transition-colors text-left ${
                      isActive
                        ? 'bg-[var(--accent)]/15 text-[var(--accent)] font-medium'
                        : 'text-[var(--text)] hover:bg-[var(--surface)]'
                    }`}
                  >
                    <Icon size={16} />
                    <span className="truncate">{t(section.labelKey)}</span>
                  </button>
                );
              })}
            </div>

            {/* Right Content */}
            <div className="flex-1 flex flex-col">
              {/* Header */}
              <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--border)]">
                <h2 className="text-base font-semibold text-[var(--text)]">
                  {t(SETTINGS_SECTIONS.find((s) => s.id === activeTab)?.labelKey || '')}
                </h2>
                <button
                  onClick={() => setShowSettings(false)}
                  className="p-1 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text-muted)]"
                >
                  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                    <line x1="1" y1="1" x2="13" y2="13" />
                    <line x1="13" y1="1" x2="1" y2="13" />
                  </svg>
                </button>
              </div>

              {/* Settings Content */}
              <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5">
                {activeTab === 'api' && (
                  <div>
                    <div>
                      <label className="block text-sm font-medium text-[var(--text)] mb-2">
                        {t('settings.defaultToken')}
                      </label>
                      <input
                        type="password"
                        value={defaultToken}
                        onChange={(e) => setDefaultToken(e.target.value)}
                        className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                        placeholder={t('settings.defaultTokenPlaceholder')}
                      />
                      <p className="mt-1.5 text-xs text-[var(--text-muted)]">
                        {t('settings.defaultTokenHint')}
                      </p>
                    </div>
                  </div>
                 )}

                 {activeTab === 'custom-providers' && (
                   <div className="space-y-4 max-w-lg">
                     <div>
                       <div className="text-sm font-medium text-[var(--text)] mb-1">{t('providers.title')}</div>
                       <div className="text-xs text-[var(--text-muted)] mb-4">
                         {t('providers.description')}
                       </div>
                     </div>

                      {/* Existing providers */}
                      {customProviders.map((provider) => (
                        <CustomProviderCard
                          key={provider.id}
                          provider={provider}
                          isEditing={editingProvider === provider.id}
                          onSave={(updates) => {
                            updateProvider(provider.id, updates);
                            setEditingProvider(null);
                          }}
                          onCancel={() => setEditingProvider(null)}
                          onEdit={() => {
                            console.log('Edit clicked for provider:', provider.id);
                            setEditingProvider(provider.id);
                            console.log('editingProvider set to:', provider.id);
                          }}
                          onDelete={() => deleteProvider(provider.id)}
                        />
                      ))}

                     {/* Add new provider */}
                     {showAddProvider && (
                       <CustomProviderCard
                         isNew
                         onSave={(providerData) => {
                           addProvider(providerData);
                           setShowAddProvider(false);
                         }}
                         onCancel={() => setShowAddProvider(false)}
                       />
                     )}

                     {/* Add button */}
                     {!showAddProvider && (
                       <button
                         onClick={() => setShowAddProvider(true)}
                         className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[var(--background)] border border-dashed border-[var(--border)] rounded-lg text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--accent)] transition-colors"
                       >
                         <Plus size={16} />
                         {t('providers.addNew')}
                       </button>
                     )}
                    </div>
                  )}

                  {activeTab === 'custom-models' && (
                    <div className="space-y-4 max-w-lg">
                      <div>
                        <div className="text-sm font-medium text-[var(--text)] mb-1">{t('models.title')}</div>
                        <div className="text-xs text-[var(--text-muted)] mb-4">
                          {t('models.description')}
                        </div>
                      </div>

                       {/* Existing models */}
                       {customModels.map((model) => (
                         <div key={model.id} className="flex items-center gap-3 p-3 bg-[var(--background)] border border-[var(--border)] rounded-lg">
                          <div className="flex-1">
                            <div className="text-sm font-medium text-[var(--text)]">{model.name}</div>
                            {model.provider_id && (
                              <div className="text-xs text-[var(--text-muted)]">
                                Provider: {customProviders.find(p => p.id === model.provider_id)?.name || 'Unknown'}
                              </div>
                            )}
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => {
                                const newName = prompt('Enter new model name:', model.name);
                                if (newName && newName.trim() && newName.trim() !== model.name) {
                                  updateModel(model.id, { name: newName.trim() });
                                  // Try to update on API if authenticated
                                  if (isAuthenticated) {
                                    customProviderService.updateCustomModel(model.id, { name: newName.trim() })
                                      .catch(error => console.warn('API update failed:', error));
                                  }
                                }
                              }}
                              className="p-1 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text-muted)]"
                              title="Edit"
                            >
                              <Edit size={14} />
                            </button>
                            <button
                              onClick={async () => {
                                // Always update local store first
                                deleteModel(model.id);

                                // Then try to delete from API if authenticated
                                if (isAuthenticated) {
                                  try {
                                    await customProviderService.deleteCustomModel(model.id);
                                  } catch (error) {
                                    console.warn('API not available, deleted locally only');
                                  }
                                } else {
                                  console.log('User not authenticated, deleted locally only');
                                }
                              }}
                              className="p-1 hover:bg-red-500/20 rounded-md transition-colors text-red-500"
                              title={t('models.delete')}
                            >
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </div>
                      ))}

                      {/* Add new model */}
                      <button
                        onClick={() => setShowAddModel(true)}
                        className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[var(--background)] border border-dashed border-[var(--border)] rounded-lg text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--accent)] transition-colors"
                      >
                        <Plus size={16} />
                        {t('models.addNew')}
                      </button>
                    </div>
                  )}

                  {activeTab === 'language' && (
                  <div className="space-y-4">
                    <div>
                      <div className="text-sm font-medium text-[var(--text)] mb-1">{t('settings.interfaceLanguage')}</div>
                      <div className="text-xs text-[var(--text-muted)] mb-3">
                        {t('settings.languageHint')}
                      </div>
                      <div className="grid grid-cols-2 gap-2 max-h-[280px] overflow-y-auto pr-1">
                        {SUPPORTED_LANGUAGES.map((code) => {
                          const isActive = language === code;
                          const langInfo = LANGUAGES_INFO[code as LanguageCode];

                          const flagSvg = langInfo ? (
                            <img
                              src={`https://flagcdn.com/w40/${langInfo.flag.toLowerCase()}.png`}
                              alt={langInfo.name}
                              className="w-5 h-4 object-cover rounded-sm"
                              loading="lazy"
                            />
                          ) : null;

                          return (
                            <button
                              key={code}
                              onClick={() => changeLanguage(code as LanguageCode)}
                              className={`flex items-center gap-2 px-3 py-2.5 rounded-lg border text-left transition-all ${
                                isActive
                                  ? 'border-[var(--accent)] bg-[var(--accent)]/10 text-[var(--accent)]'
                                  : 'border-[var(--border)] bg-[var(--background)] hover:border-[var(--accent)]/50'
                              }`}
                            >
                              {flagSvg && <span className="w-5 flex-shrink-0">{flagSvg}</span>}
                              <span className="text-sm font-medium">{langInfo?.nativeName || code}</span>
                              {isActive && <span className="ml-auto text-xs">✓</span>}
                            </button>
                          );
                        })}
                       </div>
                     </div>
                   </div>
                 )}

                {activeTab === 'appearance' && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm font-medium text-[var(--text)]">{t('settings.themeTitle')}</div>
                        <div className="text-xs text-[var(--text-muted)]">
                          {t('settings.themeCurrent')}: {isDark ? t('settings.themeDark') : t('settings.themeLight')}
                        </div>
                      </div>
                      <button
                        onClick={toggleTheme}
                        className="px-3 py-1.5 text-xs font-medium bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] hover:border-[var(--accent)] transition-colors"
                      >
                        {t('settings.themeToggle')}
                      </button>
                    </div>
                  </div>
                )}

                {activeTab === 'visibility' && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm font-medium text-[var(--text)]">
                          {t('settings.hideTokenField')}
                        </div>
                        <div className="text-xs text-[var(--text-muted)]">
                          {t('settings.hideTokenFieldHint')}
                        </div>
                      </div>
                      <button
                        onClick={() => setHideApiKeyInput(!hideApiKeyInput)}
                        className={`relative w-12 h-7 rounded-full transition-colors duration-200 ${
                          hideApiKeyInput ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
                        }`}
                      >
                        <span
                          className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform duration-200 ${
                            hideApiKeyInput ? 'translate-x-5' : 'translate-x-0'
                          }`}
                        />
                      </button>
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm font-medium text-[var(--text)]">
                          Скрыть консоль
                        </div>
                        <div className="text-xs text-[var(--text-muted)]">
                          Скрывает нижнюю панель логов и событий.
                        </div>
                      </div>
                      <button
                        onClick={() => setHideConsole(!hideConsole)}
                        className={`relative w-12 h-7 rounded-full transition-colors duration-200 ${
                          hideConsole ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
                        }`}
                      >
                        <span
                          className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform duration-200 ${
                            hideConsole ? 'translate-x-5' : 'translate-x-0'
                          }`}
                        />
                      </button>
                    </div>
                  </div>
                )}

                {activeTab === 'integrations' && (
                  <div className="space-y-4 max-w-lg">
                    <div>
                      <div className="text-sm font-medium text-[var(--text)] mb-1">{t('integrations.title')}</div>
                      <div className="text-xs text-[var(--text-muted)] mb-4">
                        {t('integrations.description')}
                      </div>
                    </div>

                    <IntegrationCard
                      type="lefine"
                      name={t('integrations.lefine.name')}
                      description={t('integrations.lefine.description')}
                      icon={lefineIcon}
                      connected={lefineIntegration.connected}
                      config={lefineIntegration.config}
                      onConnect={(config) => setIntegrationConnected('lefine', true, config)}
                      onDisconnect={() => disconnectIntegration('lefine')}
                    />

                    <IntegrationCard
                      type="telegram"
                      name={t('integrations.telegram.name')}
                      description={t('integrations.telegram.description')}
                      icon={telegramIcon}
                      connected={telegramIntegration.connected}
                      config={telegramIntegration.config}
                      onConnect={(config) => setIntegrationConnected('telegram', true, config)}
                      onDisconnect={() => disconnectIntegration('telegram')}
                    />

                    <IntegrationCard
                      type="n8n"
                      name={t('integrations.n8n.name')}
                      description={t('integrations.n8n.description')}
                      icon={n8nIcon}
                      connected={n8nIntegration.connected}
                      config={n8nIntegration.config}
                      onConnect={(config) => setIntegrationConnected('n8n', true, config)}
                      onDisconnect={() => disconnectIntegration('n8n')}
                    />
                  </div>
                )}


              </div>
            </div>
          </div>
        </div>
      )}

      {/* Add Model Modal */}
      {showAddModel && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          onClick={() => setShowAddModel(false)}
        >
          <div
            className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-[400px] max-w-[95vw] p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-[var(--text)]">{t('models.addNew')}</h3>
              <button
                onClick={() => setShowAddModel(false)}
                className="p-1 hover:bg-[var(--background)] rounded-md transition-colors text-[var(--text-muted)]"
              >
                <X size={18} />
              </button>
            </div>

            <AddModelForm
              onSave={async (modelData) => {
                console.log('Saving model to API:', modelData, 'isAuthenticated:', isAuthenticated);
                // Save locally first
                addModel(modelData);

                // Only try API if user is authenticated
                if (isAuthenticated) {
                  // Convert provider_id to proper format for API
                  const apiData = {
                    name: modelData.name,
                    provider_id: modelData.provider_id ? modelData.provider_id : null
                  };

                  // Then try to save to API
                  try {
                    const result = await customProviderService.createCustomModel(apiData);
                    console.log('API response:', result);
                  } catch (error) {
                    console.warn('API not available, saved locally only:', error);
                  }
                } else {
                  console.log('User not authenticated, saved locally only');
                }

                setShowAddModel(false);
              }}
              onCancel={() => setShowAddModel(false)}
              providers={customProviders}
            />
          </div>
        </div>
      )}

      {/* User Profile Modal */}
      {showUserProfile && isAuthenticated && (
        <UserProfile onClose={() => setShowUserProfile(false)} />
      )}
    </>
  );
}

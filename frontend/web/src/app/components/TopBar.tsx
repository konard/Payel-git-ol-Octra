import { useState, useEffect } from 'react';
import { Sun, Moon, Download, Settings, User, Key, Palette, Eye, Languages, LogOut, Crown, Puzzle } from 'lucide-react';
import { useTaskStore } from '../../stores/taskStore';
import { useSettingsStore } from '../../stores/settingsStore';
import { useI18n, SUPPORTED_LANGUAGES, type LanguageCode } from '../../hooks/useI18n';
import { LANGUAGES_INFO } from '../../config/languages';
import { useAuthStore } from '../../stores/authStore';
import { useThemeStore } from '../../stores/themeStore';
import { useIntegrationStore, type IntegrationType } from '../../stores/integrationStore';
import { IntegrationCard } from '../../components/IntegrationCard';
import { UserProfile } from '../../components/UserProfile';
import lefineIcon from '../../images/lefine.pro.jpg';
import telegramIcon from '../../images/Telegram.webp';

type SettingsTab = 'api' | 'language' | 'appearance' | 'visibility' | 'integrations';

interface SettingsSection {
  id: SettingsTab;
  labelKey: string;
  icon: typeof Key;
}

const SETTINGS_SECTIONS: SettingsSection[] = [
  { id: 'api', labelKey: 'settings.apiTokens', icon: Key },
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
}

export function TopBar({ isAuthenticated, hasSubscription, onShowAuth, onShowSubscription }: TopBarProps) {
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
  
  // Integration state
  const lefineIntegration = useIntegrationStore((state) => state.integrations.lefine);
  const telegramIntegration = useIntegrationStore((state) => state.integrations.telegram);
  const setIntegrationConnected = useIntegrationStore((state) => state.setIntegrationConnected);
  const disconnectIntegration = useIntegrationStore((state) => state.disconnectIntegration);

  useEffect(() => {
    if (!showSettings) {
      setActiveTab('api');
    }
  }, [showSettings]);

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
        {/* Left: Logo */}
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">C</span>
          </div>
          <h1 className="text-lg font-semibold text-[var(--text)]">CrewAI</h1>
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
                          
                          // SVG flag emoji using country code
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
                              {isActive && (
                                <span className="ml-auto text-xs">✓</span>
                              )}
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
                          {t('settings.hideServerStatus')}
                        </div>
                        <div className="text-xs text-[var(--text-muted)]">
                          {t('settings.hideServerStatusHint')}
                        </div>
                      </div>
                      <button
                        onClick={() => setHideServerStatus(!hideServerStatus)}
                        className={`relative w-12 h-7 rounded-full transition-colors duration-200 ${
                          hideServerStatus ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
                        }`}
                      >
                        <span
                          className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform duration-200 ${
                            hideServerStatus ? 'translate-x-5' : 'translate-x-0'
                          }`}
                        />
                      </button>
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-sm font-medium text-[var(--text)]">
                          {t('settings.hideConsole')}
                        </div>
                        <div className="text-xs text-[var(--text-muted)]">
                          {t('settings.hideConsoleHint')}
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
                    
                    {/* Lefine.pro Integration */}
                    <IntegrationCard
                      type="lefine"
                      name={t('integrations.lefine.name')}
                      description={t('integrations.lefine.description')}
                      icon={lefineIcon}
                      connected={lefineIntegration.connected}
                      config={lefineIntegration.config}
                      onConnect={(config) => {
                        setIntegrationConnected('lefine', true, config);
                      }}
                      onDisconnect={() => {
                        disconnectIntegration('lefine');
                      }}
                    />

                    {/* Telegram Integration */}
                    <IntegrationCard
                      type="telegram"
                      name={t('integrations.telegram.name')}
                      description={t('integrations.telegram.description')}
                      icon={telegramIcon}
                      connected={telegramIntegration.connected}
                      config={telegramIntegration.config}
                      onConnect={(config) => {
                        setIntegrationConnected('telegram', true, config);
                      }}
                      onDisconnect={() => {
                        disconnectIntegration('telegram');
                      }}
                    />
                  </div>
                )}
              </div>
            </div>
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

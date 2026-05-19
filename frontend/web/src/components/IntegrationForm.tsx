import { useState, useEffect } from 'react';
import { Check } from 'lucide-react';
import { useI18n } from '../hooks/useI18n';
import { useSettingsStore } from '../stores/settingsStore';
import { useIntegrationStore } from '../stores/integrationStore';
import type { IntegrationType, IntegrationConfig } from '../stores/integrationStore';
import { getMyWorkflows, type Workflow } from '../services/workflowService';

interface IntegrationFormProps {
  type: IntegrationType;
  initialConfig: IntegrationConfig;
  onSave: (config: IntegrationConfig) => void;
  onCancel: () => void;
  connected: boolean;
}

export function IntegrationForm({
  type,
  initialConfig,
  onSave,
  onCancel,
  connected,
}: IntegrationFormProps) {
  const { t } = useI18n();
  const defaultToken = useSettingsStore((state) => state.defaultToken);
  const [useDefaultKey, setUseDefaultKey] = useState(initialConfig.useDefaultKey);
  const [apiKey, setApiKey] = useState(initialConfig.apiKey || '');
  const [workflowId, setWorkflowId] = useState(initialConfig.workflowId || '');
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loadingWorkflows, setLoadingWorkflows] = useState(false);
  
  // Lefine.pro specific
  const [activityPubUrl, setActivityPubUrl] = useState(initialConfig.activityPubUrl || 'https://exchange.lefine.pro');
  const [outboxEndpoint, setOutboxEndpoint] = useState(initialConfig.outboxEndpoint || 'https://exchange.lefine.pro/outbox');
  const [inboxEndpoint, setInboxEndpoint] = useState(initialConfig.inboxEndpoint || 'https://exchange.lefine.pro/inbox');

  // Load workflows on mount (only for integrations that need workflows)
  useEffect(() => {
    if (type === 'github') return; // GitHub doesn't need workflows

    const loadWorkflows = async () => {
      setLoadingWorkflows(true);
      try {
        const myWorkflows = await getMyWorkflows();
        setWorkflows(myWorkflows);
      } catch (error) {
        console.error('Failed to load workflows:', error);
      } finally {
        setLoadingWorkflows(false);
      }
    };
    loadWorkflows();
  }, [type]);

  const handleSave = () => {
    const config: IntegrationConfig = {
      useDefaultKey: type === 'n8n' ? false : useDefaultKey,
      apiKey: type === 'n8n' ? apiKey : (useDefaultKey ? defaultToken : apiKey),
      workflowId: type === 'n8n' ? '' : workflowId,
      ...(type === 'lefine' && {
        activityPubUrl,
        outboxEndpoint,
        inboxEndpoint,
      }),
      ...(type === 'n8n' && {
        activityPubUrl, // Server URL for n8n
      }),
      ...(type === 'github' && {
        publishNewProjects: initialConfig.publishNewProjects,
        createPullRequests: initialConfig.createPullRequests,
      }),
    };
    onSave(config);
  };

  // GitHub Integration
  if (type === 'github') {
    const [publishNewProjects, setPublishNewProjects] = useState<boolean>(initialConfig.publishNewProjects ?? false);
    const [createPullRequests, setCreatePullRequests] = useState<boolean>(initialConfig.createPullRequests ?? false);

    const handleGitHubSave = () => {
      const config: IntegrationConfig = {
        useDefaultKey: false,
        apiKey: '',
        workflowId: '',
        publishNewProjects,
        createPullRequests,
      };
      onSave(config);
    };

    return (
      <div className="space-y-6">
        <div>
          <div className="flex items-center gap-3 mb-4">
            <img 
              src="/assets/github-image.png" 
              alt="GitHub" 
              className="w-8 h-8" 
            />
            <div>
              <h3 className="font-semibold text-[var(--text)]">GitHub Integration</h3>
              <p className="text-sm text-[var(--text-muted)]">Автоматизация работы с репозиториями</p>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          {/* Toggle 1 */}
          <div className="flex items-center justify-between p-4 bg-[var(--background)] border border-[var(--border)] rounded-xl">
            <div className="flex-1 pr-4">
              <div className="font-medium text-[var(--text)]">Публиковать на GitHub новые проекты</div>
              <div className="text-xs text-[var(--text-muted)] mt-1">
                При создании нового workflow автоматически создавать репозиторий на GitHub
              </div>
            </div>
            <button
              onClick={() => setPublishNewProjects(!publishNewProjects)}
              className={`relative w-12 h-7 rounded-full transition-all flex-shrink-0 ${
                publishNewProjects ? 'bg-green-500' : 'bg-[var(--border)]'
              }`}
            >
              <span className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform ${publishNewProjects ? 'translate-x-5' : ''}`} />
            </button>
          </div>

          {/* Toggle 2 */}
          <div className="flex items-center justify-between p-4 bg-[var(--background)] border border-[var(--border)] rounded-xl">
            <div className="flex-1 pr-4">
              <div className="font-medium text-[var(--text)]">Делать pull requests</div>
              <div className="text-xs text-[var(--text-muted)] mt-1">
                Автоматически создавать Pull Request при обновлении проектов
              </div>
            </div>
            <button
              onClick={() => setCreatePullRequests(!createPullRequests)}
              className={`relative w-12 h-7 rounded-full transition-all flex-shrink-0 ${
                createPullRequests ? 'bg-green-500' : 'bg-[var(--border)]'
              }`}
            >
              <span className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform ${createPullRequests ? 'translate-x-5' : ''}`} />
            </button>
          </div>
        </div>

        <div className="flex gap-2 pt-2">
          <button
            onClick={handleGitHubSave}
            className="flex-1 px-4 py-2.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-lg transition-colors"
          >
            {connected ? 'Сохранить настройки' : 'Интегрировать с GitHub'}
          </button>
          <button
            onClick={onCancel}
            className="px-4 py-2.5 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-lg transition-colors border border-[var(--border)]"
          >
            Отмена
          </button>
        </div>
      </div>
    );
  }

  // Telegram bot link
  if (type === 'telegram') {
    const telegramBotUrl = 'https://t.me/CrewAIBot'; // Замените на ваш бот URL
    
    return (
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-[var(--text)] mb-2">
            {t('integrations.telegramBotDescription')}
          </label>
          
          {connected ? (
            <div className="flex items-center gap-2 p-3 bg-green-500/10 border border-green-500/30 rounded-md">
              <Check size={16} className="text-green-500" />
              <span className="text-sm text-green-500">{t('integrations.telegramConnected')}</span>
            </div>
          ) : (
            <div className="space-y-3">
              <a
                href={telegramBotUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white text-sm font-medium rounded-md transition-colors"
              >
                {t('integrations.openTelegramBot')}
              </a>
              
              <div>
                <label className="block text-sm font-medium text-[var(--text)] mb-2">
                  {t('integrations.selectWorkflow')}
                </label>
                <select
                  value={workflowId}
                  onChange={(e) => setWorkflowId(e.target.value)}
                  className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm focus:outline-none focus:border-[var(--accent)] transition-colors"
                >
                  <option value="">{t('integrations.selectWorkflowPlaceholder')}</option>
                  {loadingWorkflows ? (
                    <option value="" disabled>{t('workflowLibrary.loading')}</option>
                  ) : (
                    workflows.map((wf) => (
                      <option key={wf.id} value={wf.id}>{wf.name}</option>
                    ))
                  )}
                </select>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => {
                    setUseDefaultKey(true);
                    handleSave();
                  }}
                  disabled={!workflowId}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
                >
                  {t('integrations.saveIntegration')}
                </button>
                <button
                  onClick={onCancel}
                  className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
                >
                  {t('integrations.cancelIntegration')}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }

  // N8n form
  if (type === 'n8n') {
    return (
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-[var(--text)] mb-2">
            {t('integrations.n8n.serverUrl')}
          </label>
          <input
            type="url"
            value={activityPubUrl}
            onChange={(e) => setActivityPubUrl(e.target.value)}
            placeholder={t('integrations.n8n.serverUrlPlaceholder')}
            className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
          />
          <p className="text-xs text-[var(--text-muted)] mt-1">
            Адрес вашего n8n сервера с MCP endpoint
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-[var(--text)] mb-2">
            {t('integrations.n8n.accessToken')}
          </label>
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={t('integrations.n8n.accessTokenPlaceholder')}
            className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text)] placeholder-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
          />
          <p className="text-xs text-[var(--text-muted)] mt-1">
            Токен доступа для аутентификации в n8n
          </p>
        </div>

        <div className="flex items-center gap-2 pt-2">
          <button
            onClick={handleSave}
            disabled={!activityPubUrl || !apiKey}
            className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
          >
            {t('integrations.saveIntegration')}
          </button>
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
          >
            {t('integrations.cancelIntegration')}
          </button>
        </div>
      </div>
    );
  }

  // Lefine.pro form
  return (
    <div className="space-y-4">
      {/* Use default key toggle */}
      <div className="flex items-center justify-between">
        <div>
          <div className="text-sm font-medium text-[var(--text)]">
            {t('integrations.useDefaultKey')}
          </div>
          <div className="text-xs text-[var(--text-muted)]">
            {t('integrations.useDefaultKeyHint')}
          </div>
        </div>
        <button
          onClick={() => setUseDefaultKey(!useDefaultKey)}
          className={`relative w-12 h-7 rounded-full transition-colors duration-200 ${
            useDefaultKey ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
          }`}
        >
          <span
            className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform duration-200 ${
              useDefaultKey ? 'translate-x-5' : 'translate-x-0'
            }`}
          />
        </button>
      </div>

      {/* API Key (if not using default) */}
      {!useDefaultKey && (
        <div>
          <label className="block text-sm font-medium text-[var(--text)] mb-2">
            {t('integrations.apiKey')}
          </label>
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
            placeholder={t('integrations.apiKeyPlaceholder')}
          />
        </div>
      )}

      {/* Workflow selector */}
      <div>
        <label className="block text-sm font-medium text-[var(--text)] mb-2">
          {t('integrations.selectWorkflow')}
        </label>
        <select
          value={workflowId}
          onChange={(e) => setWorkflowId(e.target.value)}
          className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm focus:outline-none focus:border-[var(--accent)] transition-colors"
        >
          <option value="">{t('integrations.selectWorkflowPlaceholder')}</option>
          {loadingWorkflows ? (
            <option value="" disabled>{t('workflowLibrary.loading')}</option>
          ) : (
            workflows.map((wf) => (
              <option key={wf.id} value={wf.id}>{wf.name}</option>
            ))
          )}
        </select>
      </div>

      {/* Lefine.pro ActivityPub endpoints */}
      {type === 'lefine' && (
        <>
          <div className="pt-2 border-t border-[var(--border)]">
            <div className="text-sm font-medium text-[var(--text)] mb-2">
              ActivityPub
            </div>
            <p className="text-xs text-[var(--text-muted)] mb-3">
              {t('integrations.activityPubDescription')}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('integrations.activityPubUrl')}
            </label>
            <input
              type="url"
              value={activityPubUrl}
              onChange={(e) => setActivityPubUrl(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('integrations.activityPubUrlPlaceholder')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('integrations.outboxEndpoint')}
            </label>
            <input
              type="url"
              value={outboxEndpoint}
              onChange={(e) => setOutboxEndpoint(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('integrations.outboxEndpointPlaceholder')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('integrations.inboxEndpoint')}
            </label>
            <input
              type="url"
              value={inboxEndpoint}
              onChange={(e) => setInboxEndpoint(e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('integrations.inboxEndpointPlaceholder')}
            />
          </div>
        </>
      )}

      {/* Actions */}
      <div className="flex items-center gap-2 pt-2">
        <button
          onClick={handleSave}
          disabled={type === 'n8n' ? !apiKey : (type !== 'github' && !workflowId)}
          className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
        >
          {t('integrations.saveIntegration')}
        </button>
        <button
          onClick={onCancel}
          className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
        >
          {t('integrations.cancelIntegration')}
        </button>
      </div>
    </div>
  );
}

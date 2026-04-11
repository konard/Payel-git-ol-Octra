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

  // Load workflows on mount
  useEffect(() => {
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
  }, []);

  const handleSave = () => {
    const config: IntegrationConfig = {
      useDefaultKey,
      apiKey: useDefaultKey ? defaultToken : apiKey,
      workflowId,
      ...(type === 'lefine' && {
        activityPubUrl,
        outboxEndpoint,
        inboxEndpoint,
      }),
    };
    onSave(config);
  };

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
          disabled={!workflowId}
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

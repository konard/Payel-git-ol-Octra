import { useState } from 'react';
import { Plug, Check, Settings2, ChevronDown, ChevronUp } from 'lucide-react';
import { useI18n } from '../hooks/useI18n';
import { IntegrationForm } from './IntegrationForm';
import type { IntegrationType, IntegrationConfig } from '../stores/integrationStore';

interface IntegrationCardProps {
  type: IntegrationType;
  name: string;
  description: string;
  icon: string;
  connected: boolean;
  config: IntegrationConfig;
  onConnect: (config: IntegrationConfig) => void;
  onDisconnect: () => void;
}

export function IntegrationCard({
  type,
  name,
  description,
  icon,
  connected,
  config,
  onConnect,
  onDisconnect,
}: IntegrationCardProps) {
  const { t } = useI18n();
  const [showForm, setShowForm] = useState(false);
  const [expanded, setExpanded] = useState(false);

  const handleIntegrate = () => {
    setShowForm(true);
    setExpanded(true);
  };

  const handleConfigure = () => {
    setExpanded(!expanded);
  };

  const handleSave = (newConfig: IntegrationConfig) => {
    onConnect(newConfig);
    setShowForm(false);
    setExpanded(false);
  };

  const handleCancel = () => {
    setShowForm(false);
    setExpanded(false);
  };

  const handleDisconnect = () => {
    onDisconnect();
    setShowForm(false);
    setExpanded(false);
  };

  return (
    <div className="bg-[var(--background)] border border-[var(--border)] rounded-lg overflow-hidden transition-all max-w-md">
      {/* Header */}
      <div className="flex items-center gap-3 p-3">
        {/* Icon */}
        <div className="w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 bg-[var(--surface)] flex items-center justify-center">
          <img
            src={icon}
            alt={name}
            className="w-full h-full object-cover"
          />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-[var(--text)] truncate">{name}</h3>
            {connected && (
              <span className="flex items-center gap-1 text-xs text-green-500">
                <Check size={12} />
                {t('integrations.connected')}
              </span>
            )}
          </div>
          <p className="text-xs text-[var(--text-muted)] mt-0.5 truncate">{description}</p>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1.5 flex-shrink-0">
          {connected ? (
            <>
              <button
                onClick={handleConfigure}
                className="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]"
                title={expanded ? t('integrations.collapse') : t('integrations.expand')}
              >
                {expanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
              </button>
              <button
                onClick={handleConfigure}
                className="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]"
                title={t('integrations.configure')}
              >
                <Settings2 size={14} />
              </button>
              <button
                onClick={handleDisconnect}
                className="px-2.5 py-1.5 text-xs font-medium bg-red-600 hover:bg-red-700 text-white rounded-md transition-colors"
              >
                {t('integrations.disconnect')}
              </button>
            </>
          ) : (
            <button
              onClick={handleIntegrate}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white text-xs font-medium rounded-md transition-colors"
            >
              <Plug size={12} />
              {t('integrations.integrate')}
            </button>
          )}
        </div>
      </div>

      {/* Expandable Form */}
      {expanded && (
        <div className="border-t border-[var(--border)] p-4 bg-[var(--surface)]">
          <IntegrationForm
            type={type}
            initialConfig={config}
            onSave={handleSave}
            onCancel={handleCancel}
            connected={connected}
          />
        </div>
      )}
    </div>
  );
}

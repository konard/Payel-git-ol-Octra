import { useState } from 'react';
import { Settings2, ChevronDown, ChevronUp, Trash2, Edit, Plus } from 'lucide-react';
import { useI18n } from '../hooks/useI18n';
import { customProviderService } from '../services/customProviderService';
import { CustomProvider } from '../stores/customProvidersStore';

interface CustomProviderCardProps {
  provider?: CustomProvider;
  isNew?: boolean;
  onSave: (provider: Omit<CustomProvider, 'id' | 'createdAt'>) => void;
  onCancel: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function CustomProviderCard({
  provider,
  isNew = false,
  onSave,
  onCancel,
  onEdit,
  onDelete,
}: CustomProviderCardProps) {
  const { t } = useI18n();
  const [isExpanded, setIsExpanded] = useState(isNew);
  const [formData, setFormData] = useState({
    name: provider?.name || '',
    baseUrl: provider?.base_url || '',
    apiKey: provider?.api_key || '',
  });

  const handleSave = async () => {
    if (!formData.name || !formData.baseUrl || !formData.apiKey) {
      return;
    }

    try {
      if (isNew) {
        const created = await customProviderService.createCustomProvider({
          name: formData.name,
          base_url: formData.baseUrl,
          api_key: formData.apiKey,
        });
        onSave(created);
      } else if (provider) {
        const updated = await customProviderService.updateCustomProvider(provider.id, {
          name: formData.name,
          base_url: formData.baseUrl,
          api_key: formData.apiKey,
        });
        onSave(updated);
      }
    } catch (error) {
      console.error('Failed to save custom provider:', error);
      // TODO: Show error message to user
    }
  };

  const handleDelete = async () => {
    if (!provider) return;

    try {
      await customProviderService.deleteCustomProvider(provider.id);
      onDelete?.();
    } catch (error) {
      console.error('Failed to delete custom provider:', error);
      // TODO: Show error message to user
    }
  };

  const handleInputChange = (field: string, value: string | number) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  return (
    <div className="bg-[var(--background)] border border-[var(--border)] rounded-lg overflow-hidden transition-all max-w-md">
      {/* Header */}
      <div className="flex items-center gap-3 p-3">
        {/* Icon */}
        <div className="w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 bg-[var(--accent)] flex items-center justify-center">
          <Plus size={20} className="text-white" />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-[var(--text)] truncate">
              {isNew ? t('providers.newProvider') : provider?.name}
            </h3>
            {!isNew && (
              <span className="flex items-center gap-1 text-xs text-green-500">
                ✓ {t('providers.configured')}
              </span>
            )}
          </div>
          <p className="text-xs text-[var(--text-muted)] mt-0.5 truncate">
            {isNew ? t('providers.addCustomProvider') : provider?.base_url}
          </p>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1.5 flex-shrink-0">
          {!isNew ? (
            <>
              <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]"
                title={isExpanded ? t('providers.collapse') : t('providers.expand')}
              >
                {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
              </button>
              <button
                onClick={onEdit}
                className="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]"
                title={t('providers.edit')}
              >
                <Edit size={14} />
              </button>
              <button
                onClick={onDelete}
                className="p-1.5 hover:bg-red-500/20 rounded-md transition-colors text-red-500"
                title={t('providers.delete')}
              >
                <Trash2 size={14} />
              </button>
            </>
          ) : (
            <button
              onClick={() => setIsExpanded(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white text-xs font-medium rounded-md transition-colors"
            >
              <Plus size={12} />
              {t('providers.add')}
            </button>
          )}
        </div>
      </div>

      {/* Expandable Form */}
      {isExpanded && (
        <div className="border-t border-[var(--border)] p-4 bg-[var(--surface)] space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('providers.name')}
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('providers.namePlaceholder')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('providers.baseUrl')}
            </label>
            <input
              type="url"
              value={formData.baseUrl}
              onChange={(e) => handleInputChange('baseUrl', e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder="https://api.example.com/v1"
            />
            <p className="text-xs text-[var(--text-muted)] mt-1">
              {t('providers.baseUrlHint')}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--text)] mb-2">
              {t('providers.apiKey')}
            </label>
            <input
              type="password"
              value={formData.apiKey}
              onChange={(e) => handleInputChange('apiKey', e.target.value)}
              className="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              placeholder={t('providers.apiKeyPlaceholder')}
            />
          </div>



          {/* Actions */}
          <div className="flex items-center gap-2 pt-2">
            <button
              onClick={handleSave}
              disabled={!formData.name || !formData.baseUrl || !formData.apiKey || !formData.model}
              className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors"
            >
              {t('providers.save')}
            </button>
            <button
              onClick={onCancel}
              className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors"
            >
              {t('providers.cancel')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
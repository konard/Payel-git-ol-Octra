'use client';

import {
  Box,
  Braces,
  Cpu,
  Globe2,
  Layers3,
  Plus,
  Search,
  Server,
  X,
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { searchCatalog, type CatalogCategory, type CatalogItem } from '../server/catalog';
import type { DashboardEnvironment } from '../server/environments';

type CatalogSearchModalProps = {
  open: boolean;
  environments: DashboardEnvironment[];
  onClose: () => void;
  onSelect: (item: CatalogItem) => void;
};

const categories: Array<{ id: CatalogCategory; label: string }> = [
  { id: 'all', label: 'All' },
  { id: 'providers', label: 'Providers' },
  { id: 'cli', label: 'CLI' },
  { id: 'skills', label: 'Skills' },
  { id: 'custom', label: 'Custom' },
];

const iconByType = {
  provider: Globe2,
  cli: Cpu,
  skill: Box,
  custom_provider: Braces,
  environment: Layers3,
} as const;

export function CatalogSearchModal({ open, environments, onClose, onSelect }: CatalogSearchModalProps) {
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<CatalogCategory>('all');
  const [items, setItems] = useState<CatalogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [customName, setCustomName] = useState('Custom provider');
  const [customBaseURL, setCustomBaseURL] = useState('');
  const [customModel, setCustomModel] = useState('');
  const [customAPIKey, setCustomAPIKey] = useState('');

  useEffect(() => {
    if (!open) return;
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [open, onClose]);

  useEffect(() => {
    if (!open) return;
    const controller = new AbortController();
    const handle = window.setTimeout(() => {
      setLoading(true);
      setError('');
      searchCatalog(query, category, controller.signal)
        .then((result) => setItems(result.items))
        .catch((err: Error) => {
          if (controller.signal.aborted) return;
          setItems([]);
          setError(err.message || 'Search failed');
        })
        .finally(() => {
          if (!controller.signal.aborted) setLoading(false);
        });
    }, 140);

    return () => {
      controller.abort();
      window.clearTimeout(handle);
    };
  }, [open, query, category]);

  const environmentItems = useMemo<CatalogItem[]>(() => {
    if (category !== 'all') return [];
    const needle = query.trim().toLowerCase();
    return environments
      .filter((env) => !needle || env.name.toLowerCase().includes(needle) || env.id.toLowerCase().includes(needle))
      .slice(0, 6)
      .map((env) => ({
        id: env.id,
        type: 'environment',
        name: env.name,
        subtitle: env.visibility,
        description: env.id,
      }));
  }, [category, environments, query]);

  const visibleItems = [...environmentItems, ...items];
  const showCustomForm = category === 'custom' || category === 'providers' || category === 'all';

  function addCustomProvider() {
    onSelect({
      id: `custom-${Date.now()}`,
      type: 'custom_provider',
      name: customName.trim() || 'Custom provider',
      subtitle: customBaseURL.trim() || 'OpenAI-compatible endpoint',
      description: customAPIKey ? 'API key set' : 'No API key',
      base_url: customBaseURL.trim(),
      default_model: customModel.trim(),
      auth_env: 'CUSTOM_API_KEY',
      api_key: customAPIKey,
    });
  }

  if (!open) return null;

  return (
    <div className="catalog-search-overlay" role="dialog" aria-modal="true" aria-label="Search catalogue">
      <section className="catalog-search-dialog">
        <div className="catalog-search-bar">
          <Search size={22} />
          <input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search providers, CLI, skills, environments..."
          />
          {query ? (
            <button type="button" className="catalog-icon-button" onClick={() => setQuery('')} aria-label="Clear search">
              <X size={18} />
            </button>
          ) : null}
          <button type="button" className="catalog-icon-button" onClick={onClose} aria-label="Close search">
            <X size={18} />
          </button>
        </div>

        <div className="catalog-tabs" role="tablist" aria-label="Catalogue categories">
          {categories.map((item) => (
            <button
              key={item.id}
              type="button"
              className={`catalog-tab${category === item.id ? ' active' : ''}`}
              onClick={() => setCategory(item.id)}
              role="tab"
              aria-selected={category === item.id}
            >
              {item.label}
            </button>
          ))}
        </div>

        {showCustomForm ? (
          <section className="catalog-custom-provider" aria-label="Custom provider">
            <div className="catalog-custom-grid">
              <input value={customName} onChange={(event) => setCustomName(event.target.value)} placeholder="Provider name" />
              <input value={customBaseURL} onChange={(event) => setCustomBaseURL(event.target.value)} placeholder="Base URL" />
              <input value={customModel} onChange={(event) => setCustomModel(event.target.value)} placeholder="Model" />
              <input
                value={customAPIKey}
                onChange={(event) => setCustomAPIKey(event.target.value)}
                placeholder="API key"
                type="password"
              />
            </div>
            <button type="button" className="catalog-add-custom" onClick={addCustomProvider}>
              <Plus size={16} />
              Add custom
            </button>
          </section>
        ) : null}

        <div className="catalog-results" aria-live="polite">
          {loading ? <p className="catalog-state">Searching...</p> : null}
          {!loading && error ? <p className="catalog-state">{error}</p> : null}
          {!loading && !error && visibleItems.length === 0 ? <p className="catalog-state">No results</p> : null}
          {!loading && !error ? visibleItems.map((item) => <CatalogResult key={`${item.type}-${item.id}`} item={item} onSelect={onSelect} />) : null}
        </div>
      </section>
    </div>
  );
}

function CatalogResult({ item, onSelect }: { item: CatalogItem; onSelect: (item: CatalogItem) => void }) {
  const Icon = iconByType[item.type] ?? Server;
  return (
    <button type="button" className={`catalog-result type-${item.type}`} onClick={() => onSelect(item)}>
      <span className="catalog-result-icon">
        <Icon size={18} />
      </span>
      <span className="catalog-result-copy">
        <strong>{item.name}</strong>
        <span>{item.subtitle || item.description || item.type}</span>
      </span>
      <Plus size={18} />
    </button>
  );
}

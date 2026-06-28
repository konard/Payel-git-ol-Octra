'use client';

import {
  Box,
  Braces,
  Check,
  Cpu,
  Globe2,
  Plus,
  Search,
  Server,
  X,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { searchCatalog, type CatalogCategory, type CatalogItem } from '../server/catalog';

type CatalogSearchModalProps = {
  open: boolean;
  onClose: () => void;
  onSelect: (item: CatalogItem) => void;
  addedKeys?: Set<string>;
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
} as const;

export function CatalogSearchModal({ open, onClose, onSelect, addedKeys }: CatalogSearchModalProps) {
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<CatalogCategory>('all');
  const [items, setItems] = useState<CatalogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

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

  const visibleItems = items;

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

        <div className="catalog-results" aria-live="polite">
          {loading ? <p className="catalog-state">Searching...</p> : null}
          {!loading && error ? <p className="catalog-state">{error}</p> : null}
          {!loading && !error && visibleItems.length === 0 ? <p className="catalog-state">No results</p> : null}
          {!loading && !error ? visibleItems.map((item) => <CatalogResult key={`${item.type}-${item.id}`} item={item} onSelect={onSelect} addedKeys={addedKeys} />) : null}
        </div>
      </section>
    </div>
  );
}

function CatalogResult({ item, onSelect, addedKeys }: { item: CatalogItem; onSelect: (item: CatalogItem) => void; addedKeys?: Set<string> }) {
  const itemKey = `${item.type}-${item.id}`;
  const [added, setAdded] = useState(addedKeys?.has(itemKey) ?? false);
  const Icon = iconByType[item.type] ?? Server;

  const handleAdd = useCallback(() => {
    setAdded(true);
    onSelect(item);
  }, [item, onSelect]);

  return (
    <section className={`catalog-result type-${item.type}${added ? ' added' : ''}`}>
      <span className="catalog-result-icon">
        <Icon size={18} />
      </span>
      <span className="catalog-result-copy">
        <strong>{item.name}</strong>
        <span>{item.subtitle || item.description || item.type}</span>
      </span>
      <button type="button" className="catalog-result-add" onClick={handleAdd} aria-label={`Add ${item.name}`}>
        <span className="catalog-result-add-icon">
          <Plus size={18} />
          <Check size={18} />
        </span>
      </button>
    </section>
  );
}

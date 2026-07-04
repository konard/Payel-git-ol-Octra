import { API } from '../config/routes';

export type CatalogCategory = 'all' | 'providers' | 'cli' | 'skills' | 'custom' | 'mcp' | 'adapters';

export type CatalogItem = {
  id: string;
  type: 'provider' | 'cli' | 'skill' | 'custom_provider' | 'mcp_server' | 'adapter';
  name: string;
  subtitle?: string;
  description?: string;
  key?: string;
  base_url?: string;
  auth_env?: string;
  default_model?: string;
  nix_attr?: string;
  install_cmd?: string;
  skill_id?: string;
  source?: string;
  api_key?: string;
};

export type CatalogSearchResponse = {
  query: string;
  category: CatalogCategory;
  items: CatalogItem[];
  count: number;
};

export async function searchCatalog(
  query: string,
  category: CatalogCategory,
  offset: number = 0,
  signal?: AbortSignal,
): Promise<CatalogSearchResponse> {
  const params = new URLSearchParams({
    q: query,
    category,
    limit: '30',
    offset: String(offset),
  });
  const res = await fetch(`${API.CATALOG_SEARCH}?${params.toString()}`, {
    cache: 'no-cache',
    signal,
  });
  if (!res.ok) {
    throw new Error(await res.text());
  }
  return res.json();
}

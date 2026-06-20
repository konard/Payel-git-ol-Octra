import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import {
  DEFAULT_HIDE_API_KEY_INPUT,
  DEFAULT_MODEL,
  DEFAULT_PROVIDER,
  DEFAULT_SEARCH_PROVIDER_CONFIG,
  DEFAULT_SEARCH_PROVIDER_ID,
  DEFAULT_TOKEN,
} from '../config/defaultSettings';

export type SearchProviderKind = 'apodex' | 'custom';

export interface SearchProviderConfig {
  id: string;
  provider: SearchProviderKind;
  name: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  streaming: boolean;
}

interface SettingsState {
  defaultToken: string;
  hideApiKeyInput: boolean;
  hideServerStatus: boolean;
  hideConsole: boolean;
  defaultProvider: string;
  defaultModel: string;
  searchProviderId: string;
  searchProviders: SearchProviderConfig[];

  setDefaultToken: (token: string) => void;
  setHideApiKeyInput: (hide: boolean) => void;
  setHideServerStatus: (hide: boolean) => void;
  setHideConsole: (hide: boolean) => void;
  setDefaultProvider: (provider: string) => void;
  setDefaultModel: (model: string) => void;
  setSearchProviderId: (providerId: string) => void;
  updateSearchProvider: (providerId: string, updates: Partial<SearchProviderConfig>) => void;
  addSearchProvider: (provider?: Partial<SearchProviderConfig>) => string;
  deleteSearchProvider: (providerId: string) => void;
}

function createSearchProviderId(): string {
  return globalThis.crypto?.randomUUID?.() || `search-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function defaultSearchProvider(): SearchProviderConfig {
  return { ...DEFAULT_SEARCH_PROVIDER_CONFIG };
}

function normalizeSearchProviders(providers?: Partial<SearchProviderConfig>[]): SearchProviderConfig[] {
  const normalized = (providers || [])
    .map((provider) => ({
      id: provider.id?.trim() || createSearchProviderId(),
      provider: provider.provider === 'custom' ? 'custom' : 'apodex',
      name: provider.name?.trim() || (provider.provider === 'custom' ? 'Custom Search' : 'Apodex'),
      baseUrl: provider.baseUrl?.trim() || '',
      apiKey: provider.apiKey || '',
      model: provider.model?.trim() || '',
      streaming: provider.streaming ?? true,
    }))
    .filter((provider) => provider.id && provider.name);

  const hasApodex = normalized.some((provider) => provider.id === DEFAULT_SEARCH_PROVIDER_ID);
  if (!hasApodex) {
    normalized.unshift(defaultSearchProvider());
  }

  return normalized.map((provider) => {
    if (provider.id !== DEFAULT_SEARCH_PROVIDER_ID) {
      return provider;
    }

    const fallback = defaultSearchProvider();
    return {
      ...fallback,
      ...provider,
      provider: 'apodex',
      name: provider.name || fallback.name,
      baseUrl: provider.baseUrl || fallback.baseUrl,
      model: provider.model || fallback.model,
      streaming: provider.streaming ?? fallback.streaming,
    };
  });
}

function withSearchDefaults(state: Partial<SettingsState> | undefined): Partial<SettingsState> {
  const searchProviders = normalizeSearchProviders(state?.searchProviders);
  const activeProvider = searchProviders.some((provider) => provider.id === state?.searchProviderId)
    ? state?.searchProviderId
    : DEFAULT_SEARCH_PROVIDER_ID;

  return {
    ...state,
    searchProviderId: activeProvider,
    searchProviders,
  };
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      defaultToken: DEFAULT_TOKEN,
      hideApiKeyInput: DEFAULT_HIDE_API_KEY_INPUT,
      hideServerStatus: false,
      hideConsole: false,
      defaultProvider: DEFAULT_PROVIDER,
      defaultModel: DEFAULT_MODEL,
      searchProviderId: DEFAULT_SEARCH_PROVIDER_ID,
      searchProviders: [defaultSearchProvider()],

      setDefaultToken: (token) => set({ defaultToken: token }),
      setHideApiKeyInput: (hide) => set({ hideApiKeyInput: hide }),
      setHideServerStatus: (hide) => set({ hideServerStatus: hide }),
      setHideConsole: (hide) => set({ hideConsole: hide }),
      setDefaultProvider: (provider) => set({ defaultProvider: provider }),
      setDefaultModel: (model) => set({ defaultModel: model }),
      setSearchProviderId: (providerId) => set((state) => ({
        searchProviderId: state.searchProviders.some((provider) => provider.id === providerId)
          ? providerId
          : DEFAULT_SEARCH_PROVIDER_ID,
      })),
      updateSearchProvider: (providerId, updates) => set((state) => ({
        searchProviders: normalizeSearchProviders(state.searchProviders.map((provider) =>
          provider.id === providerId ? { ...provider, ...updates, id: provider.id } : provider
        )),
      })),
      addSearchProvider: (provider) => {
        const id = provider?.id?.trim() || createSearchProviderId();
        set((state) => ({
          searchProviderId: id,
          searchProviders: normalizeSearchProviders([
            ...state.searchProviders,
            {
              id,
              provider: 'custom',
              name: 'Custom Search',
              baseUrl: '',
              apiKey: '',
              model: '',
              streaming: false,
              ...provider,
            },
          ]),
        }));
        return id;
      },
      deleteSearchProvider: (providerId) => set((state) => {
        if (providerId === DEFAULT_SEARCH_PROVIDER_ID) {
          return state;
        }

        const searchProviders = normalizeSearchProviders(
          state.searchProviders.filter((provider) => provider.id !== providerId)
        );
        return {
          searchProviders,
          searchProviderId: state.searchProviderId === providerId ? DEFAULT_SEARCH_PROVIDER_ID : state.searchProviderId,
        };
      }),
    }),
    {
      name: 'crewai-settings',
      version: 2,
      migrate: (persistedState, version) => {
        const state = persistedState as Partial<SettingsState> | undefined;

        if (version < 1) {
          return withSearchDefaults({
            ...state,
            defaultToken: state?.defaultToken?.trim() || DEFAULT_TOKEN,
            hideApiKeyInput: DEFAULT_HIDE_API_KEY_INPUT,
            defaultProvider: state?.defaultProvider || DEFAULT_PROVIDER,
            defaultModel: state?.defaultModel || DEFAULT_MODEL,
          });
        }

        if (version < 2) {
          return withSearchDefaults(state);
        }

        return withSearchDefaults(state);
      },
    }
  )
);

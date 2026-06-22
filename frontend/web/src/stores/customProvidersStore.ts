import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import {
  normalizeCustomModel,
  normalizeCustomModelList,
  normalizeCustomProvider,
  normalizeCustomProviderList,
  normalizeCustomProvidersState,
  type CustomModelInput,
  type CustomProviderInput,
  upsertCustomModel,
  upsertCustomProvider,
} from '../utils/customProviders';
import { scopedStorage } from './storageScope';

export interface CustomProvider {
  id: string;
  user_id: string;
  name: string;
  base_url: string;
  api_key: string;
  requires_api_key: boolean;
  created_at: string;
  updated_at: string;
}

export interface CustomModel {
  id: string;
  user_id: string;
  name: string;
  provider_id?: string;
  created_at: string;
  updated_at: string;
}

interface CustomProvidersState {
  providers: CustomProvider[];
  models: CustomModel[];

  // Actions
  addProvider: (provider: CustomProviderInput) => void;
  updateProvider: (id: string, updates: CustomProviderInput) => void;
  deleteProvider: (id: string) => void;
  getProvider: (id: string) => CustomProvider | undefined;
  setProviders: (providers: CustomProviderInput[]) => void;

  addModel: (model: CustomModelInput) => void;
  updateModel: (id: string, updates: CustomModelInput) => void;
  deleteModel: (id: string) => void;
  getModel: (id: string) => CustomModel | undefined;
  setModels: (models: CustomModelInput[]) => void;
}

function createLocalId(): string {
  return globalThis.crypto?.randomUUID?.() || `local-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export const useCustomProvidersStore = create<CustomProvidersState>()(
  persist(
    (set, get) => ({
      providers: [],
      models: [],

      addProvider: (providerData) => {
        const newProvider = normalizeCustomProvider(providerData, { fallbackId: createLocalId() });
        if (!newProvider) {
          return;
        }

        set((state) => ({
          providers: upsertCustomProvider(state.providers, newProvider),
        }));
      },

      updateProvider: (id, updates) => {
        set((state) => ({
          providers: normalizeCustomProviderList(state.providers).map((provider) =>
            provider.id === id
              ? normalizeCustomProvider({ ...provider, ...updates }, { fallbackId: provider.id }) || provider
              : provider
          ),
        }));
      },

      deleteProvider: (id) => {
        set((state) => ({
          providers: normalizeCustomProviderList(state.providers).filter((p) => p.id !== id),
        }));
      },

      getProvider: (id) => {
        return normalizeCustomProviderList(get().providers).find((p) => p.id === id);
      },

      addModel: (modelData) => {
        const newModel = normalizeCustomModel(modelData, { fallbackId: createLocalId() });
        if (!newModel) {
          return;
        }

        set((state) => ({
          models: upsertCustomModel(state.models, newModel),
        }));
      },

      updateModel: (id, updates) => {
        set((state) => ({
          models: normalizeCustomModelList(state.models).map((model) =>
            model.id === id
              ? normalizeCustomModel({ ...model, ...updates }, { fallbackId: model.id }) || model
              : model
          ),
        }));
      },

      deleteModel: (id) => {
        set((state) => ({
          models: normalizeCustomModelList(state.models).filter((m) => m.id !== id),
        }));
      },

      getModel: (id) => {
        return normalizeCustomModelList(get().models).find((m) => m.id === id);
      },

      setProviders: (providers) => {
        set({ providers: normalizeCustomProviderList(providers) });
      },

      setModels: (models) => {
        set({ models: normalizeCustomModelList(models) });
      },
    }),
    {
      name: 'crewai-custom-providers',
      storage: createJSONStorage(() => scopedStorage),
      version: 1,
      migrate: (persistedState) =>
        normalizeCustomProvidersState(persistedState) as CustomProvidersState,
      merge: (persistedState, currentState) => ({
        ...currentState,
        ...normalizeCustomProvidersState(persistedState),
      }),
    }
  )
);

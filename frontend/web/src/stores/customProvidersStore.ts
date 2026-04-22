import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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
  addProvider: (provider: Omit<CustomProvider, 'id' | 'createdAt'>) => void;
  updateProvider: (id: string, updates: Partial<CustomProvider>) => void;
  deleteProvider: (id: string) => void;
  getProvider: (id: string) => CustomProvider | undefined;

  addModel: (model: Omit<CustomModel, 'id' | 'createdAt'>) => void;
  updateModel: (id: string, updates: Partial<CustomModel>) => void;
  deleteModel: (id: string) => void;
  getModel: (id: string) => CustomModel | undefined;
}

export const useCustomProvidersStore = create<CustomProvidersState>()(
  persist(
    (set, get) => ({
      providers: [],
      models: [],

      addProvider: (providerData) => {
        const newProvider: CustomProvider = {
          ...providerData,
          id: crypto.randomUUID(),
          createdAt: new Date().toISOString(),
        };
        set((state) => ({
          providers: [...state.providers, newProvider],
        }));
      },

      updateProvider: (id, updates) => {
        set((state) => ({
          providers: state.providers.map((p) =>
            p.id === id ? { ...p, ...updates } : p
          ),
        }));
      },

      deleteProvider: (id) => {
        set((state) => ({
          providers: state.providers.filter((p) => p.id !== id),
        }));
      },

      getProvider: (id) => {
        return get().providers.find((p) => p.id === id);
      },

      addModel: (modelData) => {
        const newModel: CustomModel = {
          ...modelData,
          id: crypto.randomUUID(),
          createdAt: new Date().toISOString(),
        };
        set((state) => ({
          models: [...state.models, newModel],
        }));
      },

      updateModel: (id, updates) => {
        set((state) => ({
          models: state.models.map((m) =>
            m.id === id ? { ...m, ...updates } : m
          ),
        }));
      },

      deleteModel: (id) => {
        set((state) => ({
          models: state.models.filter((m) => m.id !== id),
        }));
      },

      getModel: (id) => {
        return get().models.find((m) => m.id === id);
      },
    }),
    {
      name: 'crewai-custom-providers',
    }
  )
);
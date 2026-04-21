import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CustomProvider {
  id: string;
  user_id: string;
  name: string;
  base_url: string;
  api_key: string;
  created_at: string;
  updated_at: string;
}

interface CustomProvidersState {
  providers: CustomProvider[];

  // Actions
  addProvider: (provider: Omit<CustomProvider, 'id' | 'createdAt'>) => void;
  updateProvider: (id: string, updates: Partial<CustomProvider>) => void;
  deleteProvider: (id: string) => void;
  getProvider: (id: string) => CustomProvider | undefined;
}

export const useCustomProvidersStore = create<CustomProvidersState>()(
  persist(
    (set, get) => ({
      providers: [],

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
    }),
    {
      name: 'crewai-custom-providers',
    }
  )
);
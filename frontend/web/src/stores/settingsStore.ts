import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SettingsState {
  defaultToken: string;
  hideApiKeyInput: boolean;
  hideServerStatus: boolean;
  hideConsole: boolean;
  defaultProvider: string;
  defaultModel: string;

  setDefaultToken: (token: string) => void;
  setHideApiKeyInput: (hide: boolean) => void;
  setHideServerStatus: (hide: boolean) => void;
  setHideConsole: (hide: boolean) => void;
  setDefaultProvider: (provider: string) => void;
  setDefaultModel: (model: string) => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      defaultToken: '',
      hideApiKeyInput: false,
      hideServerStatus: false,
      hideConsole: false,
      defaultProvider: 'openrouter',
      defaultModel: 'qwen/qwen3-coder',

      setDefaultToken: (token) => set({ defaultToken: token }),
      setHideApiKeyInput: (hide) => set({ hideApiKeyInput: hide }),
      setHideServerStatus: (hide) => set({ hideServerStatus: hide }),
      setHideConsole: (hide) => set({ hideConsole: hide }),
      setDefaultProvider: (provider) => set({ defaultProvider: provider }),
      setDefaultModel: (model) => set({ defaultModel: model }),
    }),
    {
      name: 'crewai-settings',
    }
  )
);

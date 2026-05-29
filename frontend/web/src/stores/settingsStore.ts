import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import {
  DEFAULT_HIDE_API_KEY_INPUT,
  DEFAULT_MODEL,
  DEFAULT_PROVIDER,
  DEFAULT_TOKEN,
} from '../config/defaultSettings';

interface SettingsState {
  defaultToken: string;
  hideApiKeyInput: boolean;
  hideServerStatus: boolean;
  hideConsole: boolean;
  defaultProvider: string;
  defaultModel: string;
  ambientLighting: boolean;

  setDefaultToken: (token: string) => void;
  setHideApiKeyInput: (hide: boolean) => void;
  setHideServerStatus: (hide: boolean) => void;
  setHideConsole: (hide: boolean) => void;
  setDefaultProvider: (provider: string) => void;
  setDefaultModel: (model: string) => void;
  setAmbientLighting: (enabled: boolean) => void;
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
      ambientLighting: true,

      setDefaultToken: (token) => set({ defaultToken: token }),
      setHideApiKeyInput: (hide) => set({ hideApiKeyInput: hide }),
      setHideServerStatus: (hide) => set({ hideServerStatus: hide }),
      setHideConsole: (hide) => set({ hideConsole: hide }),
      setDefaultProvider: (provider) => set({ defaultProvider: provider }),
      setDefaultModel: (model) => set({ defaultModel: model }),
      setAmbientLighting: (enabled) => set({ ambientLighting: enabled }),
    }),
    {
      name: 'crewai-settings',
      version: 1,
      migrate: (persistedState, version) => {
        const state = persistedState as Partial<SettingsState> | undefined;

        if (version < 1) {
          return {
            ...state,
            defaultToken: state?.defaultToken?.trim() || DEFAULT_TOKEN,
            hideApiKeyInput: DEFAULT_HIDE_API_KEY_INPUT,
            defaultProvider: state?.defaultProvider || DEFAULT_PROVIDER,
            defaultModel: state?.defaultModel || DEFAULT_MODEL,
          };
        }

        return state || {};
      },
    }
  )
);

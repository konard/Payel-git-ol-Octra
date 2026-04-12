import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SettingsState {
  defaultToken: string;
  hideApiKeyInput: boolean;
  hideServerStatus: boolean;
  hideConsole: boolean;

  setDefaultToken: (token: string) => void;
  setHideApiKeyInput: (hide: boolean) => void;
  setHideServerStatus: (hide: boolean) => void;
  setHideConsole: (hide: boolean) => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      defaultToken: '',
      hideApiKeyInput: false,
      hideServerStatus: false,
      hideConsole: false,

      setDefaultToken: (token) => set({ defaultToken: token }),
      setHideApiKeyInput: (hide) => set({ hideApiKeyInput: hide }),
      setHideServerStatus: (hide) => set({ hideServerStatus: hide }),
      setHideConsole: (hide) => set({ hideConsole: hide }),
    }),
    {
      name: 'crewai-settings',
    }
  )
);

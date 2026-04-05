import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SettingsState {
  defaultToken: string;
  hideApiKeyInput: boolean;

  setDefaultToken: (token: string) => void;
  setHideApiKeyInput: (hide: boolean) => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      defaultToken: '',
      hideApiKeyInput: false,

      setDefaultToken: (token) => set({ defaultToken: token }),
      setHideApiKeyInput: (hide) => set({ hideApiKeyInput: hide }),
    }),
    {
      name: 'crewai-settings',
    }
  )
);

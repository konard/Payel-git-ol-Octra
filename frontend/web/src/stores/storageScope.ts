import { useAuthStore } from './authStore';
import type { StateStorage } from 'zustand/middleware';

function getBrowserStorage(): Storage | null {
  return typeof localStorage === 'undefined' ? null : localStorage;
}

function getScopedKey(base: string): string {
  const userId = useAuthStore.getState().user?.id;
  if (!userId) return base;
  return `${base}_${userId}`;
}

export const scopedStorage: StateStorage = {
  getItem: (name: string): string | null => {
    const key = getScopedKey(name);
    return getBrowserStorage()?.getItem(key) ?? null;
  },
  setItem: (name: string, value: string): void => {
    const key = getScopedKey(name);
    getBrowserStorage()?.setItem(key, value);
  },
  removeItem: (name: string): void => {
    const key = getScopedKey(name);
    getBrowserStorage()?.removeItem(key);
  },
};

const STORE_PREFIXES = [
  'crewai-custom-providers',
  'crewai-settings',
  'crewai-integrations',
  'octra-token-statistics',
];

export function clearUserScopedData(): void {
  const userId = useAuthStore.getState().user?.id;
  if (!userId) return;
  const storage = getBrowserStorage();
  if (!storage) return;
  for (const prefix of STORE_PREFIXES) {
    const key = `${prefix}_${userId}`;
    storage.removeItem(key);
  }
}

import { useAuthStore } from './authStore';
import type { StateStorage } from 'zustand/middleware';

const hasLocalStorage = typeof localStorage !== 'undefined';

function getScopedKey(base: string): string {
  const userId = useAuthStore.getState().user?.id;
  if (!userId) return base;
  return `${base}_${userId}`;
}

export const scopedStorage: StateStorage = {
  getItem: (name: string): string | null => {
    if (!hasLocalStorage) return null;
    const key = getScopedKey(name);
    return localStorage.getItem(key);
  },
  setItem: (name: string, value: string): void => {
    if (!hasLocalStorage) return;
    const key = getScopedKey(name);
    localStorage.setItem(key, value);
  },
  removeItem: (name: string): void => {
    if (!hasLocalStorage) return;
    const key = getScopedKey(name);
    localStorage.removeItem(key);
  },
};

const STORE_PREFIXES = [
  'crewai-custom-providers',
  'crewai-settings',
  'crewai-integrations',
  'octra-token-statistics',
];

export function clearUserScopedData(): void {
  if (!hasLocalStorage) return;
  const userId = useAuthStore.getState().user?.id;
  if (!userId) return;
  for (const prefix of STORE_PREFIXES) {
    const key = `${prefix}_${userId}`;
    localStorage.removeItem(key);
  }
}

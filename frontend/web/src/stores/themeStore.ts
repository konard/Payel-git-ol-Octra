/**
 * Theme Store
 * Zustand store для управления темой
 */

import { create } from 'zustand';

const THEME_KEY = 'theme';

function getStoredTheme(): boolean {
  const stored = localStorage.getItem(THEME_KEY);
  if (stored !== null) {
    return stored === 'dark';
  }
  // По умолчанию темная тема
  return true;
}

function storeTheme(isDark: boolean): void {
  localStorage.setItem(THEME_KEY, isDark ? 'dark' : 'light');
}

interface ThemeState {
  isDark: boolean;
  toggleTheme: () => void;
  setTheme: (isDark: boolean) => void;
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  isDark: getStoredTheme(),
  toggleTheme: () => {
    set((state) => {
      const newTheme = !state.isDark;
      storeTheme(newTheme);
      return { isDark: newTheme };
    });
  },
  setTheme: (isDark: boolean) => {
    storeTheme(isDark);
    set({ isDark });
  },
}));

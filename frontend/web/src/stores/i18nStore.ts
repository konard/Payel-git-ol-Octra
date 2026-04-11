import { create } from 'zustand';
import { SUPPORTED_LANGUAGES, type LanguageCode } from '../config/languages';

type TranslationValue = string | Record<string, any>;
type TranslationDict = Record<string, TranslationValue>;

interface I18nState {
  language: LanguageCode;
  translations: Record<string, TranslationDict>;
  loaded: boolean;

  // Actions
  setLanguage: (lang: LanguageCode) => void;
  loadTranslations: (lang: LanguageCode, data: TranslationDict) => void;
}

function getNestedValue(obj: TranslationDict, path: string): string {
  const keys = path.split('.');
  let current: any = obj;
  for (const key of keys) {
    if (current && typeof current === 'object' && key in current) {
      current = current[key];
    } else {
      return path;
    }
  }
  return typeof current === 'string' ? current : path;
}

export const useI18nStore = create<I18nState>((set, get) => {
  // Detect initial language
  const saved = localStorage.getItem('crewai-language');
  let initialLang: LanguageCode = 'en';
  if (saved && SUPPORTED_LANGUAGES.includes(saved as LanguageCode)) {
    initialLang = saved as LanguageCode;
  } else {
    const browserLang = navigator.language.split('-')[0];
    if (SUPPORTED_LANGUAGES.includes(browserLang as LanguageCode)) {
      initialLang = browserLang as LanguageCode;
    }
  }

  return {
    language: initialLang,
    translations: {},
    loaded: false,

    setLanguage: async (lang) => {
      const { translations, loadTranslations } = get();

      // Already loaded
      if (translations[lang]) {
        set({ language: lang });
        localStorage.setItem('crewai-language', lang);
        document.documentElement.lang = lang;
        return;
      }

      // Load from file
      try {
        const response = await fetch(`/languages/${lang}.json`);
        if (!response.ok) throw new Error(`Failed to load ${lang}`);
        const data = await response.json();
        loadTranslations(lang, data);
      } catch (error) {
        console.error(`Failed to load language ${lang}:`, error);
      }
    },

    loadTranslations: (lang, data) => {
      set((state) => ({
        language: lang,
        translations: { ...state.translations, [lang]: data },
        loaded: true,
      }));
      localStorage.setItem('crewai-language', lang);
      document.documentElement.lang = lang;
    },
  };
});

// Translate helper - reads from store lazily
function getCurrentStoreState() {
  return useI18nStore.getState();
}

export function t(key: string): string {
  const state = getCurrentStoreState();
  const dict = state.translations[state.language];
  if (!dict) return key;

  const keys = key.split('.');
  let current: any = dict;
  for (const k of keys) {
    if (current && typeof current === 'object' && k in current) {
      current = current[k];
    } else {
      return key;
    }
  }
  return typeof current === 'string' ? current : key;
}

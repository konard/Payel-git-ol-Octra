import { create } from 'zustand';
import { SUPPORTED_LANGUAGES, COUNTRY_TO_LANGUAGE, type LanguageCode } from '../config/languages';

type TranslationValue = string | Record<string, any>;
type TranslationDict = Record<string, TranslationValue>;

interface I18nState {
  language: LanguageCode;

  // Actions
  setLanguage: (lang: LanguageCode) => void;
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

    setLanguage: (lang) => {
      localStorage.setItem('crewai-language', lang);
      document.documentElement.lang = lang;
      set({ language: lang });
    },
  };
});

// Translate helper removed - use useI18n hook instead

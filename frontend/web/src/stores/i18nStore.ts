import { create } from 'zustand';
import { SUPPORTED_LANGUAGES, COUNTRY_TO_LANGUAGE, type LanguageCode } from '../config/languages';
import { useAuthStore } from './authStore';

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

export const useI18nStore = create<I18nState>((set) => {
  const saved = localStorage.getItem('octra-language');

  // Default
  let initialLang: LanguageCode = 'en';

  const isAuthenticated = useAuthStore.getState().isAuthenticated;

  if (isAuthenticated && saved && SUPPORTED_LANGUAGES.includes(saved as LanguageCode)) {
    // Authenticated users: respect saved choice
    initialLang = saved as LanguageCode;
  } else if (!isAuthenticated) {
    // === UNAUTHENTICATED USERS: always detect by geolocation ===
    fetch('https://api.ipbase.com/v1/json/')
      .then(res => res.json())
      .then(data => {
        const country = (data.country_code || '').toUpperCase();
        console.log('[Geo] Country detected via ipbase:', country);

        let detectedLang: LanguageCode = 'en';

        if (country === 'BY') {
          detectedLang = 'ru';
        } else if (['AT', 'CH', 'BE', 'LU'].includes(country)) {
          detectedLang = 'en';
        } else if (COUNTRY_TO_LANGUAGE[country]) {
          detectedLang = COUNTRY_TO_LANGUAGE[country];
        }

        if (SUPPORTED_LANGUAGES.includes(detectedLang)) {
          console.log('[Geo] Auto-setting language to:', detectedLang);
          localStorage.setItem('octra-language', detectedLang);
          document.documentElement.lang = detectedLang;
          set({ language: detectedLang });
        }
      })
      .catch((err) => {
        console.warn('[Geo] ipbase request failed:', err);
        const browserLang = navigator.language.split('-')[0];
        if (SUPPORTED_LANGUAGES.includes(browserLang as LanguageCode)) {
          set({ language: browserLang as LanguageCode });
        }
      });

    // Temporary initial value (will be overwritten by geo result)
    if (saved && SUPPORTED_LANGUAGES.includes(saved as LanguageCode)) {
      initialLang = saved as LanguageCode;
    } else {
      const browserLang = navigator.language.split('-')[0];
      if (SUPPORTED_LANGUAGES.includes(browserLang as LanguageCode)) {
        initialLang = browserLang as LanguageCode;
      }
    }
  } else {
    // Authenticated but no saved language
    const browserLang = navigator.language.split('-')[0];
    if (SUPPORTED_LANGUAGES.includes(browserLang as LanguageCode)) {
      initialLang = browserLang as LanguageCode;
    }
  }

  document.documentElement.lang = initialLang;

  return {
    language: initialLang,
    setLanguage: (lang) => {
      localStorage.setItem('octra-language', lang);
      document.documentElement.lang = lang;
      set({ language: lang });
    },
  };
});

// Translate helper removed - use useI18n hook instead

import { SUPPORTED_LANGUAGES, type LanguageCode } from '../config/languages';
import { useI18nStore } from '../stores/i18nStore';
import { useState, useEffect, useCallback } from 'react';

export { SUPPORTED_LANGUAGES };
export type { LanguageCode };

// Cache for loaded translations
const translationsCache: Record<string, Record<string, any>> = {};

// Static t function for non-React usage
export function t(key: string): string {
  const { language } = useI18nStore.getState();
  const translations = translationsCache[language] || {};
  
  if (Object.keys(translations).length === 0) return key;

  const keys = key.split('.');
  let current: any = translations;
  for (const k of keys) {
    if (current && typeof current === 'object' && k in current) {
      current = current[k];
    } else {
      return key;
    }
  }
  return typeof current === 'string' ? current : key;
}

// Expose cache for static t function
(window as any).__translationsCache = translationsCache;

// React hook for i18n
export function useI18n() {
  const language = useI18nStore((state) => state.language);
  const setLanguage = useI18nStore((state) => state.setLanguage);

  // Keep translations in local React state to ensure proper re-rendering
  const [translations, setTranslations] = useState<Record<string, any>>(
    () => translationsCache[language] || {}
  );
  const [loading, setLoading] = useState(Object.keys(translations).length === 0);

  const changeLanguage = useCallback((lang: LanguageCode) => {
    setLanguage(lang);
  }, [setLanguage]);

  // Load translations when language changes
  useEffect(() => {
    // Check cache first
    if (translationsCache[language]) {
      setTranslations(translationsCache[language]);
      setLoading(false);
      return;
    }

    // Load from file
    setLoading(true);
    fetch(`/languages/${language}.json`)
      .then((res) => {
        if (!res.ok) throw new Error(`Failed to load ${language}`);
        return res.json();
      })
      .then((data) => {
        translationsCache[language] = data;
        setTranslations(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error(`Failed to load ${language}:`, err);
        setLoading(false);
      });
  }, [language]);

  // Reactive t function
  const t = useCallback(
    (key: string): string => {
      if (Object.keys(translations).length === 0) return key;

      const keys = key.split('.');
      let current: any = translations;
      for (const k of keys) {
        if (current && typeof current === 'object' && k in current) {
          current = current[k];
        } else {
          return key;
        }
      }
      return typeof current === 'string' ? current : key;
    },
    [translations]
  );

  return {
    language,
    loading,
    changeLanguage,
    t,
  };
}

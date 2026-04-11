import { SUPPORTED_LANGUAGES, type LanguageCode } from '../config/languages';
import { useI18nStore, t as translate } from '../stores/i18nStore';
import { useState, useEffect, useCallback } from 'react';

export { SUPPORTED_LANGUAGES, translate as t };
export type { LanguageCode };

// React hook for i18n
export function useI18n() {
  const language = useI18nStore((state) => state.language);
  const loaded = useI18nStore((state) => state.loaded);
  const setLanguage = useI18nStore((state) => state.setLanguage);
  const loadTranslations = useI18nStore((state) => state.loadTranslations);
  const translations = useI18nStore((state) => state.translations);

  const changeLanguage = useCallback((lang: LanguageCode) => {
    setLanguage(lang);
  }, [setLanguage]);

  // Load initial language
  useEffect(() => {
    if (!translations[language]) {
      fetch(`/languages/${language}.json`)
        .then((res) => res.json())
        .then((data) => {
          loadTranslations(language, data);
        })
        .catch((err) => console.error(`Failed to load ${language}:`, err));
    }
  }, []);

  return {
    language,
    loading: !loaded && !translations[language],
    changeLanguage,
    t: translate,
  };
}

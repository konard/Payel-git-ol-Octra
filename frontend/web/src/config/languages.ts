// Supported languages - defined here to avoid circular imports
export const SUPPORTED_LANGUAGES = [
  'en', 'ru', 'hy', 'kk', 'uz', // Existing
  'fr', 'de', 'es', 'zh', 'ja', 'ko', 'it', // Western & Asian
  'uk', 'be', 'pl', 'pt', 'ar', 'tr', 'nl', // European
  'fi', 'sv', 'da', 'no', 'el', 'hu', 'cs', 'sk', 'sl', 'hr', 'sr', 'bg', 'ro', // More European
] as const;

export type LanguageCode = (typeof SUPPORTED_LANGUAGES)[number];

// Language metadata for display
export interface LanguageInfo {
  code: LanguageCode;
  name: string;
  nativeName: string;
  flag: string;
}

export const LANGUAGES_INFO: Record<LanguageCode, LanguageInfo> = {
  // Existing
  en: { code: 'en', name: 'English', nativeName: 'English', flag: 'GB' },
  ru: { code: 'ru', name: 'Russian', nativeName: 'Русский', flag: 'RU' },
  hy: { code: 'hy', name: 'Armenian', nativeName: 'Հայերեն', flag: 'AM' },
  kk: { code: 'kk', name: 'Kazakh', nativeName: 'Қазақша', flag: 'KZ' },
  uz: { code: 'uz', name: 'Uzbek', nativeName: "O'zbekcha", flag: 'UZ' },
  // Western & Asian
  fr: { code: 'fr', name: 'French', nativeName: 'Français', flag: 'FR' },
  de: { code: 'de', name: 'German', nativeName: 'Deutsch', flag: 'DE' },
  es: { code: 'es', name: 'Spanish', nativeName: 'Español', flag: 'ES' },
  zh: { code: 'zh', name: 'Chinese', nativeName: '中文', flag: 'CN' },
  ja: { code: 'ja', name: 'Japanese', nativeName: '日本語', flag: 'JP' },
  ko: { code: 'ko', name: 'Korean', nativeName: '한국어', flag: 'KR' },
  it: { code: 'it', name: 'Italian', nativeName: 'Italiano', flag: 'IT' },
  // European
  uk: { code: 'uk', name: 'Ukrainian', nativeName: 'Українська', flag: 'UA' },
  be: { code: 'be', name: 'Belarusian', nativeName: 'Беларуская', flag: 'BY' },
  pl: { code: 'pl', name: 'Polish', nativeName: 'Polski', flag: 'PL' },
  pt: { code: 'pt', name: 'Portuguese', nativeName: 'Português', flag: 'PT' },
  ar: { code: 'ar', name: 'Arabic', nativeName: 'العربية', flag: 'SA' },
  tr: { code: 'tr', name: 'Turkish', nativeName: 'Türkçe', flag: 'TR' },
  nl: { code: 'nl', name: 'Dutch', nativeName: 'Nederlands', flag: 'NL' },
  // More European
  fi: { code: 'fi', name: 'Finnish', nativeName: 'Suomi', flag: 'FI' },
  sv: { code: 'sv', name: 'Swedish', nativeName: 'Svenska', flag: 'SE' },
  da: { code: 'da', name: 'Danish', nativeName: 'Dansk', flag: 'DK' },
  no: { code: 'no', name: 'Norwegian', nativeName: 'Norsk', flag: 'NO' },
  el: { code: 'el', name: 'Greek', nativeName: 'Ελληνικά', flag: 'GR' },
  hu: { code: 'hu', name: 'Hungarian', nativeName: 'Magyar', flag: 'HU' },
  cs: { code: 'cs', name: 'Czech', nativeName: 'Čeština', flag: 'CZ' },
  sk: { code: 'sk', name: 'Slovak', nativeName: 'Slovenčina', flag: 'SK' },
  sl: { code: 'sl', name: 'Slovenian', nativeName: 'Slovenščina', flag: 'SI' },
  hr: { code: 'hr', name: 'Croatian', nativeName: 'Hrvatski', flag: 'HR' },
  sr: { code: 'sr', name: 'Serbian', nativeName: 'Српски', flag: 'RS' },
  bg: { code: 'bg', name: 'Bulgarian', nativeName: 'Български', flag: 'BG' },
  ro: { code: 'ro', name: 'Romanian', nativeName: 'Română', flag: 'RO' },
};

// Country code to language code mapping for auto-detection
export const COUNTRY_TO_LANGUAGE: Record<string, LanguageCode> = {
  RU: 'ru', UA: 'uk', BY: 'be', KZ: 'kk', UZ: 'uz', AM: 'hy',
  FR: 'fr', DE: 'de', ES: 'es', CN: 'zh', JP: 'ja', KR: 'ko', IT: 'it',
  PL: 'pl', PT: 'pt', SA: 'ar', TR: 'tr', NL: 'nl',
  FI: 'fi', SE: 'sv', DK: 'da', NO: 'no', GR: 'el', HU: 'hu',
  CZ: 'cs', SK: 'sk', SI: 'sl', HR: 'hr', RS: 'sr', BG: 'bg', RO: 'ro',
  US: 'en', GB: 'en', CA: 'en', AU: 'en',
};

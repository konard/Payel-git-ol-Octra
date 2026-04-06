// Supported languages - defined here to avoid circular imports
export const SUPPORTED_LANGUAGES = ['en', 'ru', 'hy', 'kk', 'uz'] as const;
export type LanguageCode = (typeof SUPPORTED_LANGUAGES)[number];

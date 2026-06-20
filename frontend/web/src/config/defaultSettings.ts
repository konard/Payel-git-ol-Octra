export const DEFAULT_PROVIDER = 'openrouter';
export const DEFAULT_MODEL = 'qwen/qwen3-coder';
export const DEFAULT_HIDE_API_KEY_INPUT = true;
export const DEFAULT_TOKEN = import.meta.env.VITE_OPENROUTER_API_KEY || '';
export const DEFAULT_SEARCH_PROVIDER_ID = 'apodex';

export interface DefaultSearchProviderConfig {
  id: string;
  provider: 'apodex' | 'custom';
  name: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  streaming: boolean;
}

export const DEFAULT_SEARCH_PROVIDER_CONFIG: DefaultSearchProviderConfig = {
  id: DEFAULT_SEARCH_PROVIDER_ID,
  provider: 'apodex',
  name: 'Apodex',
  baseUrl: 'https://api.apodex.ai/v1/responses',
  apiKey: '',
  model: 'apodex-1-0-deepresearch-mini',
  streaming: true,
};

export function resolveDefaultToken(token?: string | null): string {
  const trimmedToken = token?.trim();
  return trimmedToken || DEFAULT_TOKEN;
}

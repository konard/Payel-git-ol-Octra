export const DEFAULT_PROVIDER = 'openrouter';
export const DEFAULT_MODEL = 'qwen/qwen3-coder';
export const DEFAULT_HIDE_API_KEY_INPUT = true;
export const DEFAULT_TOKEN = import.meta.env.VITE_OPENROUTER_API_KEY || '';

export function resolveDefaultToken(token?: string | null): string {
  const trimmedToken = token?.trim();
  return trimmedToken || DEFAULT_TOKEN;
}

import { resolveDefaultToken } from '../config/defaultSettings';

interface CustomProviderAuth {
  base_url: string;
  api_key?: string | null;
}

interface BuildTaskProviderAuthInput {
  provider: string;
  apiKey?: string | null;
  defaultToken?: string | null;
  customProvider?: CustomProviderAuth | null;
}

interface TaskProviderAuth {
  provider: string;
  tokens: Record<string, string>;
}

function firstConfiguredToken(...tokens: Array<string | null | undefined>): string {
  for (const token of tokens) {
    const trimmedToken = token?.trim();
    if (trimmedToken) {
      return trimmedToken;
    }
  }

  return resolveDefaultToken();
}

export function buildTaskProviderAuth({
  provider,
  apiKey,
  defaultToken,
  customProvider,
}: BuildTaskProviderAuthInput): TaskProviderAuth {
  if (customProvider) {
    return {
      provider: 'ollama',
      tokens: {
        ollama: firstConfiguredToken(apiKey, customProvider.api_key, defaultToken),
        base_url: customProvider.base_url,
      },
    };
  }

  return {
    provider,
    tokens: {
      [provider]: firstConfiguredToken(apiKey, defaultToken),
    },
  };
}

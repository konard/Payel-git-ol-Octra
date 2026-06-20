import assert from 'node:assert/strict';
import { createServer } from 'vite';

// Issue #94: Apodex is the default configurable search provider. When the user
// supplies its credentials, task payloads must include the gateway-facing
// `search` object with the historical "striming" key, and custom search
// providers must carry their own model.

const server = await createServer({
  root: process.cwd(),
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: { middlewareMode: true },
  optimizeDeps: { noDiscovery: true, include: [] },
  appType: 'custom',
});

try {
  const defaults = await server.ssrLoadModule('/src/config/defaultSettings.ts');
  const settings = await server.ssrLoadModule('/src/stores/settingsStore.ts');
  const payload = await server.ssrLoadModule('/src/app/taskPayload.ts');

  assert.equal(defaults.DEFAULT_SEARCH_PROVIDER_ID, 'apodex');
  assert.equal(defaults.DEFAULT_SEARCH_PROVIDER_CONFIG.provider, 'apodex');
  assert.equal(defaults.DEFAULT_SEARCH_PROVIDER_CONFIG.model, 'apodex-1-0-deepresearch-mini');
  assert.equal(defaults.DEFAULT_SEARCH_PROVIDER_CONFIG.streaming, true);
  assert.match(defaults.DEFAULT_SEARCH_PROVIDER_CONFIG.baseUrl, /\/v1\/responses$/);

  const state = settings.useSettingsStore.getState();
  assert.equal(state.searchProviderId, 'apodex', 'Apodex must be the default search provider');
  assert.ok(state.searchProviders.some((provider) => provider.id === 'apodex'), 'default provider list must include Apodex');
  assert.equal(typeof state.updateSearchProvider, 'function');
  assert.equal(typeof state.addSearchProvider, 'function');

  const configuredApodex = payload.buildSearchConfigPayload({
    searchProviderId: 'apodex',
    searchProviders: [{
      ...defaults.DEFAULT_SEARCH_PROVIDER_CONFIG,
      apiKey: 'sk-apodex-test',
    }],
  });

  assert.deepEqual(configuredApodex, {
    provider: 'apodex',
    model: 'apodex-1-0-deepresearch-mini',
    'base-url': defaults.DEFAULT_SEARCH_PROVIDER_CONFIG.baseUrl,
    'api-key': 'sk-apodex-test',
    striming: true,
  });

  assert.equal(
    payload.buildSearchConfigPayload({
      searchProviderId: 'apodex',
      searchProviders: [defaults.DEFAULT_SEARCH_PROVIDER_CONFIG],
    }),
    null,
    'empty Apodex credentials must not be sent to the backend',
  );

  const customSearch = payload.buildSearchConfigPayload({
    searchProviderId: 'search-custom-1',
    searchProviders: [{
      id: 'search-custom-1',
      provider: 'custom',
      name: 'Internal Search Model',
      baseUrl: 'https://search.example.com/v1/chat/completions',
      apiKey: 'custom-key',
      model: 'search-model',
      streaming: false,
    }],
  });

  assert.deepEqual(customSearch, {
    provider: 'custom',
    model: 'search-model',
    'base-url': 'https://search.example.com/v1/chat/completions',
    'api-key': 'custom-key',
    striming: false,
  });

  console.log('check-search-provider-config: all assertions passed');
} finally {
  await server.close();
}

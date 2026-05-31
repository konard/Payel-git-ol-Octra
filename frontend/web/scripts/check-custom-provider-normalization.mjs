import assert from 'node:assert/strict';
import { createServer } from 'vite';

const storage = new Map();
globalThis.localStorage = {
  get length() {
    return storage.size;
  },
  clear() {
    storage.clear();
  },
  getItem(key) {
    return storage.get(key) ?? null;
  },
  key(index) {
    return Array.from(storage.keys())[index] ?? null;
  },
  removeItem(key) {
    storage.delete(key);
  },
  setItem(key, value) {
    storage.set(key, String(value));
  },
};

const originalWarn = console.warn;
console.warn = (...args) => {
  if (String(args[0]).includes('[zustand persist middleware] Unable to update item')) {
    return;
  }
  originalWarn(...args);
};

const originalLog = console.log;
console.log = (...args) => {
  if (String(args[0]).startsWith('Auth tokens - access:')) {
    return;
  }
  originalLog(...args);
};

const apiProvider = {
  ID: 'server-provider-1',
  UserID: 'user-1',
  Name: 'Local Ollama',
  BaseURL: 'http://localhost:11434/v1',
  APIKey: '',
  RequiresApiKey: false,
  CreatedAt: '2026-05-31T00:00:00Z',
  UpdatedAt: '2026-05-31T00:00:00Z',
};

const apiModel = {
  ID: 'server-model-1',
  UserID: 'user-1',
  Name: 'qwen3',
  ProviderID: 'server-provider-1',
  CreatedAt: '2026-05-31T00:00:00Z',
  UpdatedAt: '2026-05-31T00:00:00Z',
};

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
  const customProviders = await server.ssrLoadModule('/src/utils/customProviders.ts');
  const normalizedProvider = customProviders.normalizeCustomProvider({
    ...apiProvider,
    id: 'random-local-id-from-old-store',
    createdAt: '2026-05-31T01:00:00Z',
  });

  assert.deepEqual(normalizedProvider, {
    id: 'server-provider-1',
    user_id: 'user-1',
    name: 'Local Ollama',
    base_url: 'http://localhost:11434/v1',
    api_key: '',
    requires_api_key: false,
    created_at: '2026-05-31T00:00:00Z',
    updated_at: '2026-05-31T00:00:00Z',
  });

  assert.deepEqual(customProviders.normalizeCustomModel(apiModel), {
    id: 'server-model-1',
    user_id: 'user-1',
    name: 'qwen3',
    provider_id: 'server-provider-1',
    created_at: '2026-05-31T00:00:00Z',
    updated_at: '2026-05-31T00:00:00Z',
  });

  assert.equal(
    customProviders.normalizeCustomProviderList([
      { ...apiProvider, id: 'first-random-local-id' },
      { ...apiProvider, id: 'second-random-local-id' },
      { id: 'blank-row', name: '', base_url: '' },
    ]).length,
    1,
    'stale duplicate API providers should collapse to one rendered provider',
  );

  const { customProviderService } = await server.ssrLoadModule('/src/services/customProviderService.ts');
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => ({
    ok: true,
    json: async () => ({ status: 'success', data: [apiProvider] }),
  });

  assert.deepEqual(await customProviderService.getUserCustomProviders(), [normalizedProvider]);
  globalThis.fetch = originalFetch;

  const { useCustomProvidersStore } = await server.ssrLoadModule('/src/stores/customProvidersStore.ts');
  useCustomProvidersStore.setState({ providers: [], models: [] });
  const store = useCustomProvidersStore.getState();

  store.addProvider(apiProvider);
  store.addProvider(apiProvider);
  store.addModel(apiModel);
  store.addModel(apiModel);

  const state = useCustomProvidersStore.getState();
  assert.equal(state.providers.length, 1, 'same API provider should upsert, not duplicate');
  assert.equal(state.providers[0].id, 'server-provider-1');
  assert.equal(state.providers[0].name, 'Local Ollama');
  assert.equal(state.providers[0].base_url, 'http://localhost:11434/v1');
  assert.equal(state.models.length, 1, 'same API model should upsert, not duplicate');
  assert.equal(state.models[0].provider_id, 'server-provider-1');

  console.log('check-custom-provider-normalization: all assertions passed');
} finally {
  await server.close();
}

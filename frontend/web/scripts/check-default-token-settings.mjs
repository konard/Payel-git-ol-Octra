import assert from 'node:assert/strict';
import { createServer } from 'vite';

process.env.VITE_OPENROUTER_API_KEY = 'test-default-openrouter-token';

const server = await createServer({
  root: process.cwd(),
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: {
    middlewareMode: true,
  },
  optimizeDeps: {
    noDiscovery: true,
    include: [],
  },
  appType: 'custom',
});

try {
  const defaults = await server.ssrLoadModule('/src/config/defaultSettings.ts');
  assert.equal(defaults.DEFAULT_HIDE_API_KEY_INPUT, true);
  assert.equal(defaults.DEFAULT_PROVIDER, 'openrouter');
  assert.equal(defaults.DEFAULT_MODEL, 'qwen/qwen3-coder');
  assert.equal(defaults.DEFAULT_TOKEN, 'test-default-openrouter-token');
  assert.equal(defaults.resolveDefaultToken(''), 'test-default-openrouter-token');
  assert.equal(defaults.resolveDefaultToken(' user-token '), 'user-token');

  const payload = await server.ssrLoadModule('/src/app/taskPayload.ts');
  assert.deepEqual(
    payload.buildTaskProviderAuth({ provider: 'openrouter' }),
    {
      provider: 'openrouter',
      tokens: { openrouter: 'test-default-openrouter-token' },
    },
  );
  assert.deepEqual(
    payload.buildTaskProviderAuth({ provider: 'deepseek', apiKey: ' deepseek-token ' }),
    {
      provider: 'deepseek',
      tokens: { deepseek: 'deepseek-token' },
    },
  );
  assert.deepEqual(
    payload.buildTaskProviderAuth({
      provider: 'local-provider',
      defaultToken: 'global-token',
      customProvider: {
        base_url: 'http://localhost:11434/v1',
        api_key: 'custom-provider-token',
      },
    }),
    {
      provider: 'ollama',
      tokens: {
        ollama: 'custom-provider-token',
        base_url: 'http://localhost:11434/v1',
      },
    },
  );
} finally {
  await server.close();
}

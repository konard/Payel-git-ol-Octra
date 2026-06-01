import assert from 'node:assert/strict';
import { createServer } from 'vite';

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
  const calls = [];
  globalThis.localStorage = {
    getItem: () => 'token',
  };
  globalThis.fetch = async (url, init) => {
    calls.push({ url, init });
    return {
      ok: true,
      async json() {
        return { data: [] };
      },
    };
  };

  const { getChatHistory } = await server.ssrLoadModule('/src/services/chatHistoryService.ts');
  await getChatHistory('user-1');

  assert.equal(calls.length, 1);
  assert.equal(calls[0].url, '/chat/history?user_id=user-1');
  assert.equal(calls[0].init.headers.Authorization, 'Bearer token');

  console.log('check-chat-history-service: all assertions passed');
} finally {
  await server.close();
  delete globalThis.fetch;
  delete globalThis.localStorage;
}

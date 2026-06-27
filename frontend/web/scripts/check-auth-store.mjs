// Exercises the registration / login flow in the auth store against a mocked
// auth API. Guards three regressions fixed for issue #109:
//   1. The store must load in a non-browser (SSR/test) environment where
//      localStorage and document are undefined — module-level access used to
//      crash the whole frontend test suite.
//   2. A successful register() must persist tokens, mark the user authenticated
//      and start the trial period.
//   3. A failed register() must surface the server error message and rethrow so
//      the UI can react.
//   4. Per-user data must be stored under user-scoped keys so two accounts on
//      the same browser never read each other's settings.
import assert from 'node:assert/strict';
import { createServer } from 'vite';

const storage = new Map();
globalThis.localStorage = {
  getItem: (key) => (storage.has(key) ? storage.get(key) : null),
  setItem: (key, value) => storage.set(key, String(value)),
  removeItem: (key) => storage.delete(key),
};
globalThis.document = { cookie: '' };
globalThis.window = {};

// Minimal mock of the auth API. Routes by URL.
globalThis.fetch = (url, options = {}) => {
  const ok = (data) => Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve({ status: 'ok', data }),
  });
  const fail = (status, error) => Promise.resolve({
    ok: false,
    status,
    json: () => Promise.resolve({ status: 'error', error }),
  });

  if (url.endsWith('/register')) {
    const body = JSON.parse(options.body);
    if (body.email === 'taken@example.com') {
      return fail(409, 'Пользователь с таким email или именем уже существует');
    }
    return ok({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
      user: { id: 'user-1', username: body.username, email: body.email },
    });
  }
  if (url.endsWith('/me')) {
    return ok({
      user_id: 'user-1',
      username: 'newuser',
      email: 'new@example.com',
      created_at: new Date().toISOString(),
      has_subscription: false,
    });
  }
  return fail(404, 'not found');
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
  // (1) Module loads without throwing in the SSR/test environment.
  const { useAuthStore } = await server.ssrLoadModule('/src/stores/authStore');
  const { scopedStorage } = await server.ssrLoadModule('/src/stores/storageScope');

  // (2) Successful registration.
  await useAuthStore.getState().register('newuser', 'new@example.com', 'secret123');
  const afterRegister = useAuthStore.getState();
  assert.equal(afterRegister.isAuthenticated, true, 'register must authenticate the user');
  assert.equal(afterRegister.accessToken, 'access-123', 'register must store the access token');
  assert.equal(afterRegister.user?.email, 'new@example.com', 'register must store the user');
  assert.equal(afterRegister.isInTrial, true, 'new users must start in the trial period');
  assert.equal(localStorage.getItem('access_token'), 'access-123', 'tokens must be persisted');
  assert.equal(afterRegister.error, null, 'successful register must not set an error');

  // (4) Per-user scoped storage keys.
  scopedStorage.setItem('crewai-settings', '{"theme":"dark"}');
  assert.equal(
    localStorage.getItem('crewai-settings_user-1'),
    '{"theme":"dark"}',
    'settings must be stored under a user-scoped key',
  );
  assert.equal(
    localStorage.getItem('crewai-settings'),
    null,
    'settings must not be stored under the unscoped key while authenticated',
  );

  // (3) Failed registration surfaces the error and rethrows.
  await assert.rejects(
    () => useAuthStore.getState().register('dupe', 'taken@example.com', 'secret123'),
    /уже существует/,
    'duplicate registration must reject with the server error',
  );
  assert.match(
    useAuthStore.getState().error || '',
    /уже существует/,
    'failed register must surface the error message in the store',
  );

  console.log('check-auth-store: all assertions passed');
} finally {
  await server.close();
}

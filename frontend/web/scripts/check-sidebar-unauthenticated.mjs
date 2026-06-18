import assert from 'node:assert/strict';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { createServer } from 'vite';

const storage = new Map([['octra-language', 'ru']]);

globalThis.localStorage = {
  getItem: (key) => storage.get(key) ?? null,
  setItem: (key, value) => storage.set(key, value),
  removeItem: (key) => storage.delete(key),
};

globalThis.document = {
  documentElement: { lang: '' },
  addEventListener: () => {},
  removeEventListener: () => {},
};

Object.defineProperty(globalThis, 'navigator', {
  value: { language: 'ru-RU' },
  configurable: true,
});

globalThis.fetch = () => new Promise(() => {});
globalThis.window = {};

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
  const { useAuthStore } = await server.ssrLoadModule('/src/stores/authStore');
  const { useI18nStore } = await server.ssrLoadModule('/src/stores/i18nStore.ts');
  const { translationsCache } = await server.ssrLoadModule('/src/hooks/useI18n.ts');

  translationsCache.ru = {
    common: { close: 'Закрыть' },
    chatSidebar: {
      history: 'История',
      newChat: 'Новый чат',
      noChats: 'Чатов пока нет',
      notLoggedIn: 'Вы не вошли в аккаунт',
    },
  };

  useI18nStore.setState({ language: 'ru' });

  const { Sidebar } = await server.ssrLoadModule('/src/components/workspace/Sidebar.tsx');
  useAuthStore.setState({
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
  });
  const html = renderToStaticMarkup(React.createElement(Sidebar, {
    isOpen: true,
    onClose: () => {},
    onSelectChat: () => {},
    onNewChat: () => {},
  }));

  assert.match(
    html,
    /Вы не вошли в аккаунт/,
    'unauthenticated history sidebar must show a direct login message',
  );
  assert.doesNotMatch(
    html,
    /animate-spin/,
    'unauthenticated history sidebar must not show an endless loading spinner',
  );

  console.log('check-sidebar-unauthenticated: all assertions passed');
} finally {
  await server.close();
}

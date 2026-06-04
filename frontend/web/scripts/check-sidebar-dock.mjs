import assert from 'node:assert/strict';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { createServer } from 'vite';

// Regression guard for the docked Sessions pane introduced for the desktop
// workspace (issue #40). The same Sidebar component must render as a slide-in
// overlay on narrow screens and as an inline, always-open dock on wide screens.

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
  await server.ssrLoadModule('/src/hooks/useI18n.ts');

  window.__translationsCache.ru = {
    common: { close: 'Закрыть' },
    chatSidebar: {
      history: 'История',
      newChat: 'Новый чат',
      noChats: 'Чатов пока нет',
      notLoggedIn: 'Вы не вошли в аккаунт',
    },
  };

  useI18nStore.setState({ language: 'ru' });

  const { Sidebar } = await server.ssrLoadModule('/src/components/Sidebar.tsx');
  useAuthStore.setState({
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
  });

  const overlayHtml = renderToStaticMarkup(React.createElement(Sidebar, {
    isOpen: true,
    variant: 'overlay',
    onClose: () => {},
    onSelectChat: () => {},
    onNewChat: () => {},
  }));

  const dockHtml = renderToStaticMarkup(React.createElement(Sidebar, {
    isOpen: false,
    variant: 'dock',
    onClose: () => {},
    onSelectChat: () => {},
    onNewChat: () => {},
  }));

  // The overlay keeps its slide-in drawer chrome.
  assert.match(
    overlayHtml,
    /fixed inset-y-0 left-0/,
    'overlay sidebar must remain a fixed slide-in drawer',
  );

  // The dock is rendered inline and must not use the fixed/translate drawer
  // classes, otherwise it would float over the workspace instead of sitting in
  // its resizable pane.
  assert.doesNotMatch(
    dockHtml,
    /fixed inset-y-0 left-0/,
    'docked sidebar must not render as a fixed drawer',
  );
  assert.doesNotMatch(
    dockHtml,
    /-translate-x-full/,
    'docked sidebar must not translate off-screen',
  );
  assert.match(
    dockHtml,
    /relative h-full w-full/,
    'docked sidebar must fill its pane inline',
  );

  // The dock is always visible, so it must not offer a close (X) button.
  assert.doesNotMatch(
    dockHtml,
    /Закрыть/,
    'docked sidebar must not show a close button',
  );

  // The session list must still load while docked even when isOpen is false.
  assert.match(
    dockHtml,
    /Вы не вошли в аккаунт/,
    'docked sidebar must render its content regardless of the isOpen flag',
  );

  // The dock reads as a recessed panel using the slightly darker "sunken"
  // surface (issue #40: "Сделай серый чуть потемнее").
  assert.match(
    dockHtml,
    /surface-sunken/,
    'docked sidebar must use the darker --surface-sunken background',
  );
  assert.doesNotMatch(
    overlayHtml,
    /surface-sunken/,
    'overlay sidebar keeps the regular --surface background',
  );

  // The refreshed history header offers a compact "New" action with a Ctrl+N
  // hint in the dock; the overlay keeps its full-width button instead.
  assert.match(
    dockHtml,
    /Ctrl\+N/,
    'docked sidebar header must advertise the Ctrl+N new-chat shortcut',
  );
  assert.doesNotMatch(
    overlayHtml,
    /Ctrl\+N/,
    'overlay sidebar must not show the compact Ctrl+N header button',
  );

  console.log('check-sidebar-dock: all assertions passed');
} finally {
  await server.close();
}

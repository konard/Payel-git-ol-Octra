/*
 * Renderer-side access to the desktop backend.
 *
 * Inside Electron, `window.octra` is injected by preload.js. When the same HTML
 * is opened in a plain browser (used for visual previews and the Playwright
 * screenshot test), `window.octra` is absent, so we fall back to an in-memory
 * mock that keeps the chrome fully interactive without touching the filesystem.
 */
(function (root) {
  const real = root.octra;

  // In-memory mock mirroring the main-process project store, seeded to match the
  // reference 4 screenshot (octra / LeFine + recent projects).
  function createMock() {
    let recent = [
      { name: 'octra', path: '/home/user/dev/octra', openedAt: 4 },
      { name: 'LeFine', path: '/home/user/dev/lefine', openedAt: 3 },
      { name: 'Вакансии столяр (помощь отцу)', path: '/home/user/dev/vakansii-stolyar', openedAt: 2 },
      { name: 'TradeFast', path: '/home/user/dev/tradefast', openedAt: 1 },
    ];
    const baseName = (p) => p.split(/[\\/]/).filter(Boolean).pop() || p;
    const remember = (path, name) => {
      recent = recent.filter((r) => r.path !== path);
      const entry = { name: name || baseName(path), path, openedAt: Date.now ? 0 : 0 };
      recent = [entry, ...recent].slice(0, 12);
      return { project: entry, recent };
    };
    return {
      isElectron: false,
      platform: 'browser',
      window: {
        minimize() {},
        toggleMaximize() {},
        close() {},
        isMaximized: async () => false,
      },
      projects: {
        listRecent: async () => recent,
        open: async () => remember('/home/user/dev/new-folder', null),
        openPath: async (p) => remember(p, null),
        create: async (name) => remember('/home/user/dev/' + name, name),
        forget: async (p) => {
          recent = recent.filter((r) => r.path !== p);
          return recent;
        },
      },
      shell: { openExternal() {} },
      onMenuCommand() {},
    };
  }

  root.OctraApi = real && real.isElectron ? real : createMock();
})(window);
